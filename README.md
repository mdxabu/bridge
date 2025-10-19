# Bridge


```
+-----------+       IPv6          +------------------+       IPv4           +-----------+
| IPv6-only | --->  TCP/UDP  -->  | NAT64 Translator | --->  TCP/UDP  -->   | IPv4-only |
|  Client   | <---  TCP/UDP  ---  |   (Go process)   | <---  TCP/UDP  ---   |  Server   |
+-----------+                     +------------------+                      +-----------+
                              converts IPv6<->IPv4 packets
```