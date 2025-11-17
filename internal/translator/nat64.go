package translator

import (
	"fmt"
	"net"
	"strings"
)

// IsNAT64Address checks if an IP address is in the NAT64 prefix range
func IsNAT64Address(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	ipv6 := parsedIP.To16()
	if ipv6 == nil {
		return false
	}

	// Check for well-known NAT64 prefix 64:ff9b::/96
	return ipv6[0] == 0x00 && ipv6[1] == 0x64 &&
		ipv6[2] == 0xff && ipv6[3] == 0x9b &&
		ipv6[4] == 0x00 && ipv6[5] == 0x00 &&
		ipv6[6] == 0x00 && ipv6[7] == 0x00 &&
		ipv6[8] == 0x00 && ipv6[9] == 0x00 &&
		ipv6[10] == 0x00 && ipv6[11] == 0x00
}

// GetIPV4fromNAT64 extracts the embedded IPv4 address from a NAT64 IPv6 address
func GetIPV4fromNAT64(nat64 string) (string, error) {
	ip := net.ParseIP(nat64)
	if ip == nil {
		return "", fmt.Errorf("invalid IPv6 address: %s", nat64)
	}

	ip = ip.To16()
	if ip == nil || len(ip) != 16 {
		return "", fmt.Errorf("invalid IPv6 address format")
	}

	if !strings.HasPrefix(nat64, "64:ff9b::") && !(ip[0] == 0x00 && ip[1] == 0x64 && ip[2] == 0xff && ip[3] == 0x9b) {
		return "", fmt.Errorf("address is not in NAT64 prefix")
	}

	ipv4 := net.IPv4(ip[12], ip[13], ip[14], ip[15])
	return ipv4.String(), nil
}
