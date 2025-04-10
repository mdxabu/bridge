package gateway

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/core"
	"github.com/mdxabu/bridge/internal/core/state"
	"github.com/mdxabu/bridge/internal/logger"
	"github.com/mdxabu/bridge/internal/network"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Gateway struct {
	config             config.Config
	log                *logger.Logger
	stateTable         *state.Table
	listener           *network.Listener
	writer             *network.Writer
	nat64Prefix        net.IP
	externalIPv4s      []net.IP
	nextAvailablePort  uint16
	portAllocatorMutex sync.Mutex
}

func New(cfg config.Config, log *logger.Logger) (*Gateway, error) {
	nat64PrefixStr := strings.TrimSuffix(cfg.NAT64Prefix, "/96")
	nat64Prefix := net.ParseIP(nat64PrefixStr)
	if nat64Prefix == nil || nat64Prefix.To16() == nil || len(nat64Prefix.To16()) != net.IPv6len-12 {
		return nil, fmt.Errorf("invalid NAT64 prefix: %s", cfg.NAT64Prefix)
	}

	stateTable := state.NewTable(cfg.StateTimeout)
	stateTable.SetLogger(log)

	listener, err := network.NewListener(cfg.InterfaceName, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	writer, err := network.NewWriter(cfg.InterfaceName, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer: %w", err)
	}

	externalIPv4 := net.ParseIP("192.168.1.100")
	if externalIPv4 == nil || externalIPv4.To4() == nil {
		log.Warn("Using a placeholder external IPv4 address. Please configure appropriately.")
		externalIPv4 = net.IPv4(192, 168, 1, 100)
	}

	return &Gateway{
		config:            cfg,
		log:               log,
		stateTable:        stateTable,
		listener:          listener,
		writer:            writer,
		nat64Prefix:       nat64Prefix,
		externalIPv4s:     []net.IP{externalIPv4},
		nextAvailablePort: 1024,
	}, nil
}

func (g *Gateway) Start() error {
	g.log.Info("Starting NAT64 gateway...")

	packetSource := g.listener.Start()
	if packetSource == nil {
		return fmt.Errorf("failed to start packet listener")
	}

	go g.stateTable.StartCleanupRoutine()

	for packet := range packetSource.Packets() {
		g.handlePacket(packet)
	}

	return nil
}

func (g *Gateway) Stop() error {
	g.log.Info("Stopping NAT64 gateway...")
	if err := g.listener.Stop(); err != nil {
		g.log.Error("Error stopping listener: %v", err)
	}
	return nil
}

func (g *Gateway) handlePacket(packet gopacket.Packet) {
	if err := packet.ErrorLayer(); err != nil {
		g.log.Error("Error decoding packet: %v", err)
		return
	}

	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	ipv6Layer := packet.Layer(layers.LayerTypeIPv6)

	switch {
	case ipv6Layer != nil && ipv4Layer == nil:
		g.log.Debug("Received IPv6 packet")
		g.processIPv6Packet(packet)
	case ipv4Layer != nil && ipv6Layer == nil:
		g.log.Debug("Received IPv4 packet")
		g.processIPv4Packet(packet)
	default:
		g.log.Debug("Received non-IPv4/IPv6 packet or both present")
	}
}

func (g *Gateway) processIPv6Packet(packet gopacket.Packet) {
	ipv6 := packet.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
	transportLayer := packet.TransportLayer()
	if transportLayer == nil {
		g.log.Debug("IPv6 packet without transport layer, skipping.")
		return
	}
	srcPort, dstPort := core.GetTransportPorts(transportLayer)

	if core.IsIPv6InNAT64Prefix(ipv6.DstIP, g.nat64Prefix) {
		embeddedIPv4 := core.ExtractIPv4FromIPv6(ipv6.DstIP, g.nat64Prefix)
		if embeddedIPv4 != nil {
			g.log.Debug("Translating IPv6 to IPv4: %s:%d -> %s:%d", ipv6.SrcIP, srcPort, embeddedIPv4, dstPort)
			translatedPacket, err := core.TranslateIPv6ToIPv4(packet, embeddedIPv4)
			if err != nil {
				g.log.Error("Error translating IPv6 to IPv4: %v", err)
				return
			}
			g.writer.WritePacket(translatedPacket)
		} else {
			g.log.Warn("Destination IPv6 does not contain a valid embedded IPv4 address: %s", ipv6.DstIP)
		}
	} else {
		entry := g.stateTable.LookupIPv6ToIPv4(ipv6.SrcIP, srcPort, ipv6.DstIP, dstPort)
		if entry != nil {
			g.log.Debug("State entry found for outgoing IPv6: %s:%d -> %s:%d (via %s:%d)", ipv6.SrcIP, srcPort, ipv6.DstIP, dstPort, entry.IPv4SrcIP, entry.IPv4SrcPort)
		} else {
			if len(g.externalIPv4s) > 0 {
				externalIPv4 := g.externalIPv4s[0]
				externalPort := g.allocatePort()
				g.log.Debug("Creating new state for outgoing IPv6: %s:%d -> %s:%d, mapping to %s:%d", ipv6.SrcIP, srcPort, ipv6.DstIP, dstPort, externalIPv4, externalPort)
				g.stateTable.CreateEntry(ipv6.SrcIP, srcPort, ipv6.DstIP, dstPort, externalIPv4, externalPort)
				g.log.Warn("Translation of direct IPv6 to arbitrary IPv4 is not fully implemented. Requires DNS64 or similar.")
			} else {
				g.log.Error("No external IPv4 address available for NAT.")
			}
		}
	}
}

func (g *Gateway) processIPv4Packet(packet gopacket.Packet) {
	ipv4 := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	transportLayer := packet.TransportLayer()
	if transportLayer == nil {
		g.log.Debug("IPv4 packet without transport layer, skipping.")
		return
	}
	srcPort, dstPort := core.GetTransportPorts(transportLayer)

	for _, externalIP := range g.externalIPv4s {
		if ipv4.DstIP.Equal(externalIP) {
			entry := g.stateTable.LookupIPv4ToIPv6(ipv4.DstIP, uint16(dstPort), ipv4.SrcIP, uint16(srcPort))
			if entry != nil {
				g.log.Debug("Found state entry for IPv4 to IPv6: %s:%d -> %s:%d", ipv4.SrcIP, srcPort, entry.IPv6SrcIP, entry.IPv6SrcPort)
				translatedPacket, err := core.TranslateIPv4ToIPv6(packet, entry.IPv6SrcIP)
				if err != nil {
					g.log.Error("Error translating IPv4 to IPv6: %v", err)
					return
				}
				g.writer.WritePacket(translatedPacket)
			} else {
				g.log.Debug("No state entry found for incoming IPv4: %s:%d -> %s:%d", ipv4.SrcIP, srcPort, ipv4.DstIP, dstPort)
			}
			return
		}
	}

	g.log.Debug("IPv4 packet to non-gateway destination: %s -> %s", ipv4.SrcIP, ipv4.DstIP)
}

func (g *Gateway) allocatePort() uint16 {
	g.portAllocatorMutex.Lock()
	defer g.portAllocatorMutex.Unlock()
	port := g.nextAvailablePort
	g.nextAvailablePort++
	if g.nextAvailablePort < 1024 {
		g.nextAvailablePort = 1024
	}
	return port
}
