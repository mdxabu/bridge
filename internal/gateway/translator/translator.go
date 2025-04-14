package translator

import (
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
