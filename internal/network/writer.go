package network

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/mdxabu/bridge/internal/logger"
)

type Writer struct {
	interfaceName string
	handle        *pcap.Handle
	logger        *logger.Logger
}

func NewWriter(ifaceName string, logger *logger.Logger) (*Writer, error) {
	handle, err := pcap.OpenLive(ifaceName, 65536, false, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("failed to open device %s for writing: %w", ifaceName, err)
	}

	logger.Info("Writing packets on interface: %s", ifaceName)

	return &Writer{
		interfaceName: ifaceName,
		handle:        handle,
		logger:        logger,
	}, nil
}

func (w *Writer) WritePacket(packet gopacket.Packet) error {
	rawBytes := packet.Data()

	err := w.handle.WritePacketData(rawBytes)
	if err != nil {
		return fmt.Errorf("error sending packet on %s: %w", w.interfaceName, err)
	}
	w.logger.Debug("Packet sent successfully")
	return nil
}
