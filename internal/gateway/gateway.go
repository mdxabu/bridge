package gateway

import (
	"fmt"
	"net"

	"github.com/mdxabu/bridge/internal/config"
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

	fmt.Println("IPv6 Address: ", ipv6Addr)
	fmt.Println("is NAT64 Address: ", translator.IsNAT64Address(ipv6Addr.String()))

	logger.Info("Successfully translated IPv6 to IPv4")
}
