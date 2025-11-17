package tun

import (
	"fmt"
	"io"
	"net"

	"github.com/mdxabu/bridge/internal/logger"
	"github.com/mdxabu/bridge/internal/nat"
	"github.com/mdxabu/bridge/internal/translator"
	"github.com/songgao/water"
)

// Bridge represents the NAT64 bridge
type Bridge struct {
	tunIPv6     *water.Interface
	tunIPv4     *water.Interface
	natTable    *nat.NATTable
	nat64Prefix string
	running     bool
	packetChan  chan []byte
}

// NewBridge creates a new NAT64 bridge
func NewBridge(nat64Prefix string) (*Bridge, error) {
	return &Bridge{
		natTable:    nat.NewNATTable(),
		nat64Prefix: nat64Prefix,
		packetChan:  make(chan []byte, 1000),
	}, nil
}

// CreateTUNInterface creates a TUN interface
func (b *Bridge) CreateTUNInterface(name string, isIPv6 bool) error {
	config := water.Config{
		DeviceType: water.TUN,
	}

	// On macOS, TUN interface names must be utun[0-9]+
	// Let the system auto-assign the name by not setting it
	// The water library will handle platform-specific naming

	iface, err := water.New(config)
	if err != nil {
		return fmt.Errorf("failed to create TUN interface: %w", err)
	}

	if isIPv6 {
		b.tunIPv6 = iface
		logger.Success("Created IPv6 TUN interface: %s", iface.Name())
	} else {
		b.tunIPv4 = iface
		logger.Success("Created IPv4 TUN interface: %s", iface.Name())
	}

	return nil
}

// ConfigureInterface configures the TUN interface with IP address
func ConfigureInterface(ifaceName string, ipAddr string, isIPv6 bool) error {
	// Note: This requires system commands and privileges
	// In production, this would use syscalls or exec commands
	logger.Info("Interface %s should be configured with %s", ifaceName, ipAddr)
	return nil
}

// Start starts the NAT64 bridge
func (b *Bridge) Start() error {
	if b.tunIPv6 == nil || b.tunIPv4 == nil {
		return fmt.Errorf("TUN interfaces not created")
	}

	b.running = true
	b.natTable.StartCleanupRoutine()

	// Start packet processing goroutines
	go b.readIPv6Packets()
	go b.readIPv4Packets()
	go b.processPackets()

	logger.Success("NAT64 Bridge started successfully")
	return nil
}

// Stop stops the NAT64 bridge
func (b *Bridge) Stop() error {
	b.running = false

	if b.tunIPv6 != nil {
		b.tunIPv6.Close()
	}

	if b.tunIPv4 != nil {
		b.tunIPv4.Close()
	}

	close(b.packetChan)
	logger.Info("NAT64 Bridge stopped")
	return nil
}

// readIPv6Packets reads packets from the IPv6 TUN interface
func (b *Bridge) readIPv6Packets() {
	buffer := make([]byte, 2000)

	for b.running {
		n, err := b.tunIPv6.Read(buffer)
		if err != nil {
			if err != io.EOF && b.running {
				logger.Error("Error reading from IPv6 TUN: %v", err)
			}
			continue
		}

		// Create a copy of the packet
		packet := make([]byte, n)
		copy(packet, buffer[:n])

		// Send to processing channel
		select {
		case b.packetChan <- packet:
		default:
			logger.Warn("Packet channel full, dropping IPv6 packet")
		}
	}
}

// readIPv4Packets reads packets from the IPv4 TUN interface
func (b *Bridge) readIPv4Packets() {
	buffer := make([]byte, 2000)

	for b.running {
		n, err := b.tunIPv4.Read(buffer)
		if err != nil {
			if err != io.EOF && b.running {
				logger.Error("Error reading from IPv4 TUN: %v", err)
			}
			continue
		}

		// Create a copy of the packet
		packet := make([]byte, n)
		copy(packet, buffer[:n])

		// Process IPv4 to IPv6 translation
		go b.translateIPv4ToIPv6(packet)
	}
}

// processPackets processes packets from the channel
func (b *Bridge) processPackets() {
	for packet := range b.packetChan {
		go b.translateIPv6ToIPv4(packet)
	}
}

// translateIPv6ToIPv4 translates and forwards IPv6 packets to IPv4
func (b *Bridge) translateIPv6ToIPv4(data []byte) {
	// Parse IPv6 packet
	pkt, err := translator.ParseIPv6Packet(data)
	if err != nil {
		logger.Error("Failed to parse IPv6 packet: %v", err)
		return
	}

	// Check if destination is NAT64 address
	if !translator.IsNAT64Address(pkt.DstIP.String()) {
		logger.Debug("Packet destination is not NAT64 address: %s", pkt.DstIP)
		return
	}

	// Extract IPv4 destination
	ipv4Dst, err := translator.GetIPV4fromNAT64(pkt.DstIP.String())
	if err != nil {
		logger.Error("Failed to extract IPv4 from NAT64: %v", err)
		return
	}

	ipv4DstIP := net.ParseIP(ipv4Dst).To4()
	if ipv4DstIP == nil {
		logger.Error("Invalid IPv4 destination: %s", ipv4Dst)
		return
	}

	// Create or lookup NAT session
	session, err := b.natTable.CreateSession(
		pkt.Protocol,
		pkt.SrcIP,
		pkt.SrcPort,
		pkt.DstIP,
		pkt.DstPort,
		ipv4DstIP,
	)
	if err != nil {
		logger.Error("Failed to create NAT session: %v", err)
		return
	}

	// Translate packet
	ipv4Packet, err := translator.TranslateIPv6ToIPv4(pkt, b.nat64Prefix)
	if err != nil {
		logger.Error("Failed to translate packet: %v", err)
		return
	}

	// Update session statistics
	b.natTable.UpdateSession(session.ID, uint64(len(ipv4Packet)), "outbound")

	// Write to IPv4 TUN interface
	_, err = b.tunIPv4.Write(ipv4Packet)
	if err != nil {
		logger.Error("Failed to write to IPv4 TUN: %v", err)
		return
	}

	logger.Debug("Translated IPv6->IPv4: %s", pkt.String())
}

// translateIPv4ToIPv6 translates and forwards IPv4 packets to IPv6
func (b *Bridge) translateIPv4ToIPv6(data []byte) {
	// Parse IPv4 packet
	pkt, err := translator.ParseIPv4Packet(data)
	if err != nil {
		logger.Error("Failed to parse IPv4 packet: %v", err)
		return
	}

	// Lookup NAT session (reverse direction)
	session, found := b.natTable.LookupSessionIPv4toIPv6(pkt.Protocol, pkt.DstPort)
	if !found {
		logger.Debug("No NAT session found for IPv4 packet: %s", pkt.String())
		return
	}

	// Translate packet
	ipv6Packet, err := translator.TranslateIPv4ToIPv6(pkt, b.nat64Prefix)
	if err != nil {
		logger.Error("Failed to translate packet: %v", err)
		return
	}

	// Update session statistics
	b.natTable.UpdateSession(session.ID, uint64(len(ipv6Packet)), "inbound")

	// Write to IPv6 TUN interface
	_, err = b.tunIPv6.Write(ipv6Packet)
	if err != nil {
		logger.Error("Failed to write to IPv6 TUN: %v", err)
		return
	}

	logger.Debug("Translated IPv4->IPv6: %s", pkt.String())
}

// GetStats returns bridge statistics
func (b *Bridge) GetStats() map[string]interface{} {
	return b.natTable.GetStats()
}

// GetActiveSessions returns all active NAT sessions
func (b *Bridge) GetActiveSessions() []*nat.SessionState {
	return b.natTable.GetAllSessions()
}
