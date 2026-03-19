# Redis Integration Guide

This guide explains how to use Redis as the storage backend for the NAT Hole Punch server.

## Why Redis?

- **Persistence**: Client registrations survive server restarts
- **Scalability**: Multiple server instances can share the same Redis
- **Distributed**: Deploy servers across different regions
- **TTL Support**: Automatic client expiration with Redis TTL
- **High Performance**: In-memory storage with optional disk persistence

## Configuration

### Environment Variables

Set `REPOSITORY_TYPE=redis` to enable Redis backend:

```bash
# Redis Configuration
REPOSITORY_TYPE=redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your_password_if_needed
REDIS_DB=0
```

### Docker Compose

Use the Redis-enabled configuration:

```bash
# Start with Redis backend
docker-compose -f docker-compose.redis.yml up -d

# View logs
docker-compose -f docker-compose.redis.yml logs -f

# Stop services
docker-compose -f docker-compose.redis.yml down
```

## Local Development

### 1. Start Redis

```bash
# Using Docker
docker run -d -p 6379:6379 --name redis redis:7-alpine

# Or using Homebrew (macOS)
brew install redis
redis-server

# Or using apt (Linux)
sudo apt install redis-server
sudo systemctl start redis
```

### 2. Configure Server

```bash
cd server
export REPOSITORY_TYPE=redis
export REDIS_ADDR=localhost:6379
go run cmd/main.go
```

Expected output:
```
2024/11/20 18:00:00 Using Redis repository at localhost:6379
2024/11/20 18:00:00 Starting UDP Hole Punch Server on 0.0.0.0:3986
```

## Redis Data Structure

### Keys

The server uses the following key pattern:
```
room:{room-key}:clients
```

Example:
```
room:game-room-1:clients
room:lobby-5:clients
```

### Data Format

Each client is stored as a JSON object in a Redis Set:

```json
{
  "remote_addr": "192.168.1.10:4000",
  "create_at": 1700000000
}
```

### TTL

Room keys automatically expire based on `CLIENT_TTL` configuration (default 60 seconds).

## Redis Commands

### Monitor Client Registrations

```bash
# Connect to Redis CLI
redis-cli

# List all room keys
KEYS room:*:clients

# View clients in a specific room
SMEMBERS room:game-room-1:clients

# Check TTL for a room
TTL room:game-room-1:clients

# Monitor real-time operations
MONITOR
```

### Debugging

```bash
# Check if server is connected
redis-cli PING
# Expected: PONG

# View all keys
redis-cli KEYS '*'

# Get number of clients in a room
redis-cli SCARD room:game-room-1:clients

# Manually clear all data
redis-cli FLUSHDB
```

## Production Deployment

### Redis Configuration

For production, use a properly configured Redis instance:

```bash
# redis.conf
bind 0.0.0.0
protected-mode yes
requirepass your_strong_password
maxmemory 256mb
maxmemory-policy allkeys-lru
appendonly yes
appendfsync everysec
```

### High Availability

Consider Redis Sentinel or Redis Cluster for production:

```yaml
# Redis Sentinel
services:
  redis-sentinel:
    image: redis:7-alpine
    command: redis-sentinel /etc/redis/sentinel.conf
    volumes:
      - ./sentinel.conf:/etc/redis/sentinel.conf
```

### Security

1. **Authentication**: Always use password in production
2. **Network Isolation**: Use private networks
3. **Encryption**: Enable TLS for Redis connections
4. **Firewall**: Restrict access to Redis port (6379)

### Monitoring

Monitor Redis health and performance:

```bash
# Redis stats
redis-cli INFO

# Memory usage
redis-cli INFO memory

# Connected clients
redis-cli CLIENT LIST

# Slow queries
redis-cli SLOWLOG GET 10
```

## Performance Tuning

### Connection Pooling

The Redis client automatically manages connection pooling. Configure pool size if needed:

```go
// In redis-repository.go
client := redis.NewClient(&redis.Options{
    Addr:         config.Addr,
    Password:     config.Password,
    DB:           config.DB,
    PoolSize:     100,  // Max connections
    MinIdleConns: 10,   // Min idle connections
})
```

### Memory Optimization

```bash
# Set max memory
redis-cli CONFIG SET maxmemory 256mb

# Set eviction policy
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

## Testing

### Integration Tests

Run Redis integration tests:

```bash
# Make sure Redis is running
docker run -d -p 6379:6379 redis:7-alpine

# Run tests
cd server
go test ./pkg/repositories/adapters -v -run Redis

# Run with race detection
go test ./pkg/repositories/adapters -v -race -run Redis
```

### Skip Integration Tests

```bash
# Skip Redis tests (uses in-memory)
go test ./... -short
```

## Troubleshooting

### Connection Refused

**Problem**: `Failed to initialize Redis repository: dial tcp: connection refused`

**Solutions**:
1. Check if Redis is running: `redis-cli PING`
2. Verify Redis address: `REDIS_ADDR=localhost:6379`
3. Check firewall rules
4. Ensure Redis is listening: `redis-cli CONFIG GET bind`

### Authentication Failed

**Problem**: `NOAUTH Authentication required`

**Solution**: Set Redis password:
```bash
export REDIS_PASSWORD=your_password
```

### Fallback Behavior

If Redis connection fails, the server automatically falls back to in-memory storage:

```
Failed to initialize Redis repository: connection refused, falling back to in-memory
Using in-memory repository
```

## Comparing Storage Backends

| Feature | In-Memory | Redis |
|---------|-----------|-------|
| **Persistence** | ❌ Lost on restart | ✅ Survives restarts |
| **Scalability** | ❌ Single instance | ✅ Distributed |
| **Performance** | ⚡ Fastest | ⚡ Very fast |
| **Setup** | ✅ Zero config | ⚠️ Requires Redis |
| **Memory** | ✅ Minimal | ⚠️ Redis overhead |
| **Use Case** | Development, testing | Production, multi-instance |

## Migration

### From In-Memory to Redis

No data migration needed - just change the environment variable:

```bash
# Before
REPOSITORY_TYPE=memory

# After
REPOSITORY_TYPE=redis
REDIS_ADDR=localhost:6379
```

All new registrations will use Redis.

### From Redis to In-Memory

Remove Redis configuration:

```bash
unset REPOSITORY_TYPE  # Defaults to memory
# OR
export REPOSITORY_TYPE=memory
```

**Note**: Existing Redis data is not migrated to memory.

## Resources

- [Redis Documentation](https://redis.io/documentation)
- [go-redis Client](https://github.com/redis/go-redis)
- [Redis Best Practices](https://redis.io/topics/best-practices)
