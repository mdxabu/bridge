# Bridge - User-Space NAT64 Translator for Docker

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go" alt="Go Version" />
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker" alt="Docker" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License" />
</p>

**Bridge** is a Golang-based NAT64 translator designed to enable seamless communication between IPv6-only and IPv4-only Docker containers — without relying on external VMs, kernel modules, or system-level NAT64 daemons.

It acts as a **user-space packet translator**, running as a lightweight Docker container, and dynamically converts IPv6 traffic into IPv4 traffic (and vice versa) at the TCP/UDP level.

This allows developers to deploy mixed IPv4/IPv6 environments locally, test dual-stack applications, and experiment with modern networking stacks — all within Docker on **macOS, Linux, or Windows**.

## How It Works

```
```
┌─────────────┐     IPv6 Packet      ┌──────────────────┐     IPv4 Packet      ┌─────────────┐
│ IPv6-only   │ ───────────────────> │  Bridge NAT64    │ ───────────────────> │ IPv4-only   │
│  Client     │                      │  Translator      │                      │   Server    │
│             │ <─────────────────── │  (Go process)    │ <─────────────────── │             │
└─────────────┘     IPv6 Packet      └──────────────────┘     IPv4 Packet      └─────────────┘
                                      ┌──────────────┐
                                      │ NAT State    │
                                      │ Table        │
                                      │ Port Mapping │
                                      └──────────────┘
```

Bridge intercepts IPv6 packets from an IPv6-only Docker network, rewrites the headers into IPv4 form, and forwards them to an IPv4-only network — maintaining connection state, transport checksums, and NAT mappings for reverse translation.

## Features

- **User-space NAT64** — Fully written in Go, no kernel modules required
- **Docker-native** — Runs entirely inside a container, attaches to IPv4 and IPv6 Docker networks
- **Dual-network bridging** — Translates live traffic between IPv4-only and IPv6-only containers
- **Connection tracking** — Maintains NAT state tables for TCP and UDP sessions
- **Metrics & Monitoring** — Real-time statistics via REST API and dashboard
- **Extensible CLI** — Manage setup, start/stop bridge, and view stats via `bridge` CLI tool
- **Cross-platform** — Works on macOS, Linux, and Windows (through Docker)
- **Stateful translation** — Tracks ports, sessions, timeouts, and bandwidth

## Architecture

| Component            | Description                                                                            |
| -------------------- | -------------------------------------------------------------------------------------- |
| **Bridge CLI**       | Command-line tool written in Go to manage Docker networks and the translator container |
| **Bridge Container** | Runs the NAT64 Go process; attaches to both IPv4 and IPv6 Docker networks              |
| **IPv6 Network**     | Docker network hosting IPv6-only containers                                            |
| **IPv4 Network**     | Docker network hosting IPv4-only containers                                            |
| **Translator Core**  | User-space engine handling IPv6↔IPv4 packet translation and NAT state management       |
| **NAT State Table**  | Tracks active sessions, port mappings, and connection timeouts                         |
| **TUN Interfaces**   | Virtual network interfaces for packet capture and injection                            |
| **Metrics API**      | REST API exposing statistics, sessions, and health status                              |

## Quick Start

### Prerequisites

- **Go 1.23+** installed
- **Docker** installed and running
- **Root/Administrator privileges** (for TUN interface creation)

### Installation

```bash
# Clone the repository
git clone https://github.com/mdxabu/bridge.git
cd bridge

# Build the CLI
go build -o bridge .

# Make it executable (Unix/macOS)
chmod +x bridge

# Optional: Install globally
sudo mv bridge /usr/local/bin/
```

### Basic Usage

```bash
# 1. Initialize configuration
bridge init

# 2. Setup Docker networks
bridge setup

# 3. Start the NAT64 translator (requires root)
sudo bridge start

# 4. In another terminal, check status
bridge status

# 5. View metrics
bridge metrics

# 6. Stop the bridge
# Press Ctrl+C in the bridge terminal
```

## Commands

### Network Setup

```bash
# Create required Docker networks
bridge setup

# Remove Docker networks
bridge cleanup
```

### Bridge Operations

```bash
# Start the NAT64 bridge
sudo bridge start

# Check bridge status
bridge status

# View active NAT sessions
bridge sessions

# Display real-time metrics
bridge metrics

# Stop the bridge
# Use Ctrl+C or send SIGTERM
```

### Configuration

```bash
# Initialize default configuration
bridge init

# Run NAT64 translation (legacy ping test)
bridge nat64

# Perform DNS64 resolution
bridge dns <domain>

# View version
bridge version
```

## Configuration

The configuration file `bridgeconfig.yaml` contains:

```yaml
interface: ""                    # Network interface to use (auto-detect if empty)
nat64_ip: 64:ff9b::7f00:1       # NAT64 gateway IPv6 address
dest_ip_path: ipv4.txt          # File containing IPv4 destinations
dest_domain_path: domains.txt    # File containing domains for DNS64
```

## Technical Highlights

### Core Technologies

- **TUN Interfaces** — Using `github.com/songgao/water` for packet-level access
- **IP Header Parsing** — Built with `golang.org/x/net/ipv4` and `ipv6` packages
- **NAT State Management** — Custom state tables with port allocation and timeouts
- **DNS64 Synthesis** — Optional IPv6 address synthesis for IPv4-only domains
- **Metrics Collection** — Real-time stats via REST API

### Translation Process

1. **Packet Capture** — TUN interface captures IPv6 packets
2. **Header Parsing** — Extract IP addresses, ports, and protocol
3. **NAT Session Lookup** — Check existing sessions or create new mapping
4. **Address Translation** — Convert IPv6 addresses to/from NAT64 format (64:ff9b::/96)
5. **Header Rewriting** — Build new IPv4/IPv6 headers with correct checksums
6. **Forwarding** — Inject translated packet into destination TUN interface
7. **State Update** — Update session statistics and last activity time

### NAT64 Address Format

Bridge uses the well-known NAT64 prefix **64:ff9b::/96** (RFC 6052):

```
IPv4: 192.0.2.1
  ↓
IPv6: 64:ff9b::c000:201
```

## Metrics & Monitoring

### REST API Endpoints

```bash
# Health check
curl http://localhost:8080/api/health

# Statistics
curl http://localhost:8080/api/stats

# Active sessions
curl http://localhost:8080/api/sessions

# Status
curl http://localhost:8080/api/status
```

### Example Stats Response

```json
{
  "total_sessions": 42,
  "tcp_sessions": 30,
  "udp_sessions": 12,
  "bytes_sent": 1048576,
  "bytes_received": 2097152,
  "allocated_ports": 42,
  "uptime": 3600.5
}
```

## Use Cases

- Testing dual-stack applications in Docker
- Local NAT64 gateway for microservice development  
- Educational tool for understanding IPv6/IPv4 translation
- IPv6 adoption experimentation without touching host network stack
- Simulating production NAT64 environments locally
- Testing IPv6-only client applications against IPv4 APIs

## Docker Example

### Running IPv6 and IPv4 Containers

```bash
# Start Bridge
sudo bridge start

# Run IPv6-only container
docker run -d --name ipv6-client \
  --network bridge-ipv6 \
  alpine sleep infinity

# Run IPv4-only container  
docker run -d --name ipv4-server \
  --network bridge-ipv4 \
  -p 8080:80 \
  nginx

# Test connectivity from IPv6 container
docker exec ipv6-client ping6 64:ff9b::a40:1  # Pings 10.64.0.1
```

## Development

### Project Structure

```
bridge/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── start.go           # Start bridge
│   ├── docker.go          # Docker network management
│   ├── nat64.go           # NAT64 operations
│   └── ...
├── internal/
│   ├── translator/        # Packet translation logic
│   │   ├── packet.go      # Packet parsing
│   │   ├── converter.go   # IPv6<->IPv4 conversion
│   │   └── nat64.go       # NAT64 address handling
│   ├── nat/               # NAT state management
│   │   └── table.go       # Session tracking
│   ├── tun/               # TUN interface handling
│   │   └── bridge.go      # Bridge orchestration
│   ├── api/               # REST API server
│   │   └── server.go      # HTTP endpoints
│   ├── config/            # Configuration
│   ├── logger/            # Logging utilities
│   └── ...
├── main.go                # Entry point
├── bridgeconfig.yaml      # Configuration file
└── README.md
```

### Building from Source

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o bridge .

# Run with race detector
go run -race main.go start
```

## Security Considerations

- **Root privileges required** for TUN interface creation
- Consider running in isolated network namespace
- Validate packet headers before translation
- Log all translation failures for debugging
- Implement rate limiting for production use

## Future Enhancements

- ICMPv4/ICMPv6 translation support
- Static port mapping configuration
- Web-based dashboard UI
- Prometheus metrics exporter
- Docker Compose integration
- IPv6 prefix delegation
- Performance benchmarking suite
- Windows native support (without WSL)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- TUN interface support via [water](https://github.com/songgao/water)
- IP packet handling with [golang.org/x/net](https://pkg.go.dev/golang.org/x/net)

## Contact

**Author**: mdxabu  
**Repository**: [github.com/mdxabu/bridge](https://github.com/mdxabu/bridge)

---

Made for the IPv6 transition