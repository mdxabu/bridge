# Bridge

Bridge is a NAT64 gateway implementation that enables communication between IPv6-only networks and IPv4-only resources. It translates IPv6 packets to IPv4 and vice versa, allowing IPv6 clients to access IPv4 services without requiring dual-stack support. This lightweight, Go-based solution offers efficient packet handling, customizable configurations, and easy deployment in various network environments. Bridge aims to ease the transition to IPv6 while maintaining compatibility with the existing IPv4 infrastructure.

## Running the Project

To run Bridge, follow these steps:

1. Ensure you have Go installed (version 1.19 or later)
   The latest version of GoLang can be installed from here.
   ```
   >https://go.dev/doc/install
   ```
3. Clone the repository
   ```
   git clone https://github.com/mdxabu/bridge.git
   cd bridge
   ```
4. Build the project
   ```
   > go build -o bridge
   > go install 
   ```
go: downloading github.com/fatih/color v1.18.0
go: downloading github.com/mattn/go-colorable v0.1.13
go: downloading github.com/mattn/go-isatty v0.0.20
The above lines of message will be displayed when ```go build -o bridge``` is executed

5. Bride Initialization
   ```
   > bridge init
   ```
   This initializes the Bridge CLI
6. Bridge run
   ```
   > bridge run
   ```
   This commands runs bridge CLI. The source IP, the destination IP, RTT, no. of sent and received packets, and results are listed in a Tabular coulumn.
7. Bridge dns
   ```
   > bridge dns
   ```
   By using bridge dns, the DNS64 resoluting starts
   
The application requires elevated privileges to access network interfaces for packet capture and transmission.
