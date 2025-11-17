package translator

import (
	"encoding/binary"
	"fmt"
	"net"
)

// PacketType represents the protocol type
type PacketType int

const (
	PacketTypeTCP PacketType = iota
	PacketTypeUDP
	PacketTypeICMP
	PacketTypeUnknown
)

// Packet represents a network packet with parsed headers
type Packet struct {
	Type       PacketType
	SrcIP      net.IP
	DstIP      net.IP
	SrcPort    uint16
	DstPort    uint16
	Protocol   uint8
	Payload    []byte
	RawData    []byte
	IsIPv6     bool
	IPv6Header []byte
	IPv4Header []byte
	TCPHeader  []byte
	UDPHeader  []byte
	ICMPHeader []byte
}

// ParseIPv6Packet parses an IPv6 packet
func ParseIPv6Packet(data []byte) (*Packet, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("packet too small for IPv6 header")
	}

	pkt := &Packet{
		RawData:    data,
		IsIPv6:     true,
		IPv6Header: data[:40],
	}

	// Parse IPv6 header
	pkt.Protocol = data[6]
	pkt.SrcIP = net.IP(data[8:24])
	pkt.DstIP = net.IP(data[24:40])

	payload := data[40:]

	// Parse transport layer based on protocol
	switch pkt.Protocol {
	case 6: // TCP
		pkt.Type = PacketTypeTCP
		if len(payload) < 20 {
			return nil, fmt.Errorf("packet too small for TCP header")
		}
		pkt.TCPHeader = payload[:20]
		pkt.SrcPort = binary.BigEndian.Uint16(payload[0:2])
		pkt.DstPort = binary.BigEndian.Uint16(payload[2:4])
		pkt.Payload = payload

	case 17: // UDP
		pkt.Type = PacketTypeUDP
		if len(payload) < 8 {
			return nil, fmt.Errorf("packet too small for UDP header")
		}
		pkt.UDPHeader = payload[:8]
		pkt.SrcPort = binary.BigEndian.Uint16(payload[0:2])
		pkt.DstPort = binary.BigEndian.Uint16(payload[2:4])
		pkt.Payload = payload

	case 58: // ICMPv6
		pkt.Type = PacketTypeICMP
		if len(payload) < 4 {
			return nil, fmt.Errorf("packet too small for ICMPv6 header")
		}
		pkt.ICMPHeader = payload
		pkt.Payload = payload

	default:
		pkt.Type = PacketTypeUnknown
		pkt.Payload = payload
	}

	return pkt, nil
}

// ParseIPv4Packet parses an IPv4 packet
func ParseIPv4Packet(data []byte) (*Packet, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("packet too small for IPv4 header")
	}

	pkt := &Packet{
		RawData: data,
		IsIPv6:  false,
	}

	// Parse IPv4 header
	headerLen := int(data[0]&0x0F) * 4
	if len(data) < headerLen {
		return nil, fmt.Errorf("invalid IPv4 header length")
	}

	pkt.IPv4Header = data[:headerLen]
	pkt.Protocol = data[9]
	pkt.SrcIP = net.IPv4(data[12], data[13], data[14], data[15])
	pkt.DstIP = net.IPv4(data[16], data[17], data[18], data[19])

	payload := data[headerLen:]

	// Parse transport layer
	switch pkt.Protocol {
	case 6: // TCP
		pkt.Type = PacketTypeTCP
		if len(payload) < 20 {
			return nil, fmt.Errorf("packet too small for TCP header")
		}
		pkt.TCPHeader = payload[:20]
		pkt.SrcPort = binary.BigEndian.Uint16(payload[0:2])
		pkt.DstPort = binary.BigEndian.Uint16(payload[2:4])
		pkt.Payload = payload

	case 17: // UDP
		pkt.Type = PacketTypeUDP
		if len(payload) < 8 {
			return nil, fmt.Errorf("packet too small for UDP header")
		}
		pkt.UDPHeader = payload[:8]
		pkt.SrcPort = binary.BigEndian.Uint16(payload[0:2])
		pkt.DstPort = binary.BigEndian.Uint16(payload[2:4])
		pkt.Payload = payload

	case 1: // ICMPv4
		pkt.Type = PacketTypeICMP
		if len(payload) < 4 {
			return nil, fmt.Errorf("packet too small for ICMPv4 header")
		}
		pkt.ICMPHeader = payload
		pkt.Payload = payload

	default:
		pkt.Type = PacketTypeUnknown
		pkt.Payload = payload
	}

	return pkt, nil
}

// String returns a string representation of the packet
func (p *Packet) String() string {
	proto := "Unknown"
	switch p.Type {
	case PacketTypeTCP:
		proto = "TCP"
	case PacketTypeUDP:
		proto = "UDP"
	case PacketTypeICMP:
		proto = "ICMP"
	}

	return fmt.Sprintf("%s: %s:%d -> %s:%d", proto, p.SrcIP, p.SrcPort, p.DstIP, p.DstPort)
}
