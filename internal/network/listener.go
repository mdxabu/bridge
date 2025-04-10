package network

import (
	"fmt"
	"log/slog"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// Listener handles capturing network packets using libpcap/Npcap.
type Listener struct {
	interfaceName string
	handle        *pcap.Handle
	logger        *slog.Logger
}

// NewListener creates a new Listener instance.
func NewListener(ifaceName string, logger *slog.Logger) (*Listener, error) {
	// Open the device for capturing
	handle, err := pcap.OpenLive(ifaceName, 65536, false, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("failed to open device %s: %w", ifaceName, err)
	}

	logger.Info("Listening on interface:", slog.String("interface", ifaceName))

	return &Listener{
		interfaceName: ifaceName,
		handle:        handle,
		logger:        logger,
	}, nil
}

// Start starts listening for packets on the specified interface.
// It returns a channel of gopacket.Packet.
func (l *Listener) Start() *gopacket.PacketSource {
	l.logger.Info("Starting packet capture")
	packetSource := gopacket.NewPacketSource(l.handle, l.handle.LinkType())
	return packetSource
}

// Stop closes the packet capture handle.
func (l *Listener) Stop() error {
	l.logger.Info("Stopping packet capture")
	if l.handle != nil {
		l.handle.Close()
	}
	return nil
}