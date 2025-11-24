# 🚀 Bridge NAT64 - Quick Reference Cheat Sheet

## 📦 Current Docker Setup

### Running Containers
```
bridge-nat64       → NAT64 Translator (10.64.0.2 + fd00:64::2)
test-ipv6-client   → IPv6 Client (fd00:64::3)
test-ipv4-server   → IPv4 Server (10.64.0.3 + Nginx)
```

### Networks
```
bridge-ipv6  → fd00:64::/64   (IPv6 enabled)
bridge-ipv4  → 10.64.0.0/16   (IPv4 only)
```

---

## ⚡ Quick Commands

### Docker Operations
```bash
# Start everything
docker-compose up -d

# Stop everything
docker-compose down

# View logs
docker logs bridge-nat64 -f
docker logs test-ipv6-client
docker logs test-ipv4-server

# Check status
docker ps

# Rebuild
docker-compose up -d --build
```

### Network Inspection
```bash
# See all networks
docker network ls | grep bridge

# Inspect IPv6 network
docker network inspect bridge-ipv6

# Inspect IPv4 network
docker network inspect bridge-ipv4

# Container IPs
docker inspect test-ipv6-client | grep IPAddress
docker inspect test-ipv4-server | grep IPAddress
docker inspect bridge-nat64 | grep IPAddress
```

### Container Access
```bash
# Access IPv6 client
docker exec -it test-ipv6-client sh

# Access IPv4 server
docker exec -it test-ipv4-server sh

# Access bridge
docker exec -it bridge-nat64 sh
```

### Testing Commands
```bash
# Ping from IPv6 client to bridge
docker exec test-ipv6-client ping6 -c 3 fd00:64::2

# Check IPv6 client interfaces
docker exec test-ipv6-client ip addr show

# Check bridge interfaces
docker exec bridge-nat64 ip addr show

# View bridge TUN interfaces
docker exec bridge-nat64 ip link show tun0
docker exec bridge-nat64 ip link show tun1
```

---

## 🔢 NAT64 Address Conversion

### Formula
```
IPv4: A.B.C.D
  ↓
NAT64: 64:ff9b::AABB:CCDD (hex)
```

### Examples
```
10.64.0.3    → 64:ff9b::a40:3
192.168.1.1  → 64:ff9b::c0a8:101
8.8.8.8      → 64:ff9b::808:808
```

### Python Calculator
```python
# Convert IPv4 to NAT64
ipv4 = "10.64.0.3"
octets = [int(x) for x in ipv4.split(".")]
nat64 = f"64:ff9b::{octets[0]:02x}{octets[1]:02x}:{octets[2]:02x}{octets[3]:02x}"
print(nat64)  # 64:ff9b::0a40:0003
```

### Bash One-liner
```bash
python3 -c "ip='10.64.0.3'; o=[int(x) for x in ip.split('.')]; print(f'64:ff9b::{o[0]:02x}{o[1]:02x}:{o[2]:02x}{o[3]:02x}')"
```

---

## 🏗️ Architecture Quick View

```
┌────────────────────────────────────────────────────┐
│                  Docker Host                       │
│                                                    │
│  IPv6 Network (fd00:64::/64)                      │
│  ┌──────────┐           ┌──────────┐             │
│  │ IPv6     │ ◄────────►│ Bridge   │             │
│  │ Client   │           │ NAT64    │             │
│  │          │           │          │             │
│  │fd00:64::3│           │fd00:64::2│             │
│  └──────────┘           │          │             │
│                         │  TUN0/1  │             │
│                         │          │             │
│  IPv4 Network           │10.64.0.2 │             │
│  (10.64.0.0/16)        └──────────┘             │
│                             │                     │
│                         ┌──────────┐             │
│                         │ IPv4     │             │
│                         │ Server   │             │
│                         │ (Nginx)  │             │
│                         │10.64.0.3 │             │
│                         └──────────┘             │
└────────────────────────────────────────────────────┘
```

---

## 📁 Key Files to Know

### Source Code
```
cmd/start.go              → Start command implementation
internal/nat/table.go     → NAT session management
internal/translator/      → Packet conversion logic
internal/tun/bridge.go    → Main bridge orchestrator
```

### Configuration
```
bridgeconfig.yaml         → Bridge settings
docker-compose.yml        → Container orchestration
Dockerfile                → Multi-stage build
```

### Documentation
```
README.md                 → Project overview
DEMO/INTERVIEW_GUIDE.md   → Complete interview prep
DEMO/QUICK_REFERENCE.md   → This file
```

---

## 🎯 NAT Session Structure

```go
SessionState {
    ID:              "tcp:fd00:64::3:45678->64:ff9b::a40:3:80"
    Protocol:        6 (TCP=6, UDP=17, ICMP=1)
    IPv6SrcIP:       fd00:64::3
    IPv6SrcPort:     45678
    IPv6DstIP:       64:ff9b::a40:3
    IPv6DstPort:     80
    IPv4SrcIP:       10.64.0.1      // NAT gateway
    IPv4SrcPort:     10000          // Allocated from pool
    IPv4DstIP:       10.64.0.3
    IPv4DstPort:     80
    State:           "NEW" / "ESTABLISHED" / "CLOSING"
    CreatedAt:       timestamp
    LastActivity:    timestamp
    BytesSent:       counter
    BytesReceived:   counter
}
```

---

## 🔄 Packet Flow Summary

### Outbound (IPv6 → IPv4)
```
1. IPv6 packet arrives at bridge
2. Parse headers, extract IPs/ports
3. Check if destination is NAT64 (64:ff9b::)
4. Create/lookup NAT session
5. Allocate port (10000-65000)
6. Translate IPv6 → IPv4 headers
7. Recalculate checksum
8. Forward to IPv4 network
```

### Inbound (IPv4 → IPv6)
```
1. IPv4 response arrives
2. Parse headers
3. Lookup session by destination port
4. Find original IPv6 client
5. Translate IPv4 → IPv6 headers
6. Recalculate checksum
7. Forward to IPv6 network
```

---

## 📊 Key Metrics & Limits

### Port Pool
- Range: 10,000 - 65,000
- Total: 55,000 concurrent sessions
- Protocol: TCP, UDP, ICMP

### Timeouts
- TCP: 5 minutes (300 seconds)
- UDP: 1 minute (60 seconds)
- Cleanup: Every 30 seconds

### Header Sizes
- IPv6: 40 bytes (fixed)
- IPv4: 20 bytes (minimum, no options)

### NAT64 Prefix
- Standard: 64:ff9b::/96 (RFC 6052)
- Well-known prefix for IPv4-embedded IPv6

---

## 🐛 Troubleshooting

### Bridge not starting
```bash
# Check logs
docker logs bridge-nat64

# Common issues:
# - TUN interface creation needs privileged mode
# - Configuration file missing
```

### Containers can't communicate
```bash
# Verify networks exist
docker network ls

# Check container network assignments
docker inspect <container> | grep NetworkMode

# Ensure bridge is on both networks
docker inspect bridge-nat64 --format '{{json .NetworkSettings.Networks}}'
```

### TUN interfaces down
```bash
# Inside bridge container
docker exec bridge-nat64 ip link show

# TUN interfaces should show:
# tun0: <POINTOPOINT,MULTICAST,NOARP>
# tun1: <POINTOPOINT,MULTICAST,NOARP>
```

---

## 💡 Interview Key Points

### 30-Second Elevator Pitch
"Built a user-space NAT64 translator in Go that bridges IPv6 and IPv4 Docker networks. Uses TUN interfaces for packet capture, maintains stateful NAT sessions with port allocation, and handles concurrent packet translation with goroutines. No kernel modules needed—runs entirely in Docker."

### Technical Highlights
1. ✅ User-space implementation (portable)
2. ✅ Stateful NAT with connection tracking
3. ✅ Thread-safe session management (RWMutex)
4. ✅ TUN/TAP interface handling
5. ✅ Docker-native deployment
6. ✅ Concurrent packet processing
7. ✅ Automatic session cleanup

### Why It Matters
- Enables IPv6 adoption in mixed environments
- No host system modifications required
- Perfect for containerized microservices
- Educational tool for understanding NAT64

---

## 📚 RFCs Referenced

- **RFC 6052** - IPv6 Addressing of IPv4/IPv6 Translators
- **RFC 6145** - IP/ICMP Translation Algorithm
- **RFC 6146** - Stateful NAT64

---

## 🎓 Technologies Used

| Technology | Version | Purpose |
|------------|---------|---------|
| Go         | 1.23+   | Main language |
| Docker     | Latest  | Containerization |
| Cobra      | 1.9.1   | CLI framework |
| water      | Latest  | TUN interface |
| Alpine     | Latest  | Base image |

---

## ✅ Pre-Interview Checklist

- [ ] Can explain NAT64 address format
- [ ] Understand packet flow (both directions)
- [ ] Know NAT session lifecycle
- [ ] Explain goroutine architecture
- [ ] Describe thread safety approach
- [ ] Demo: Start with docker-compose
- [ ] Show: NAT table code
- [ ] Show: Packet translator code
- [ ] Discuss: Trade-offs made
- [ ] Mention: Future improvements

---

## 🚀 Quick Demo Commands

```bash
# Complete demo in 2 minutes
cd bridge

# 1. Start (10 seconds)
docker-compose up -d

# 2. Show running (5 seconds)
docker ps

# 3. Check bridge logs (10 seconds)
docker logs bridge-nat64

# 4. Show networks (10 seconds)
docker network inspect bridge-ipv6 | jq '.IPAM.Config'
docker network inspect bridge-ipv4 | jq '.IPAM.Config'

# 5. Test connectivity (15 seconds)
docker exec test-ipv6-client ping6 -c 3 fd00:64::2

# 6. Show container IPs (10 seconds)
docker inspect test-ipv6-client | grep -E "IPAddress|GlobalIPv6"
docker inspect test-ipv4-server | grep IPAddress

# 7. Calculate NAT64 address (10 seconds)
python3 -c "print('10.64.0.3 → 64:ff9b::a40:3')"

# 8. Show TUN interfaces (10 seconds)
docker exec bridge-nat64 ip link show | grep tun

# 9. Explain code structure (30 seconds)
tree -L 2 bridge/

# 10. Cleanup (10 seconds)
docker-compose down
```

---

**Remember**: Confidence comes from understanding! You know this project inside-out. Good luck! 🎉