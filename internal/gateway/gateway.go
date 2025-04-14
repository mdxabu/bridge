package gateway

import (
	"net"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/gateway/forwarder"
	"github.com/mdxabu/bridge/internal/gateway/translator"
	"github.com/mdxabu/bridge/internal/logger"
)

func Start(ip string) {
	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Error("Failed to parse configuration: %v", err)
		return
	}

	nat64ipv6, err := cfg.GetNAT64IP()
	if err != nil {
		logger.Error("Failed to get NAT64 IP: %v", err)
		return
	}

	ipv6Addr := net.ParseIP(nat64ipv6)
	if ipv6Addr == nil {
		logger.Error("Failed to parse NAT64 IP: %s", nat64ipv6)
		return
	}

	logger.Info("is NAT64Address: %v", translator.IsNAT64Address(ip))

	logger.Success("Successfully translated IPv6 to IPv4")

	ipv4addr, err := translator.GetIPV4fromNAT64(ipv6Addr.String())
	if err != nil {
		logger.Error("Failed to get IPv4 from NAT64: %v", err)
		return
	}
	logger.Info("IPv4 Address: %s", ipv4addr)

	// Call ReadIPv4Addresses with correct capitalization
	ipv4Addresses, err := forwarder.ReadIPv4Addresses()
	if err != nil {
		logger.Error("Failed to read IPv4 addresses: %v", err)
		return
	}

	logger.Info("Found %d IPv4 addresses", len(ipv4Addresses))
}
