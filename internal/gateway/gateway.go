package gateway

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/docker/client"
	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/gateway/forwarder"
	"github.com/mdxabu/bridge/internal/gateway/translator"
	"github.com/mdxabu/bridge/internal/logger"
	"github.com/mdxabu/bridge/internal/utils"
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
	packetCount := 0
	for {
		n, addr, err := ipv6Listener.ReadFrom(packetBuffer)
		if err != nil {
			logger.Error("Error reading IPv6 packet: %v", err)
			continue
		}

		packetCount++
		logger.Info("Packet #%d received from %s (IPv6)", packetCount, addr.String())
		displayPacketInfo("Received IPv6", addr.String(), packetBuffer[:n])

		translatedPacket, err := translator.TranslateIPv6ToIPv4(packetBuffer[:n], cfg)
		if err != nil {
			logger.Error("Failed to translate IPv6 to IPv4: %v", err)
			continue
		}

		displayPacketInfo("Translated IPv4", addr.String(), translatedPacket)

		err = forwarder.ForwardIPv4Packet(translatedPacket, cfg)
		if err != nil {
			logger.Error("Failed to forward IPv4 packet: %v", err)
			continue
		}

		logger.Info("Forwarded IPv4 packet #%d", packetCount)
	}
}

func displayPacketInfo(label, addr string, packet []byte) {
	srcAddr, destAddr, err := utils.ExtractAddresses(packet)
	if err != nil {
		logger.Error("Failed to extract addresses from packet: %v", err)
		return
	}

	logger.Info("%s packet from %s\n  Source Address: %s\n  Destination Address: %s\n  Packet Data: %v",
		label, addr, srcAddr, destAddr, packet[:40])
}

func getContainerIP(containerID string) (string, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.41"))
	if err != nil {
		return "", fmt.Errorf("failed to create Docker client: %v", err)
	}
	ctx := context.Background()
	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %v", err)
	}

	ipv4Address := containerJSON.NetworkSettings.IPAddress
	if ipv4Address == "" {
		return "", fmt.Errorf("no IPv4 address found for container %s", containerID)
	}

	return ipv4Address, nil
}
