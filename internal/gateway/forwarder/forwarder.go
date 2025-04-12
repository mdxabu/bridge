package forwarder

import (
	"fmt"
	"net"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/gateway/translator"
	"github.com/mdxabu/bridge/internal/logger"
)

func ForwardIPv6(cfg *config.Config) error {
	ipv6Listener, err := net.ListenPacket("ip6:ipv6", cfg.GetInterface())
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

		err = ForwardIPv4Packet(translatedPacket, cfg)
		if err != nil {
			logger.Error("Failed to forward IPv4 packet: %v", err)
			continue
		}

		logger.Debug("Forwarded translated IPv4 packet")
	}
}

func ForwardIPv4Packet(packet []byte, cfg *config.Config) error {
	ipv4Addr := cfg.GetIPv4ContainerID()

	conn, err := net.Dial("udp", ipv4Addr)
	if err != nil {
		return fmt.Errorf("failed to connect to IPv4 container: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("failed to send IPv4 packet: %v", err)
	}

	return nil
}
