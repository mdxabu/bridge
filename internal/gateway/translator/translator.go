package translator

import (
	"errors"
	"net"

	"github.com/mdxabu/bridge/internal/config"
)

type IPv6Header struct {
	Version       uint8
	TrafficClass  uint8
	FlowLabel     uint32
	PayloadLength uint16
	NextHeader    uint8
	HopLimit      uint8
	Source        []byte
	Destination   []byte
}

func TranslateIPv6ToIPv4(pkt []byte, cfg *config.Config) ([]byte, error) {
	ipv6Header, ipv6Payload, err := parseIPv6Header(pkt)
	if err != nil {
		return nil, err
	}

	if !isNAT64Address(ipv6Header.Destination, cfg) {
		return nil, errors.New("destination address is not in NAT64 range")
	}

	ipv4Destination, err := translateToIPv4(ipv6Header.Destination, cfg)
	if err != nil {
		return nil, err
	}

	ipv4Pkt, err := createIPv4Packet(ipv6Header, ipv4Destination, ipv6Payload)
	if err != nil {
		return nil, err
	}

	return ipv4Pkt, nil
}

func parseIPv6Header(pkt []byte) (*IPv6Header, []byte, error) {
	if len(pkt) < 40 {
		return nil, nil, errors.New("packet too short for IPv6 header")
	}

	header := &IPv6Header{
		Version:       (pkt[0] >> 4) & 0x0F,
		TrafficClass:  pkt[0] & 0x0F,
		FlowLabel:     uint32(pkt[1])<<16 | uint32(pkt[2])<<8 | uint32(pkt[3]),
		PayloadLength: uint16(pkt[4])<<8 | uint16(pkt[5]),
		NextHeader:    pkt[6],
		HopLimit:      pkt[7],
		Source:        pkt[8:24],
		Destination:   pkt[24:40],
	}

	payload := pkt[40:]
	return header, payload, nil
}

func isNAT64Address(dest []byte, cfg *config.Config) bool {
	if len(dest) >= 8 {
		return dest[0] == 0x64 && dest[1] == 0xFF && dest[2] == 0x9B
	}
	return false
}

func translateToIPv4(dest []byte, cfg *config.Config) (net.IP, error) {
	if len(dest) < 16 {
		return nil, errors.New("invalid IPv6 address length")
	}

	ipv4Addr := net.IPv4(dest[12], dest[13], dest[14], dest[15])
	return ipv4Addr, nil
}

func createIPv4Packet(ipv6Header *IPv6Header, ipv4Dest net.IP, ipv6Payload []byte) ([]byte, error) {
	ipv4Packet := append([]byte{}, ipv6Payload...)

	return ipv4Packet, nil
}
