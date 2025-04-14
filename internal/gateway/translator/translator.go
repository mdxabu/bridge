package translator

import (
	"fmt"
	"net"
	"strings"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/logger"
)

func IsNAT64Address(ip string) bool {
	cfg, err := config.ParseConfig()
	if err != nil {
		return false
	}

	nat64Prefix, err := cfg.GetNAT64Prefix()
	if err != nil {
		logger.Error("Failed to get NAT64 IP: %v", err)
		return false
	}

	if len(ip) < len(nat64Prefix) {
		return false
	}

	if ip[:len(nat64Prefix)] != nat64Prefix {
		return false
	}

	return true
}

func GetIPV4fromNAT64(nat64 string) (string, error) {
	ip := net.ParseIP(nat64)
	if ip == nil {
		return "", fmt.Errorf("invalid IPv6 address: %s", nat64)
	}

	ip = ip.To16()
	if ip == nil || len(ip) != 16 {
		return "", fmt.Errorf("invalid IPv6 address format")
	}

	if !strings.HasPrefix(nat64, "64:ff9b::") && !(ip[0] == 0x64 && ip[1] == 0xff && ip[2] == 0x9b && ip[3] == 0x00) {
		return "", fmt.Errorf("address is not in NAT64 prefix")
	}

	ipv4 := net.IPv4(ip[12], ip[13], ip[14], ip[15])
	return ipv4.String(), nil
}
