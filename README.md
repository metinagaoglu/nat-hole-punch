# NAT Hole Punching Server & Client

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://golang.org/)
[![CI](https://github.com/yourusername/nat-hole-punch/actions/workflows/ci.yml/badge.svg)](https://github.com/yourusername/nat-hole-punch/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

A UDP hole punching implementation in Go for establishing peer-to-peer connections between devices behind NATs. Includes an interactive client with text messaging and file transfer.

## What is NAT Hole Punching?

NAT hole punching allows two devices behind different NATs to establish a direct connection:

1. Both clients register with a public signaling server
2. The server shares each client's public IP and port
3. Clients send packets to each other's public addresses
4. NAT devices see outgoing packets and allow return traffic

```
┌─────────────┐         ┌─────────────┐
│  Client A   │         │  Client B   │
│  (NAT 1)    │         │  (NAT 2)    │
└──────┬──────┘         └──────┬──────┘
       │                       │
       │    1. Register        │
       ├──────────┐   ┌────────┤
       │          ▼   ▼        │
       │    ┌──────────────┐   │
       │    │   Signal     │   │
       │    │   Server     │   │
       │    └──────────────┘   │
       │          │   │        │
       │    2. Exchange Info   │
       │◄─────────┘   └───────►│
       │                       │
       │    3. Direct P2P      │
       │◄─────────────────────►│
```

## Project Structure

```
.
├── server/                     # UDP Signaling Server
│   ├── cmd/main.go             # Entry point
│   └── pkg/
│       ├── config/             # Environment-based configuration
│       ├── handlers/           # Request handlers + HandlerContext (DI)
│       ├── logger/             # Structured logging (slog)
│       ├── models/             # Client, Request, HandlerFunc
│       ├── repositories/       # IRepository + adapters (memory, redis)
│       ├── router/             # Event routing
│       └── server/             # UDP server with graceful shutdown
├── client/                     # Interactive P2P Client
│   ├── main.go                 # Entry point + interactive CLI
│   ├── client.go               # Client lifecycle (register, listen, shutdown)
│   ├── protocol.go             # Message types (text, file, heartbeat)
│   ├── peers.go                # Peer tracking with dedup
│   └── transfer.go             # Chunked file transfer
├── docker-compose.yml          # Single network setup
├── docker-compose.simple-test.yml
└── docker-compose.separate-networks.yml
```

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose (optional)

### Build

```bash
# Server
cd server && go build -o udp-server ./cmd

# Client
cd client && go build -o udp-client .
```

### Run Locally

```bash
# Terminal 1: Start the server
cd server && go run ./cmd

# Terminal 2: Client A
cd client && go run . -room-key=my-room

# Terminal 3: Client B
cd client && go run . -room-key=my-room
```

Clients automatically get a free port from the OS. Once both register to the same room, they discover each other and can communicate.

## Client Usage

After connecting, the client provides an interactive CLI:

```
╔══════════════════════════════════════════╗
║        UDP Hole Punch Client             ║
╠══════════════════════════════════════════╣
║  <message>     Send text to all peers    ║
║  /send <file>  Send file to all peers    ║
║  /peers        List connected peers      ║
║  /help         Show this help            ║
║  /quit         Disconnect and exit       ║
╚══════════════════════════════════════════╝
```

### Text Messaging

```
hello world          # Sends "hello world" to all peers
```

Output on the receiving side:
```
[MSG] 192.168.1.5:52431: hello world
```

### File Transfer

```
/send photo.jpg      # Sends photo.jpg to all peers (chunked UDP)
```

Files are split into 512-byte chunks, base64-encoded, and reassembled on the receiving end. Received files are saved as `received_<filename>` in the working directory.

### Peer Management

```
/peers               # Shows connected peers with last-seen time
```

```
Connected peers (2):
  - 192.168.1.5:52431 (last seen: 3s ago)
  - 192.168.1.8:49012 (last seen: 1s ago)
```

### Client Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-signal-address` | Signal server address | `127.0.0.1:3986` |
| `-local-address` | Local bind address | `0.0.0.0:0` (auto) |
| `-room-key` | Room for peer discovery | `default` |

## Server Configuration

All configuration via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | UDP server port | `3986` |
| `SERVER_HOST` | Bind address | `0.0.0.0` |
| `BUFFER_SIZE` | UDP buffer size | `1024` |
| `CLIENT_TTL` | Registration TTL (seconds) | `60` |
| `REPOSITORY_TYPE` | Storage backend | `memory` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `REDIS_DB` | Redis database | `0` |
| `LOG_LEVEL` | Log level | `info` |
| `LOG_FORMAT` | Log format (`text`/`json`) | `text` |

### Structured Logging

```bash
# Development (human-readable)
LOG_LEVEL=debug go run ./cmd

# Production (JSON for log aggregation)
LOG_FORMAT=json LOG_LEVEL=info go run ./cmd
```

## Docker Deployment

### Quick Start

```bash
# Start server + two clients + redis
docker-compose up -d

# Watch the logs
docker-compose logs -f client1 client2

# Stop everything
docker-compose down
```

### Multi-Network Testing

```bash
# Clients on separate networks (simulates different NATs)
docker-compose -f docker-compose.simple-test.yml up -d
docker-compose -f docker-compose.simple-test.yml logs -f client1 client2
```

For advanced NAT simulation with iptables (Linux only), see [NETWORK_TESTING.md](NETWORK_TESTING.md).

## Protocol

### Signal Server Events

Client-server communication uses JSON events:

```json
{"event": "register", "payload": "{\"local_ip\":\"0.0.0.0:4000\",\"key\":\"my-room\"}"}
{"event": "logout",   "payload": "{\"local_ip\":\"0.0.0.0:4000\",\"key\":\"my-room\"}"}
```

The server responds with a comma-separated peer list:
```
192.168.1.10:4000,192.168.1.20:4001
```

### Peer-to-Peer Messages

Peers communicate using structured JSON:

```json
{"type": "heartbeat", "from": "192.168.1.10:4000"}
{"type": "text",      "from": "192.168.1.10:4000", "payload": "hello"}
{"type": "file",      "from": "192.168.1.10:4000", "payload": "{\"name\":\"photo.jpg\",\"size\":1024,\"total_chunks\":2}"}
{"type": "file_chunk","from": "192.168.1.10:4000", "payload": "{\"name\":\"photo.jpg\",\"index\":0,\"total\":2,\"data\":\"base64...\"}"}
{"type": "file_ack",  "from": "192.168.1.10:4000", "payload": "{\"name\":\"photo.jpg\",\"success\":true}"}
```

## Server Architecture

Key design decisions:

- **Dependency Injection**: `HandlerContext` carries repository + connection, no global state
- **Thread Safety**: `sync.RWMutex` on in-memory repository, mutex on server client map
- **TTL with Refresh**: Active clients extend their TTL on every message; idle clients expire and get cleaned up by a background goroutine
- **Input Validation**: Room keys validated (max 64 chars, alphanumeric + hyphen/underscore)
- **Graceful Shutdown**: Signal handling (SIGINT/SIGTERM), context cancellation, resource cleanup

## Development

```bash
cd server

# Run tests
go test -race ./...

# Run tests with coverage
go test -race -cover ./...

# Lint (requires golangci-lint)
golangci-lint run

# Static analysis
go vet ./...
```

## Security Notes

- The signaling server should be behind a firewall in production
- Consider authentication for room registration
- UDP traffic is unencrypted; use DTLS for sensitive data
- Room key validation prevents injection but rate limiting is not yet implemented

## License

MIT License - see [LICENSE](LICENSE) for details.
