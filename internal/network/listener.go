package network

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/mdxabu/bridge/internal/logger"
)

type Listener struct {
	interfaceName string
	handle        *pcap.Handle
	logger        *logger.Logger
}

func NewListener(ifaceName string, logger *logger.Logger) (*Listener, error) {
	handle, err := pcap.OpenLive(ifaceName, 65536, false, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("failed to open device %s: %w", ifaceName, err)
	}

	logger.Info("Listening on interface: %s", ifaceName)

	return &Listener{
		interfaceName: ifaceName,
		handle:        handle,
		logger:        logger,
	}, nil
}

func (l *Listener) Start() *gopacket.PacketSource {
	l.logger.Info("Starting packet capture")
	packetSource := gopacket.NewPacketSource(l.handle, l.handle.LinkType())
	return packetSource
}

func (l *Listener) Stop() error {
	l.logger.Info("Stopping packet capture")
	if l.handle != nil {
		l.handle.Close()
	}
	return nil
}
