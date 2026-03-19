# Network Testing Guide

This guide explains how to test NAT hole punching with different network configurations.

## 🎯 Testing Scenarios

### Scenario 1: Single Network (Simple Testing)
All clients and server are on the same network - **easiest to test but doesn't simulate real NAT**.

**File**: `docker-compose.yml`

```bash
docker-compose up -d
docker-compose logs -f client1 client2
```

**Expected Behavior**:
- Clients can discover each other immediately
- Direct peer-to-peer communication works
- No actual NAT traversal happening

---

### Scenario 2: Separate Networks (Real NAT Simulation)
Server on "public" network, clients behind separate NAT routers - **simulates real-world scenario**.

**File**: `docker-compose.separate-networks.yml`

```bash
docker-compose -f docker-compose.separate-networks.yml up -d
docker-compose -f docker-compose.separate-networks.yml logs -f client1 client2
```

**Network Topology**:
```
┌─────────────────┐
│  Public Network │  (172.20.0.0/16)
│   172.20.0.10   │  ← Server
│   172.20.0.20   │  ← NAT Router 1
│   172.20.0.30   │  ← NAT Router 2
└─────────────────┘
         │
    ┌────┴────┐
    │         │
┌───▼──┐  ┌──▼───┐
│NAT-1 │  │NAT-2 │
│Router│  │Router│
└───┬──┘  └──┬───┘
    │         │
┌───▼──────┐ ┌▼─────────┐
│Private-1 │ │Private-2 │
│172.21.0/16│ │172.22.0/16│
│ Client1  │ │ Client2  │
│ Client3  │ │          │
└──────────┘ └──────────┘
```

**Expected Behavior**:
- Clients register through their NAT routers
- Server sees clients' public IPs (NAT router IPs)
- Hole punching establishes direct P2P communication
- True NAT traversal demonstration

---

### Scenario 3: Services Only (Minimal Setup)
Server and clients without Redis - **lightweight testing**.

**File**: `docker-compose.services.yml`

```bash
docker-compose -f docker-compose.services.yml up -d
docker-compose -f docker-compose.services.yml logs -f
```

---

## 📊 Comparing Network Configurations

| Aspect | Single Network | Separate Networks | Services Only |
|--------|----------------|-------------------|---------------|
| **Complexity** | Low | High | Low |
| **NAT Simulation** | ❌ No | ✅ Yes | ❌ No |
| **Real-world Accuracy** | ❌ Poor | ✅ Excellent | ❌ Poor |
| **Setup Time** | Fast | Slower | Fast |
| **Resource Usage** | Low | High (5 containers) | Low |
| **Best For** | Quick testing | Production validation | CI/CD testing |

---

## 🔍 Monitoring & Debugging

### Check Network Configuration
```bash
# View network details
docker network ls
docker network inspect nat-hole-punch_public-network

# Check NAT router configuration
docker exec nat-router-1 iptables -t nat -L -v
docker exec nat-router-1 ip route
```

### View Client Logs
```bash
# Follow specific client
docker logs -f client1-nat1

# View all clients
docker-compose -f docker-compose.separate-networks.yml logs -f client1 client2 client3
```

### Test Network Connectivity
```bash
# From Client1 to Server (through NAT)
docker exec client1-nat1 ping -c 3 172.20.0.10

# Check routing
docker exec client1-nat1 ip route
docker exec client1-nat1 traceroute 172.20.0.10
```

### Verify NAT Translation
```bash
# Check NAT rules
docker exec nat-router-1 iptables -t nat -L POSTROUTING -v -n

# Monitor traffic
docker exec nat-router-1 tcpdump -i eth0 udp port 3986
```

---

## 🧪 Test Cases

### Test 1: Basic Registration
**Goal**: Verify clients can register through NAT

```bash
# Start services
docker-compose -f docker-compose.separate-networks.yml up -d

# Watch registration
docker logs -f client1-nat1 2>&1 | grep -i "registr"
```

**Expected Output**:
```
Registering to room 'nat-test-room'...
Registration sent (XX bytes)
```

---

### Test 2: Peer Discovery
**Goal**: Verify server broadcasts peer information

```bash
# Watch for incoming peer lists
docker logs -f client1-nat1 2>&1 | grep -i "incoming"
```

**Expected Output**:
```
[INCOMING] From 172.20.0.10:3986: 172.21.0.11:4002,172.22.0.10:4001
Discovered peer: 172.21.0.11:4002, starting communication...
Discovered peer: 172.22.0.10:4001, starting communication...
```

---

### Test 3: Hole Punching
**Goal**: Verify P2P communication establishes

```bash
# Monitor heartbeats
docker logs -f client1-nat1 2>&1 | grep -i "heartbeat"
```

**Expected Output**:
```
Sent heartbeat to 172.22.0.10:4001 (6 bytes)
Received heartbeat from 172.22.0.10:4001
```

---

### Test 4: Graceful Shutdown
**Goal**: Verify clean disconnect

```bash
# Stop a client
docker stop client1-nat1

# Check logout in server logs
docker logs udp-server-public 2>&1 | grep -i "logout"
```

---

## 🐛 Troubleshooting

### Problem: Clients Can't Register
**Symptoms**: No registration messages in logs

**Solutions**:
1. Check server is running: `docker ps | grep udp-server`
2. Verify network connectivity: `docker exec client1-nat1 ping 172.20.0.10`
3. Check NAT router: `docker logs nat-router-1`

---

### Problem: No Peer Discovery
**Symptoms**: Clients register but don't see peers

**Solutions**:
1. Check server logs: `docker logs udp-server-public`
2. Verify all clients use same room key
3. Check buffer size settings

---

### Problem: Hole Punching Fails
**Symptoms**: Peers discovered but no heartbeats

**Solutions**:
1. Verify NAT rules: `docker exec nat-router-1 iptables -t nat -L -v`
2. Check UDP ports are open
3. Monitor with tcpdump: `docker exec nat-router-1 tcpdump -i any udp`

---

### Problem: High Resource Usage
**Symptoms**: Docker consuming too much CPU/memory

**Solutions**:
1. Reduce heartbeat frequency in client code
2. Use services-only composition
3. Limit number of concurrent clients

---

## 📈 Performance Testing

### Load Testing
```bash
# Start multiple clients
for i in {1..10}; do
  docker run -d --name client$i \
    --network nat-hole-punch_nat-network-1 \
    udp-client:latest \
    -signal-address=172.20.0.10:3986 \
    -local-address=0.0.0.0:$((4000+i)) \
    -room-key=load-test
done
```

### Latency Testing
```bash
# Measure round-trip time
docker exec client1-nat1 sh -c 'time echo "test" | nc -u 172.22.0.10 4001'
```

---

## 🔧 Advanced Configuration

### Custom Network Subnets
Edit `docker-compose.separate-networks.yml`:
```yaml
networks:
  public-network:
    ipam:
      config:
        - subnet: 10.0.0.0/24  # Your custom subnet
```

### Multiple NAT Layers
Add additional NAT routers for multi-level NAT testing:
```yaml
nat-router-1-inner:
  image: alpine:latest
  cap_add:
    - NET_ADMIN
  # Configure as nested NAT
```

### Firewall Rules Testing
Add restrictive iptables rules:
```bash
docker exec nat-router-1 iptables -A FORWARD -p udp --dport 4000 -j DROP
```

---

## 📝 Cleanup

```bash
# Stop and remove all containers
docker-compose -f docker-compose.separate-networks.yml down

# Remove networks
docker network prune

# Clean up volumes
docker volume prune
```

---

## 🎓 Learning Resources

- **NAT Types**: Full Cone, Restricted Cone, Port Restricted, Symmetric
- **STUN Protocol**: Session Traversal Utilities for NAT
- **TURN Protocol**: Traversal Using Relays around NAT
- **ICE Framework**: Interactive Connectivity Establishment

---

## ✅ Quick Reference

### Common Commands
```bash
# Start separate networks
docker-compose -f docker-compose.separate-networks.yml up -d

# View logs
docker-compose -f docker-compose.separate-networks.yml logs -f

# Stop services
docker-compose -f docker-compose.separate-networks.yml down

# Check NAT
docker exec nat-router-1 iptables -t nat -L -v

# Monitor traffic
docker exec nat-router-1 tcpdump -i any udp port 3986
```

### Network IPs
- **Server**: 172.20.0.10
- **NAT Router 1**: 172.20.0.20 (public), 172.21.0.1 (private)
- **NAT Router 2**: 172.20.0.30 (public), 172.22.0.1 (private)
- **Client 1**: 172.21.0.10
- **Client 2**: 172.22.0.10
- **Client 3**: 172.21.0.11
