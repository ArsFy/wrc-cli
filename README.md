# WRs Cli

A Web Server in terminal.

## What is WRs Cli?

### Reverse Proxy Mode

- `-a-header "KEY1:VALUE1;KEY2:VALUE2"` Add headers to return to client
- `-r-header "KEY1:VALUE1;KEY2:VALUE2"` Add headers to send to target
- `-api "http://127.0.0.1:8000/api"` Set API target
- `-port 8000` Set port to listen

```bash
./wrs-cli [options] http://127.0.0.1:8000
```

#### File Server Mode

- `-port 8000` Set port to listen
- `-token "TOKEN"` Set token to access files
  - This token use GET query value: `?token=TOKEN`

```bash
./wrs-cli [options] /path/to/file
```

## How to install

### Download

Open release page and download the latest version.

### Build

```bash
git clone https://github.com/ArsFy/wrs-cli.git
cd wrs-cli
go build
```
