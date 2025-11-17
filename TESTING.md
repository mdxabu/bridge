# Bridge NAT64 Translator - Testing Guide

This guide provides step-by-step instructions for testing the Bridge NAT64 translator.

## Prerequisites

- Go 1.23 or later installed
- Docker installed and running
- Root/sudo access (for TUN interface creation)
- macOS or Linux operating system

## Quick Test

### 1. Build the Project

```bash
cd /path/to/bridge
make build
```

Expected output:
```
Building bridge...
Build complete: bridge
```

### 2. Initialize Configuration

```bash
./bridge init
```

Expected output:
```
Initializing bridge configuration...
Configuration file created: bridgeconfig.yaml
Default settings:
  NAT64 Prefix: 64:ff9b::/96
  Gateway IP: 64:ff9b::1
  API Port: 8080
```

### 3. Verify Configuration

```bash
cat bridgeconfig.yaml
```

Expected content:
```yaml
interface: ""
nat64_prefix: "64:ff9b::/96"
nat64_gateway: "64:ff9b::1"
api_port: 8080
```

### 4. Setup Docker Networks

```bash
./bridge setup
```

Expected output:
```
Setting up Docker networks for Bridge...
Creating IPv6-only Docker network...
IPv6 network 'bridge-ipv6' created successfully
Creating IPv4-only Docker network...
IPv4 network 'bridge-ipv4' created successfully

Networks created successfully!
Next steps:
  1. Run 'bridge start' to start the NAT64 translator
  2. Connect your IPv6-only containers to 'bridge-ipv6'
  3. Connect your IPv4-only containers to 'bridge-ipv4'
```

### 5. Verify Networks Created

```bash
docker network ls | grep bridge
```

Expected output should include:
```
bridge-ipv6
bridge-ipv4
```

### 6. Test Commands

```bash
# Check help
./bridge --help

# Check version
./bridge version

# Check status
./bridge status
```

## Integration Testing

### Test 1: Docker Network Connectivity

#### Step 1: Create Test Containers

```bash
# Start IPv4 server
docker run -d --name test-ipv4-server \
  --network bridge-ipv4 \
  nginx:alpine

# Start IPv6 client
docker run -d --name test-ipv6-client \
  --network bridge-ipv6 \
  alpine sleep infinity
```

#### Step 2: Get IPv4 Server Address

```bash
docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' test-ipv4-server
```

Example output: `10.64.0.2`

#### Step 3: Convert to NAT64 Address

For IPv4 address `10.64.0.2`:
- Hex: `0a.40.00.02`
- NAT64: `64:ff9b::a40:2`

#### Step 4: Test Connectivity (without Bridge)

```bash
docker exec test-ipv6-client ping6 -c 3 64:ff9b::a40:2
```

Expected: This will fail without Bridge running (connection timeout or network unreachable).

#### Step 5: Cleanup Test Containers

```bash
docker rm -f test-ipv6-client test-ipv4-server
```

### Test 2: API Endpoints

Bridge provides REST API endpoints for monitoring. Note: These will only work when Bridge is running.

```bash
# Health check (when bridge is running)
curl http://localhost:8080/api/health

# Expected response:
# {"status":"healthy","running":true,"uptime":XX.XX}

# Statistics
curl http://localhost:8080/api/stats

# Expected response:
# {"total_sessions":0,"tcp_sessions":0,"udp_sessions":0,...}

# Sessions
curl http://localhost:8080/api/sessions

# Expected response:
# {"sessions":[],"count":0}

# Status
curl http://localhost:8080/api/status

# Expected response:
# {"status":"running","uptime":XX.XX,"start_time":"..."}
```

## Testing with Docker Compose

### Step 1: Use Docker Compose

```bash
docker-compose up -d
```

### Step 2: Check Logs

```bash
docker-compose logs -f bridge
```

### Step 3: Test Connectivity

```bash
# From IPv6 client to IPv4 server
docker exec test-ipv6-client ping6 -c 3 64:ff9b::a40:1

# Install curl in IPv6 client
docker exec test-ipv6-client apk add --no-cache curl

# Test HTTP
docker exec test-ipv6-client curl -v http://[64:ff9b::a40:5]
```

### Step 4: Stop Services

```bash
docker-compose down
```

## Manual Testing (Requires Root)

### Important Note
Starting the bridge requires root privileges for TUN interface creation.

### Step 1: Start Bridge

```bash
sudo ./bridge start
```

Expected output:
```
Starting NAT64 Bridge...
Creating TUN interfaces...
IPv6 TUN interface created: tun-ipv6
IPv4 TUN interface created: tun-ipv4
NAT64 Bridge started successfully
NAT64 Bridge is running
NAT64 Prefix: 64:ff9b::/96
NAT64 Gateway IP: 64:ff9b::1
Press Ctrl+C to stop
```

### Step 2: Monitor in Another Terminal

```bash
# Check API health
watch -n 1 'curl -s http://localhost:8080/api/health | jq'

# Monitor sessions
watch -n 1 'curl -s http://localhost:8080/api/sessions | jq .count'

# View statistics
watch -n 1 'curl -s http://localhost:8080/api/stats | jq'
```

### Step 3: Stop Bridge

Press `Ctrl+C` in the terminal running Bridge.

Expected output:
```
Shutting down...
NAT64 Bridge stopped
Bridge stopped successfully
```

## Troubleshooting

### Issue: "Permission denied" when creating TUN interface

**Solution**: Run with sudo
```bash
sudo ./bridge start
```

### Issue: "Network already exists"

**Solution**: Clean up first
```bash
./bridge cleanup
./bridge setup
```

### Issue: "Port 8080 already in use"

**Solution**: Kill the process using port 8080
```bash
lsof -ti:8080 | xargs kill -9
```

Or edit `bridgeconfig.yaml` to use a different port:
```yaml
api_port: 8081
```

### Issue: Cannot build - missing dependencies

**Solution**: Download dependencies
```bash
go mod download
go mod tidy
```

### Issue: TUN device error on Linux

**Solution**: Load TUN kernel module
```bash
sudo modprobe tun
```

### Issue: Docker network creation fails

**Solution**: Check Docker daemon is running
```bash
docker ps
# If this fails, start Docker daemon
```

## Validation Checklist

Use this checklist to validate your Bridge installation:

- [ ] Project builds successfully (`make build`)
- [ ] Configuration file can be created (`./bridge init`)
- [ ] Docker networks can be created (`./bridge setup`)
- [ ] Help command works (`./bridge --help`)
- [ ] Version command works (`./bridge version`)
- [ ] Status command works (`./bridge status`)
- [ ] Docker networks exist (`docker network ls | grep bridge`)
- [ ] Configuration file exists (`cat bridgeconfig.yaml`)
- [ ] Binary is executable (`ls -la bridge`)

## Expected Behavior

### When Bridge is NOT running:

- API endpoints return connection refused
- IPv6 to IPv4 translation does not work
- Packets between networks are dropped

### When Bridge IS running:

- API endpoints return valid JSON
- TUN interfaces are created
- NAT sessions are tracked
- Statistics are collected
- Packets are translated between IPv6 and IPv4

## Performance Testing

### Test 1: Session Creation

```bash
# Create multiple connections
for i in {1..10}; do
  docker run -d --name client-$i \
    --network bridge-ipv6 \
    alpine ping6 -c 100 64:ff9b::a40:1 &
done

# Monitor sessions
curl http://localhost:8080/api/stats | jq .total_sessions

# Cleanup
for i in {1..10}; do docker rm -f client-$i; done
```

### Test 2: API Response Time

```bash
# Measure API response time
time curl http://localhost:8080/api/health
```

## Testing NAT64 Address Conversion

### IPv4 to NAT64 Examples

| IPv4 Address  | NAT64 Address        |
|--------------|----------------------|
| 10.64.0.1    | 64:ff9b::a40:1      |
| 10.64.0.5    | 64:ff9b::a40:5      |
| 192.0.2.1    | 64:ff9b::c000:201   |
| 8.8.8.8      | 64:ff9b::808:808    |

### Conversion Formula

For IPv4 address `A.B.C.D`:
1. Convert each octet to hex
2. Format as: `64:ff9b::AABB:CCDD`

Example: `10.64.0.5`
- 10 = 0x0a
- 64 = 0x40
- 0 = 0x00
- 5 = 0x05
- Result: `64:ff9b::a40:5`

## Cleanup After Testing

```bash
# Remove test containers
docker ps -a | grep test- | awk '{print $1}' | xargs docker rm -f

# Remove Bridge networks
./bridge cleanup

# Or manually
docker network rm bridge-ipv6 bridge-ipv4

# Remove configuration
rm bridgeconfig.yaml

# Remove binary
rm bridge
```

## CI/CD Testing

The project includes GitHub Actions workflow for automated testing:

```bash
# Run tests locally
make test

# Run with coverage
make test-coverage

# Run linters
make lint

# Run all checks
make test && make lint && make build
```

## Next Steps

After validating basic functionality:

1. Test with real applications
2. Monitor performance under load
3. Test failover scenarios
4. Validate NAT state cleanup
5. Test concurrent connections
6. Benchmark packet translation speed

## Getting Help

If you encounter issues:

1. Check this testing guide
2. Review README.md
3. Check CONTRIBUTING.md for development setup
4. Review build output for errors
5. Check Docker logs: `docker logs <container>`
6. Enable debug logging (if implemented)

## Summary

This testing guide covers:

- Basic functionality tests
- Docker network setup
- API endpoint validation
- Manual bridge testing
- Troubleshooting common issues
- Performance testing basics
- Cleanup procedures

For production deployment, additional testing of security, performance, and reliability is recommended.
