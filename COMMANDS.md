`bridge init` - for creating configuration file 
`bridge nat64` - It will start pinging the ipv4 datasets from out ipv6.
`bridge dns` - It will fetch the IPv4 and IPv6 from the domain name, if ipv6 is not present, it will synthesize the ipv6 from the ipv4 by nat64 prefix.
`bridge metrics --nat64` - It will visualize the nat64 live results with graph view and can export the JSON.