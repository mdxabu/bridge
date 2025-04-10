package gateway

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/core"
	"github.com/mdxabu/bridge/internal/core/state"
	"github.com/mdxabu/bridge/internal/logger" // Corrected import path
	"github.com/mdxabu/bridge/internal/network"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Gateway represents the NAT64 gateway.
type Gateway struct {
	config        config.Config
	log           *logger.Logger
	stateTable    *state.Table
	listener      *network.Listener
	writer        *network.Writer
	nat64Prefix   net.IP
	externalIPv4s []net.IP // For outgoing IPv6->IPv4 connections
	nextAvailablePort uint16
	portAllocatorMutex sync.Mutex
}

// New creates a new Gateway instance.
func New(cfg config.Config, log *logger.Logger) (*Gateway, error) {
	nat64PrefixStr := strings.TrimSuffix(cfg.NAT64Prefix, "/96")
	nat64Prefix := net.ParseIP(nat64PrefixStr)
	if nat64Prefix == nil || nat64Prefix.To16() == nil || len(nat64Prefix.To16()) != net.IPv6len-12 {
		return nil, fmt.Errorf("invalid NAT64 prefix: %s", cfg.NAT64Prefix)
	}

	stateTable := state.NewTable(cfg.StateTimeout)
	stateTable.SetLogger(log) // Set logger for state table

	listener, err := network.NewListener(cfg.InterfaceName, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	writer, err := network.NewWriter(cfg.InterfaceName, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer: %w", err)
	}

	// For simplicity, let's assume the gateway has one external IPv4 address for now.
	// In a real-world scenario, you might need to configure a pool of addresses.
	// For testing in Docker, you might need to use the IPv4 address of the bridge interface.
	externalIPv4 := net.ParseIP("192.168.1.100") // Replace with your gateway's IPv4
	if externalIPv4 == nil || externalIPv4.To4() == nil {
		log.Warn("Using a placeholder external IPv4 address. Please configure appropriately.")
		externalIPv4 = net.IPv4(192, 168, 1, 100)
	}

	return &Gateway{
		config:        cfg,
		log:           log,
		stateTable:    stateTable,
		listener:      listener,
		writer:        writer,
		nat64Prefix:   nat64Prefix,
		externalIPv4s: []net.IP{externalIPv4},
		nextAvailablePort: 1024, // Starting port for NAT
	}, nil
}

// Start starts the NAT64 gateway.
func (g *Gateway) Start() error {
	g.log.Info("Starting NAT64 gateway...")

	packetSource := g.listener.Start()
	if packetSource == nil {
		return fmt.Errorf("failed to start packet listener")
	}

	go g.stateTable.StartCleanupRoutine() // Start background task to remove expired entries

	for packet := range packetSource.Packets() {
		g.handlePacket(packet)
	}

	return nil
}

// Stop stops the NAT64 gateway.
func (g *Gateway) Stop() error {
	g.log.Info("Stopping NAT64 gateway...")
	if err := g.listener.Stop(); err != nil {
		g.log.Errorf("Error stopping listener: %v", err)
	}
	return nil
}

func (g *Gateway) handlePacket(packet gopacket.Packet) {
	if err := packet.ErrorLayer(); err != nil {
		g.log.Errorf("Error decoding packet: %v", err)
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
		// Ignore non-IPv4/IPv6 packets or packets with both layers
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
		// Outgoing IPv6 to IPv4
		embeddedIPv4 := core.ExtractIPv4FromIPv6(ipv6.DstIP, g.nat64Prefix)
		if embeddedIPv4 != nil {
			g.log.Debugf("Translating IPv6 to IPv4: %s:%d -> %s:%d", ipv6.SrcIP, srcPort, embeddedIPv4, dstPort)
			translatedPacket, err := core.TranslateIPv6ToIPv4(packet, embeddedIPv4)
			if err != nil {
				g.log.Errorf("Error translating IPv6 to IPv4: %v", err)
				return
			}
			g.writer.WritePacket(translatedPacket)
		} else {
			g.log.Warnf("Destination IPv6 does not contain a valid embedded IPv4 address: %s", ipv6.DstIP)
		}
	} else {
		// For outgoing IPv6 traffic not destined for NAT64 prefix, we need to establish a state.
		// Check if there's an existing state.
		entry := g.stateTable.LookupIPv6ToIPv4(ipv6.SrcIP, srcPort, ipv6.DstIP, dstPort)
		if entry != nil {
			// State exists, but this logic needs to be refined based on the flow.
			g.log.Debugf("State entry found for outgoing IPv6: %s:%d -> %s:%d (via %s:%d)", ipv6.SrcIP, srcPort, ipv6.DstIP, dstPort, entry.IPv4SrcIP, entry.IPv4SrcPort)
			// This might be a reply to an earlier IPv4->IPv6 translated packet.
			// The destination here is IPv6, so no direct translation needed here.
		} else {
			// New outgoing IPv6 connection to an IPv4 server (not using NAT64 prefix directly).
			// We need to allocate an external IPv4 address and port for this.
			if len(g.externalIPv4s) > 0 {
				externalIPv4 := g.externalIPv4s[0]
				externalPort := g.allocatePort()
				g.log.Debugf("Creating new state for outgoing IPv6: %s:%d -> %s:%d, mapping to %s:%d", ipv6.SrcIP, srcPort, ipv6.DstIP, dstPort, externalIPv4, externalPort)
				g.stateTable.CreateEntry(ipv6.SrcIP, srcPort, ipv6.DstIP, dstPort, externalIPv4, externalPort)

				// Now, we need to translate this IPv6 packet to IPv4 with the allocated source IP and port.
				// The destination IPv4 would be the original IPv6 destination (not in NAT64 prefix).
				// This requires a mechanism to resolve the IPv6 destination to an IPv4 address (e.g., DNS64).
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

	// Check if the destination IPv4 is one of the gateway's external addresses (for incoming traffic)
	for _, externalIP := range g.externalIPv4s {
		if ipv4.DstIP.Equal(externalIP) {
			// Incoming IPv4 to IPv6
			// Lookup state table
			entry := g.stateTable.LookupIPv4ToIPv6(ipv4.DstIP, uint16(dstPort), ipv4.SrcIP, uint16(srcPort))
			if entry != nil {
				g.log.Debugf("Found state entry for IPv4 to IPv6: %s:%d -> %s:%d", ipv4.SrcIP, srcPort, entry.IPv6SrcIP, entry.IPv6SrcPort)
				translatedPacket, err := core.TranslateIPv4ToIPv6(packet, entry.IPv6SrcIP)
				if err != nil {
					g.log.Errorf("Error translating IPv4 to IPv6: %v", err)
					return
				}
				g.writer.WritePacket(translatedPacket)
			} else {
				g.log.Debugf("No state entry found for incoming IPv4: %s:%d -> %s:%d", ipv4.SrcIP, srcPort, ipv4.DstIP, dstPort)
				// If no state, it might be an unsolicited incoming packet, which we might need to drop or handle differently.
			}
			return
		}
	}

	g.log.Debugf("IPv4 packet to non-gateway destination: %s -> %s", ipv4.SrcIP, ipv4.DstIP)
	// Handle other IPv4 traffic if needed.
}

// allocatePort allocates a unique source port for outgoing IPv6 connections.
func (g *Gateway) allocatePort() uint16 {
	g.portAllocatorMutex.Lock()
	defer g.portAllocatorMutex.Unlock()
	port := g.nextAvailablePort
	g.nextAvailablePort++
	if g.nextAvailablePort < 1024 { // Avoid reserved ports
		g.nextAvailablePort = 1024
	}
	return port
}