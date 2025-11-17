# Bridge NAT64 - Manual Testing Instructions

## Current Test Setup

**Status**: Bridge is working and TUN interfaces are created successfully!

### Test Environment
- IPv4 Server: `test-ipv4-web` (nginx) at **10.64.0.3**
- IPv6 Client: `test-ipv6-app` (alpine) on bridge-ipv6 network
- NAT64 Address for server: **64:ff9b::a40:3**

### Address Conversion
IPv4: `10.64.0.3`
- 10 = 0x0a = `a`
- 64 = 0x40 = `40`
- 0 = 0x00 = `0`
- 3 = 0x03 = `3`

NAT64: `64:ff9b::a40:3`

## Step-by-Step Testing

### Terminal 1: Start Bridge

```bash
cd /Users/mdxabu/Projects/bridge
sudo ./bridge start
```

**Expected Output:**
```
[INFO] Starting NAT64 Bridge...
[INFO] Creating TUN interfaces...
[SUCCESS] Created IPv6 TUN interface: utun4
[SUCCESS] Created IPv4 TUN interface: utun5
[SUCCESS] NAT64 Bridge started successfully
[SUCCESS] NAT64 Bridge is running
[INFO] NAT64 Prefix: 64:ff9b::/96
[INFO] NAT64 Gateway IP: 64:ff9b::1
[INFO] Press Ctrl+C to stop
```

**Leave this terminal running!**

### Terminal 2: Test Communication

#### Test 1: Check API Health

```bash
curl http://localhost:8080/api/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "running": true,
  "uptime": XX.XX
}
```

#### Test 2: Check Statistics

```bash
curl http://localhost:8080/api/stats
```

**Expected Response:**
```json
{
  "total_sessions": 0,
  "tcp_sessions": 0,
  "udp_sessions": 0,
  "bytes_sent": 0,
  "bytes_received": 0,
  "allocated_ports": 0,
  "uptime": XX.XX
}
```

#### Test 3: Ping from IPv6 to IPv4

```bash
# Test connectivity (without bridge - should fail)
docker exec test-ipv6-app ping6 -c 3 64:ff9b::a40:3
```

**Without Bridge Running:**
```
100% packet loss
```

**With Bridge Running:**
```
Should show successful pings with NAT64 translation
(Note: Requires ICMP translation to be implemented)
```

#### Test 4: HTTP Request from IPv6 to IPv4

```bash
# Install curl in IPv6 container
docker exec test-ipv6-app apk add --no-cache curl

# Test HTTP access via NAT64
docker exec test-ipv6-app curl -v http://[64:ff9b::a40:3]
```

**Expected (with bridge):**
- Connection should be established
- NAT64 translation should occur
- Should receive nginx welcome page

#### Test 5: Monitor Sessions

In another terminal while testing:

```bash
# Watch active sessions
watch -n 1 'curl -s http://localhost:8080/api/sessions | jq'

# Watch statistics
watch -n 1 'curl -s http://localhost:8080/api/stats | jq'
```

### Terminal 3: Monitor Bridge Logs

```bash
# Check bridge output
# Bridge will show debug messages about:
# - Packet capture
# - Translation attempts
# - Session creation
# - NAT mappings
```

## Test Scenarios

### Scenario 1: Basic Connectivity Test

1. Start Bridge (Terminal 1)
2. Check API health (Terminal 2)
3. Verify TUN interfaces created:
   ```bash
   ifconfig | grep utun
   ```
4. Test ping from IPv6 to IPv4
5. Check statistics for session count

### Scenario 2: HTTP Communication Test

1. Start Bridge
2. Install curl in IPv6 container
3. Make HTTP request to IPv4 server via NAT64
4. Monitor sessions in real-time
5. Verify packets are translated

### Scenario 3: Multiple Connections

1. Start Bridge
2. Create multiple IPv6 clients:
   ```bash
   for i in {1..5}; do
     docker run -d --name ipv6-client-$i \
       --network bridge-ipv6 \
       alpine sleep infinity
   done
   ```
3. Make concurrent connections
4. Monitor session table growth
5. Check port allocation

### Scenario 4: Session Timeout Test

1. Start Bridge
2. Create a connection
3. Let it idle
4. Monitor session cleanup after timeout
5. Verify port is released

## Verification Commands

### Check Bridge is Running

```bash
# Check process
ps aux | grep bridge

# Check API
curl http://localhost:8080/api/health

# Check TUN interfaces
ifconfig | grep utun
```

### Check Docker Networks

```bash
docker network ls | grep bridge
docker network inspect bridge-ipv6
docker network inspect bridge-ipv4
```

### Check Containers

```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Networks}}"
```

### Check NAT Sessions

```bash
curl http://localhost:8080/api/sessions | jq '.sessions | length'
```

## Expected Behavior

### When Bridge is Running:

✓ API responds on port 8080
✓ TUN interfaces created (utun4, utun5)
✓ Packets are captured from TUN interfaces
✓ IPv6 → IPv4 translation occurs
✓ NAT sessions are created and tracked
✓ Statistics are collected
✓ Sessions timeout after inactivity

### What Should Work:

- ✓ Configuration loading
- ✓ TUN interface creation
- ✓ API endpoints
- ✓ Packet parsing
- ✓ NAT table management
- ✓ Session tracking
- ⚠ TCP/UDP translation (requires actual packet flow)
- ✗ ICMP translation (not yet implemented)

## Troubleshooting

### Bridge Won't Start

```bash
# Check if port 8080 is in use
lsof -i :8080

# Check for permission issues
ls -la bridge

# Verify config
cat bridgeconfig.yaml
```

### No Translation Happening

1. Verify bridge is running
2. Check TUN interfaces exist: `ifconfig | grep utun`
3. Monitor bridge output for errors
4. Check session count: `curl localhost:8080/api/stats`
5. Verify packet capture is working

### Containers Can't Communicate

1. Verify containers are on correct networks
2. Check NAT64 address calculation
3. Verify bridge is receiving packets
4. Check session table for entries

## Cleanup

### Stop Bridge
In Terminal 1, press `Ctrl+C`

### Remove Test Containers
```bash
docker rm -f test-ipv4-web test-ipv6-app
```

### Remove Networks
```bash
./bridge cleanup
```

## Success Criteria

- [ ] Bridge starts without errors
- [ ] TUN interfaces created (utun4, utun5)
- [ ] API responds to health check
- [ ] Statistics endpoint works
- [ ] Sessions endpoint works
- [ ] Packets are captured (shown in bridge output)
- [ ] NAT sessions are created
- [ ] Translation logic executes
- [ ] No crashes or panics

## Current Test Status

**Bridge Functionality:** ✓ Working
- TUN interfaces: ✓ Created
- API Server: ✓ Running
- Packet Capture: ✓ Configured
- NAT Table: ✓ Initialized

**Ready for:** Live packet translation testing

## Next Steps

1. Run the bridge in Terminal 1
2. Test API endpoints in Terminal 2
3. Attempt IPv6 to IPv4 communication
4. Monitor sessions and statistics
5. Document any issues or successes
6. Implement ICMP translation if needed
