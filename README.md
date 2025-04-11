# Bridge

Bridge is a NAT64 gateway implementation that enables communication between IPv6-only networks and IPv4-only resources. It translates IPv6 packets to IPv4 and vice versa, allowing IPv6 clients to access IPv4 services without requiring dual-stack support. This lightweight, Go-based solution offers efficient packet handling, customizable configurations, and easy deployment in various network environments. Bridge aims to ease the transition to IPv6 while maintaining compatibility with the existing IPv4 infrastructure.

## Running the Project

To run Bridge, follow these steps:

1. Ensure you have Go installed (version 1.19 or later)
2. Clone the repository
   ```
   git clone https://github.com/mdxabu/bridge.git
   cd bridge
   ```
3. Build the project
   ```
   > go build -o bridge
   > go install .

   ```

   ```


The application requires elevated privileges to access network interfaces for packet capture and transmission.