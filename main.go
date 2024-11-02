package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
)

func main() {
	api := flag.String("api", "", "API endpoint")
	aHeader := flag.String("a-header", "", "Active Server Header")
	rHeader := flag.String("r-header", "", "Reverse Proxy Header")
	port := flag.String("port", "8080", "Port to listen on")
	token := flag.String("token", "", "Query Token (?token=xxx)")

	flag.Usage = func() {
		fmt.Println("Usage: wrc-cli [options] <host/path>")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	// A Header
	aHeaderMap := func() map[string]string {
		data := make(map[string]string)
		headers := strings.Split(*aHeader, ";")
		for _, header := range headers {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				data[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
		return data
	}()

	// R Header
	rHeaderMap := func() map[string]string {
		data := make(map[string]string)
		headers := strings.Split(*rHeader, ";")
		for _, header := range headers {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				data[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
		return data
	}()

	// Host
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Use -h to get help")
		return
	}
	host := args[0]

	// Check if host is a URL or a file path
	var proxy *httputil.ReverseProxy
	var apiProxy *httputil.ReverseProxy
	var fileServer http.Handler

	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		// Target
		target, err := url.Parse(host)
		if err != nil {
			fmt.Println("Invalid host URL:", err)
			return
		}

		// Reverse Proxy
		proxy = httputil.NewSingleHostReverseProxy(target)

		// Modify the request to include the rHeader values
		proxy.ModifyResponse = func(resp *http.Response) error {
			if *rHeader != "" {
				headers := strings.Split(*rHeader, ";")
				for _, header := range headers {
					parts := strings.SplitN(header, ":", 2)
					if len(parts) == 2 {
						resp.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
					}
				}
			}
			for key, value := range aHeaderMap {
				resp.Header.Set(key, value)
			}
			return nil
		}

		// Modify the request to include the original request
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
			req.Host = target.Host
		}

		// API Proxy
		if *api != "" {
			apiURL, err := url.Parse(*api)
			if err != nil {
				fmt.Println("Invalid API URL:", err)
				return
			}
			apiProxy = httputil.NewSingleHostReverseProxy(apiURL)
			apiProxy.Director = func(req *http.Request) {
				req.URL.Scheme = apiURL.Scheme
				req.URL.Host = apiURL.Host
				req.URL.Path = singleJoiningSlash(apiURL.Path, strings.TrimPrefix(req.URL.Path, "/api"))
				req.Host = apiURL.Host
			}
		}
	} else {
		// File Server
		fileServer = http.FileServer(http.Dir(host))
	}

	// Start the server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] (%s) %s\n", r.Method, r.RemoteAddr, r.URL)
		if apiProxy != nil && strings.HasPrefix(r.URL.Path, "/api") {
			apiProxy.ServeHTTP(w, r)
		} else if proxy != nil {
			proxy.ServeHTTP(w, r)
		} else if fileServer != nil {
			if *token != "" {
				queryToken := r.URL.Query().Get("token")
				fmt.Println(queryToken)
				if queryToken != *token {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}
			http.ServeFile(w, r, filepath.Join(host, r.URL.Path))
		}
	})

	fmt.Println("Starting reverse proxy server on 0.0.0.0:" + *port)
	fmt.Println("Target:", host)
	if *token != "" {
		fmt.Println("Query Token:", *token)
	}
	if *api != "" {
		fmt.Println("API Endpoint:", *api)
	}
	if *aHeader != "" {
		fmt.Println("\nReturn to Client Headers:")
		for key, value := range aHeaderMap {
			fmt.Println(key, ":", value)
		}
	}
	if *rHeader != "" {
		fmt.Println("\nSend to Target Headers:")
		for key, value := range rHeaderMap {
			fmt.Println(key, ":", value)
		}
	}
	fmt.Println("\nLogs:")
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

// singleJoiningSlash ensures there is exactly one slash between the base path and the request path
func singleJoiningSlash(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	aSlash := a[len(a)-1] == '/'
	bSlash := b[0] == '/'
	switch {
	case aSlash && bSlash:
		return a + b[1:]
	case !aSlash && !bSlash:
		return a + "/" + b
	}
	return a + b
}
