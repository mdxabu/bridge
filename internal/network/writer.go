package network

import (
	"fmt"
	"log/slog"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// Writer handles sending network packets using libpcap/Npcap.
type Writer struct {
	interfaceName string
	handle        *pcap.Handle
	logger        *slog.Logger
}

// NewWriter creates a new Writer instance.
func NewWriter(ifaceName string, logger *slog.Logger) (*Writer, error) {
	// Open the device for sending as well (can be the same as listener)
	handle, err := pcap.OpenLive(ifaceName, 65536, false, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("failed to open device %s for writing: %w", ifaceName, err)
	}

	logger.Info("Writing packets on interface:", slog.String("interface", ifaceName))

	return &Writer{
		interfaceName: ifaceName,
		handle:        handle,
		logger:        logger,
	}, nil
}

// WritePacket sends a raw network packet.
func (w *Writer) WritePacket(packet gopacket.Packet) error {
	// Serialize the packet
	rawBytes := packet.Data()

	// Write the packet to the wire
	err := w.handle.WritePacketData(rawBytes)
	if err != nil {
		return fmt.Errorf("error sending packet on %s: %w", w.interfaceName, err)
	}
	w.logger.Debug("Packet sent successfully")
	return nil
}