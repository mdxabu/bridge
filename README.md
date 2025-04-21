# Bridge

Bridge is a NAT64 gateway implementation that enables communication between IPv6-only networks and IPv4-only resources. It translates IPv6 packets to IPv4 and vice versa, allowing IPv6 clients to access IPv4 services without requiring dual-stack support. This lightweight, Go-based solution offers efficient packet handling, customizable configurations, and easy deployment in various network environments. Bridge aims to ease the transition to IPv6 while maintaining compatibility with the existing IPv4 infrastructure.

## Running the Project

To run Bridge, follow these steps:

1. Ensure you have Go installed (version 1.19 or later)
   The latest version of GoLang can be installed from here.
   [https://go.dev/doc/install](https://go.dev/doc/install)
3. Clone the repository
   ```bash
   git clone https://github.com/mdxabu/bridge.git
   cd bridge
   ```
4. Build the project
   ```bash
   go build -o bridge
   go install 
   ```


5. Bridge Initialization
   ```bash
   bridge init
   ```
   This initializes the `bridgeconfig.yaml` file for Bridge CLI.
6. Bridge run
   ```bash
   bridge run
   ```
   This command runs the bridge CLI. The source IP, the destination IP, RTT, the number of sent and received packets, and results are listed in a tabular column.
7. Bridge DNS
   ```bash
   bridge dns
   ```
   By using bridge DNS, the DNS64 resolving starts
   
8. Bridge Metrics
```bash
bridge metrics --nat64
```
This command will show the live metrics of translation in the web dashboard.
