# ShellFromBrowser

A modern, cross-platform web-based terminal emulator written in Go. Spiritual successor to ShellInBox with SSH client support, multi-sessions, file transfer, and session recording.

## Features

- **Browser-based terminal** — Full xterm.js terminal emulation (256 colors, Unicode, mouse)
- **Multi-session tabs** — Multiple terminal sessions in one browser window
- **SSH client** — Connect to remote hosts directly from the browser
- **Authentication** — JWT-based auth with configurable user backends
- **TLS/HTTPS** — Built-in TLS support
- **File transfer** — Upload/download files through the web interface
- **Session recording** — Record and replay sessions (asciicast v2 format)
- **Cross-platform** — Runs on Linux, macOS, Windows
- **Single binary** — Zero runtime dependencies, everything embedded
- **Docker ready** — Multi-stage Dockerfile included

## Quick Start

### Binary

```bash
# Download from releases or build from source
go install github.com/valorisa/ShellFromBrowser/cmd/shellfb@latest

# Run with defaults (no auth, port 8080)
shellfb

# Run with config
shellfb --config config.yaml
```

### Docker

```bash
docker compose up -d
```

Open http://localhost:8080

## Configuration

Copy `config.example.yaml` and customize:

```yaml
server:
  addr: ":8080"
  tls:
    enabled: true
    cert: "/path/to/cert.pem"
    key: "/path/to/key.pem"

auth:
  enabled: true
  jwt_secret: "your-secret-here"
  users:
    - username: admin
      password_hash: "$2a$10$..."

ssh:
  enabled: true

recording:
  enabled: true
  dir: "./recordings"
```

### Generate password hash

```bash
shellfb hash-password
```

## Build from source

```bash
git clone https://github.com/valorisa/ShellFromBrowser.git
cd ShellFromBrowser
make build
./bin/shellfb
```

## Security

- All WebSocket connections require authentication when auth is enabled
- JWT tokens with configurable expiry
- Rate limiting on login endpoint
- Security headers (CSP, X-Frame-Options, etc.)
- Path traversal protection on file operations
- No eval(), no inline scripts

## License

MIT
