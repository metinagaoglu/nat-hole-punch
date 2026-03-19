# NAT Hole Punching Server & Client

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-67.2%25-brightgreen.svg)](https://golang.org/)

A production-ready UDP hole punching implementation in Go for establishing peer-to-peer connections between devices behind different NATs (Network Address Translation).

## 🌟 Features

- **UDP Hole Punching**: Establish direct peer-to-peer connections through NAT
- **Room-Based Management**: Organize clients into rooms with unique keys
- **Configuration Management**: Environment variable support with sensible defaults
- **Dual Storage Backend**: In-memory (default) or Redis for persistence and scalability
- **Production Ready**: Proper error handling, logging, and test coverage (67.2%)
- **Docker Support**: Containerized deployment with Docker Compose
- **Network Testing**: Separate network configurations for realistic NAT simulation
- **Improved Client**: Context-based lifecycle management with graceful shutdown
- **Type-Safe**: Full Go type safety with clean architecture

## 📋 Table of Contents

- [What is NAT Hole Punching?](#what-is-nat-hole-punching)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Server](#server)
  - [Client](#client)
- [Docker Deployment](#docker-deployment)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## 🔍 What is NAT Hole Punching?

NAT (Network Address Translation) is commonly used in routers to map multiple private IP addresses to a single public IP address. This creates challenges for peer-to-peer communication.

NAT hole punching is a technique that allows two devices behind different NATs to establish a direct connection by:

1. Both clients register with a public signaling server
2. The server shares each client's public IP and port
3. Clients simultaneously send packets to each other's public addresses
4. NAT devices see outgoing packets and allow incoming packets from the same address

### Useful Resources

- [Wikipedia: Hole Punching](https://en.wikipedia.org/wiki/Hole_punching_(networking))
- [P2P NAT Traversal Guide](https://itnext.io/p2p-nat-traversal-how-to-punch-a-hole-9abc8ffa758e)
- [Hole Punching in Networking](https://medium.com/@surapuramakhil/hole-punching-nat-in-networking-72502c8d1b7c)

## 🏗️ Architecture

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
       │    │ Signaling    │   │
       │    │ Server       │   │
       │    └──────────────┘   │
       │          │   │        │
       │    2. Exchange Info   │
       │◄─────────┘   └────────┤
       │                       │
       │    3. Punch Hole      │
       ├───────────────────────►
       │    4. Direct P2P      │
       │◄──────────────────────►
```

### Project Structure

```
.
├── server/                          # UDP Signaling Server
│   ├── cmd/
│   │   └── main.go                 # Server entry point
│   ├── pkg/
│   │   ├── config/                 # Configuration management
│   │   ├── handlers/               # Request handlers (register, logout)
│   │   ├── models/                 # Data models (Client, Request)
│   │   ├── repositories/           # Data persistence layer
│   │   ├── router/                 # Event routing
│   │   └── server/                 # UDP server implementation
│   ├── go.mod
│   └── .env.example                # Configuration template
├── client/
│   └── client.go                   # Improved UDP Client with context
├── docker-compose.yml              # Single network setup
├── docker-compose.separate-networks.yml  # Multi-network NAT simulation
├── docker-compose.services.yml     # Minimal services setup
├── NETWORK_TESTING.md              # Network testing guide
└── LICENSE                         # MIT License

```

## 🚀 Installation

### Prerequisites

- Go 1.23 or higher
- Docker & Docker Compose (optional, for containerized deployment)

### Build from Source

#### Server

```bash
cd server
go mod download
go build -o udp-server ./cmd/main.go
```

#### Client

```bash
cd client
go build -o udp-client client.go
```

### Build for Different Platforms

#### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -o client-osx-amd64 client.go
```

#### macOS (Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -o client-osx-arm64 client.go
```

#### Linux
```bash
GOOS=linux GOARCH=amd64 go build -o client-linux client.go
```

#### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o client-windows.exe client.go
```

## ⚙️ Configuration

The server supports configuration via environment variables. Copy `.env.example` to `.env` and customize:

### Environment Variables

| Variable | Description | Default | Valid Range |
|----------|-------------|---------|-------------|
| `SERVER_PORT` | UDP server port | `3986` | 1-65535 |
| `SERVER_HOST` | Server bind address | `0.0.0.0` | Valid IP/hostname |
| `BUFFER_SIZE` | UDP buffer size (bytes) | `1024` | 512-65536 |
| `CLIENT_TTL` | Client registration TTL (seconds) | `60` | 10-3600 |
| `REPOSITORY_TYPE` | Storage backend | `memory` | `memory`, `redis` |
| `REDIS_ADDR` | Redis server address | `localhost:6379` | Host:port format |
| `REDIS_PASSWORD` | Redis authentication | `` | Optional password |
| `REDIS_DB` | Redis database number | `0` | 0-15 |

### Example Configurations

#### In-Memory Storage (Default)
```bash
# Server Configuration
SERVER_PORT=3986
SERVER_HOST=0.0.0.0
BUFFER_SIZE=1024
CLIENT_TTL=60

# Repository
REPOSITORY_TYPE=memory
```

#### Redis Storage (Production)
```bash
# Server Configuration
SERVER_PORT=3986
SERVER_HOST=0.0.0.0
BUFFER_SIZE=1024
CLIENT_TTL=60

# Repository - Redis Backend
REPOSITORY_TYPE=redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your_password
REDIS_DB=0
```

For detailed Redis setup and usage, see [REDIS.md](REDIS.md).

## 📖 Usage

### Server

#### Start the Server

```bash
# Using binary
cd server
./udp-server

# Using go run
go run cmd/main.go

# With custom configuration
SERVER_PORT=8080 BUFFER_SIZE=2048 go run cmd/main.go
```

Expected output:
```
2024/11/20 18:00:00 Starting UDP Hole Punch Server on 0.0.0.0:3986
2024/11/20 18:00:00 Adding route for [register]
2024/11/20 18:00:00 Adding route for [logout]
2024/11/20 18:00:00 Listening on [::]:3986 🚀🚀🚀
```

### Client

#### Register Client to a Room

```bash
# Client 1
go run client/client.go \
  -signal-address=127.0.0.1:3986 \
  -local-address=127.0.0.1:4000 \
  -room-key=game-room-1

# Client 2
go run client/client.go \
  -signal-address=127.0.0.1:3986 \
  -local-address=127.0.0.1:4001 \
  -room-key=game-room-1
```

#### Client Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-signal-address` | Signaling server address | `127.0.0.1:3986` |
| `-local-address` | Client's local UDP address | `127.0.0.1:4000` |
| `-room-key` | Room identifier for matching | `default` |

#### What Happens

1. **Registration**: Client sends register event to server
2. **Discovery**: Server broadcasts all room members to each client
3. **Hole Punching**: Clients send packets to each other's public addresses
4. **P2P Communication**: Direct peer-to-peer connection established

#### Graceful Shutdown

Press `Ctrl+C` to gracefully disconnect:
```
^CGot signal: interrupt
Logout client sent
```

## 🐳 Docker Deployment

### Quick Start - Single Network

```bash
# Start all services (single network - simple testing)
docker-compose up -d

# View logs
docker-compose logs -f client1 client2

# Stop services
docker-compose down
```

### Simple Multi-Network Test (Recommended for macOS)

**Best for testing on macOS Docker Desktop!**

```bash
# Start with separate networks (simplified)
docker-compose -f docker-compose.simple-test.yml up -d

# View logs
docker-compose -f docker-compose.simple-test.yml logs -f client1 client2 client3

# Stop services
docker-compose -f docker-compose.simple-test.yml down
```

**Network Topology**:
```
Server (multi-homed):
├── server-network (10.10.0.0/24)
├── client-network-1 (10.11.0.0/24) → Client1, Client3
└── client-network-2 (10.12.0.0/24) → Client2
```

This configuration demonstrates:
- Clients on different networks discovering each other
- Peer-to-peer communication across network boundaries
- Heartbeat mechanism working correctly

### Advanced - Full NAT Simulation (Linux Only)

**Note**: NAT router simulation with iptables requires Linux. On macOS, use the simple test above.

```bash
# For Linux environments with full NAT simulation
docker-compose -f docker-compose.separate-networks.yml up -d

# View logs
docker-compose -f docker-compose.separate-networks.yml logs -f client1 client2

# Stop services
docker-compose -f docker-compose.separate-networks.yml down
```

For detailed testing scenarios and troubleshooting, see **[NETWORK_TESTING.md](NETWORK_TESTING.md)**.

### Services

#### Single Network Setup (`docker-compose.yml`)
- **udp-server**: Signaling server on port 3986/udp
- **client1, client2**: Test clients on same network
- **redis**: Redis instance (optional)

#### Simple Multi-Network Setup (`docker-compose.simple-test.yml`)
- **udp-server**: Multi-homed signaling server
- **client1, client3**: On client-network-1
- **client2**: On client-network-2

#### Separate Networks Setup (`docker-compose.separate-networks.yml`) - Linux Only
- **udp-server**: Public signaling server
- **nat-router-1, nat-router-2**: NAT routers with iptables (requires Linux)
- **client1, client3**: Behind NAT Router 1
- **client2**: Behind NAT Router 2

### Custom Docker Build

```bash
# Build server image
docker build -t udp-hole-punch-server ./server

# Build client image
docker build -t udp-client ./client

# Run server
docker run -p 3986:3986/udp udp-hole-punch-server
```

## 📡 API Documentation

### Event Protocol

All client-server communication uses JSON-formatted events:

```json
{
  "event": "event_name",
  "payload": "{\"key\":\"value\"}"
}
```

### Events

#### 1. Register

Register a client to a room.

**Event Name**: `register`

**Payload**:
```json
{
  "local_ip": "127.0.0.1:4000",
  "key": "room-key"
}
```

**Response**: Server broadcasts all room members' addresses to each client
```
192.168.1.10:4000,192.168.1.20:4001
```

#### 2. Logout

Remove a client from a room.

**Event Name**: `logout`

**Payload**:
```json
{
  "local_ip": "127.0.0.1:4000",
  "key": "room-key"
}
```

**Response**: Updated member list broadcasted to remaining clients

## 🛠️ Development

### Project Requirements

- Go 1.23+
- Make (optional)

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Coverage

Current coverage: **67.2%**

| Package | Coverage |
|---------|----------|
| config | 68.0% |
| models | 100.0% |
| repositories | 75.0% |
| adapters | 100.0% |
| handlers | 84.6% |
| router | 70.8% |
| server | 36.4% |

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Vet code
go vet ./...
```

### Debug Mode

Enable verbose logging:
```bash
LOG_LEVEL=debug go run cmd/main.go
```

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Maintain test coverage above 60%
- Add tests for new features
- Update documentation for API changes
- Use meaningful commit messages

## 🔐 Security

- Never expose the signaling server directly to the internet without proper firewall rules
- Consider implementing authentication for production use
- Use TLS/DTLS for encrypted P2P connections
- Implement rate limiting to prevent DoS attacks

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by STUN/TURN protocols
- Built with clean architecture principles
- Community contributions and feedback

## 📧 Contact & Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/nat-hole-punch/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/nat-hole-punch/discussions)

---

**Made with ❤️ using Go**
