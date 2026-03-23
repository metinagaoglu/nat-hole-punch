# holepunch

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

UDP hole punching library and CLI for Go. Establish peer-to-peer connections between devices behind NATs with text messaging and file transfer.

---

## Try It Out (clone & run)

Clone the repo and start experimenting in under a minute.

### 1. Text Chat Between Two Peers

```bash
git clone https://github.com/metnagaoglu/holepunch.git
cd holepunch

# Terminal 1 — start the signal server
go run ./cmd/server

# Terminal 2 — client A
go run ./cmd/client -room-key=chat-room

# Terminal 3 — client B
go run ./cmd/client -room-key=chat-room
```

Once both clients register, type a message in either terminal:

```
> hello from client A!
# Client B sees:
# [MSG] 127.0.0.1:52431: hello from client A!
```

### 2. File Transfer

In client A's terminal:

```
/send photo.jpg
```

Client B receives it chunk by chunk:

```
Receiving file 'photo.jpg' (48231 bytes, 95 chunks)
Received chunk 1/95 for 'photo.jpg'
...
File 'photo.jpg' saved as 'received_photo.jpg' (48231 bytes)
```

### 3. Docker (no Go required)

```bash
docker-compose up -d
docker-compose logs -f client1 client2

# Both clients auto-register to "test-room" and start exchanging heartbeats
```

### 4. Multi-Network NAT Simulation

```bash
# Clients on separate Docker networks (simulates different NATs)
docker-compose -f docker-compose.simple-test.yml up -d
docker-compose -f docker-compose.simple-test.yml logs -f client1 client2
```

For advanced iptables-based NAT simulation (Linux only), see [NETWORK_TESTING.md](NETWORK_TESTING.md).

### Interactive Commands

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

### Client Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-signal-address` | Signal server address | `127.0.0.1:3986` |
| `-local-address` | Local bind address | `0.0.0.0:0` (auto) |
| `-room-key` | Room for peer discovery | `default` |

---

## Use as a Library

```bash
go get github.com/metnagaoglu/holepunch
```

Import only what you need — each package is independent:

```go
import "github.com/metnagaoglu/holepunch/peer"     // P2P client only
import "github.com/metnagaoglu/holepunch/signal"   // Signal server only
import "github.com/metnagaoglu/holepunch/protocol" // Shared types only
```

### Example: Embed a Signal Server

```go
package main

import (
    "log"
    "os/signal"
    "syscall"

    "github.com/metnagaoglu/holepunch/signal"
)

func main() {
    cfg := signal.DefaultConfig()
    cfg.Port = 3986

    srv, err := signal.NewServer(cfg)
    if err != nil {
        log.Fatal(err)
    }

    go srv.ListenAndServe()

    // Wait for Ctrl+C
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT)
    <-sig

    srv.Shutdown()
}
```

### Example: P2P Chat App

```go
package main

import (
    "bufio"
    "fmt"
    "log"
    "os"

    "github.com/metnagaoglu/holepunch/peer"
)

func main() {
    client, err := peer.Connect("localhost:3986", "0.0.0.0:0", "my-room")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Print incoming messages
    client.OnMessage(func(from string, text string) {
        fmt.Printf("\n[%s]: %s\n> ", from, text)
    })

    go client.Listen()

    // Read from stdin and send
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Print("> ")
    for scanner.Scan() {
        client.Send(scanner.Text())
        fmt.Print("> ")
    }
}
```

### Example: File Sync Tool

```go
client, _ := peer.Connect("signal.example.com:3986", "0.0.0.0:0", "sync-room")
defer client.Close()

go client.Listen()

// Send a file to all peers in the room
client.SendFile("backup.tar.gz")
```

### Example: Game Lobby with Custom Signal Server

```go
// Server side — embed in your game server
cfg := signal.DefaultConfig()
cfg.Port = 7777
cfg.ClientTTL = 120  // 2 minute timeout for game sessions
cfg.RepositoryType = "redis"
cfg.RedisAddr = "redis:6379"

srv, _ := signal.NewServer(cfg)
go srv.ListenAndServe()

// Client side — connect from game client
client, _ := peer.Connect("game.example.com:7777", "0.0.0.0:0", "lobby-42")
go client.Listen()

client.OnMessage(func(from, text string) {
    // Handle game state updates from peers
    handleGameMessage(from, text)
})

// Send game state to peers
client.Send(`{"action":"move","x":10,"y":20}`)
```

---

## How It Works

```
┌─────────────┐         ┌─────────────┐
│  Client A   │         │  Client B   │
│  (NAT 1)    │         │  (NAT 2)    │
└──────┬──────┘         └──────┬──────┘
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

1. Both clients register with the signal server
2. Server shares each client's public IP:port with room members
3. Clients send packets to each other's public addresses
4. NAT devices see outgoing packets and allow return traffic

## Project Structure

```
github.com/metnagaoglu/holepunch/
├── protocol/               # Shared message types (signal + peer)
│   └── protocol.go
├── signal/                 # Signal server library (importable)
│   ├── server.go           # NewServer, ListenAndServe, Shutdown
│   ├── config.go           # Config, LoadConfigFromEnv
│   ├── logger.go           # InitLogger (slog)
│   └── internal/           # Not importable by external code
│       ├── handler/        # DI context, register/logout, validation
│       ├── repository/     # Repository interface
│       └── repository/adapters/  # memory, redis
├── peer/                   # P2P client library (importable)
│   ├── client.go           # Connect, Send, SendFile, Listen, OnMessage
│   ├── peers.go            # PeerManager (thread-safe tracking)
│   └── transfer.go         # Chunked file transfer
├── cmd/
│   ├── server/main.go      # CLI wrapper (~40 lines)
│   └── client/main.go      # CLI wrapper (~90 lines)
├── server/Dockerfile
├── client/Dockerfile
└── docker-compose.yml
```

## Server Configuration

All via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | UDP server port | `3986` |
| `SERVER_HOST` | Bind address | `0.0.0.0` |
| `BUFFER_SIZE` | UDP buffer size | `1024` |
| `CLIENT_TTL` | Registration TTL (seconds) | `60` |
| `REPOSITORY_TYPE` | Storage backend (`memory`/`redis`) | `memory` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `REDIS_DB` | Redis database | `0` |
| `LOG_LEVEL` | Log level (`debug`/`info`/`warn`/`error`) | `info` |
| `LOG_FORMAT` | Log format (`text`/`json`) | `text` |

## Protocol

### Signal Server Events

```json
{"event": "register", "payload": "{\"local_ip\":\"0.0.0.0:4000\",\"key\":\"my-room\"}"}
{"event": "logout",   "payload": "{\"local_ip\":\"0.0.0.0:4000\",\"key\":\"my-room\"}"}
```

Response: comma-separated peer list `192.168.1.10:4000,192.168.1.20:4001`

### Peer-to-Peer Messages

```json
{"type": "heartbeat", "from": "192.168.1.10:4000"}
{"type": "text",      "from": "192.168.1.10:4000", "payload": "hello"}
{"type": "file",      "from": "...", "payload": "{\"name\":\"photo.jpg\",\"size\":1024,\"total_chunks\":2}"}
{"type": "file_chunk","from": "...", "payload": "{\"name\":\"photo.jpg\",\"index\":0,\"total\":2,\"data\":\"base64...\"}"}
{"type": "file_ack",  "from": "...", "payload": "{\"name\":\"photo.jpg\",\"success\":true}"}
```

## Architecture

- **Dependency Injection**: Handler context carries repository + connection, no global state
- **Thread Safety**: `sync.RWMutex` on in-memory repository
- **TTL with Refresh**: Active clients extend TTL on every message; idle clients expire via background goroutine
- **Input Validation**: Room keys validated (max 64 chars, alphanumeric + hyphen/underscore)
- **Graceful Shutdown**: Signal handling, context cancellation, resource cleanup
- **`internal/` packages**: Implementation details hidden from library consumers

## Development

```bash
# Build
go build -o /dev/null ./cmd/server
go build -o /dev/null ./cmd/client

# Test
go test -race ./...

# Vet
go vet ./...
```

## License

MIT - see [LICENSE](LICENSE) for details.
