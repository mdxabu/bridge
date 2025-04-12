package gateway

import (
	"fmt"
	"net"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/gateway/forwarder"
	"github.com/mdxabu/bridge/internal/gateway/translator"
	"github.com/mdxabu/bridge/internal/logger"
)

func Start(cfg *config.Config) {
	err := listenIPv6(cfg)
	if err != nil {
		logger.Error("Gateway failed to run: %v", err)
	}
}

func listenIPv6(cfg *config.Config) error {
	ipv6Listener, err := net.ListenPacket("ip6:ipv6", "::")
	if err != nil {
		return fmt.Errorf("failed to listen on IPv6 interface: %v", err)
	}
	defer ipv6Listener.Close()

	logger.Info("Listening for IPv6 packets...")

	packetBuffer := make([]byte, 4096)
	for {
		n, addr, err := ipv6Listener.ReadFrom(packetBuffer)
		if err != nil {
			logger.Error("Error reading IPv6 packet: %v", err)
			continue
		}

		logger.Debug("Received IPv6 packet from %s", addr.String())

		translatedPacket, err := translator.TranslateIPv6ToIPv4(packetBuffer[:n], cfg)
		if err != nil {
			logger.Error("Failed to translate IPv6 to IPv4: %v", err)
			continue
		}

		err = forwarder.ForwardIPv4Packet(translatedPacket, cfg)
		if err != nil {
			logger.Error("Failed to forward IPv4 packet: %v", err)
			continue
		}

		logger.Debug("Forwarded translated IPv4 packet")
	}
}
