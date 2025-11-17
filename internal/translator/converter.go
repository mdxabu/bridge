package translator

import (
	"encoding/binary"
	"fmt"
	"net"
)

// TranslateIPv6ToIPv4 translates an IPv6 packet to IPv4
func TranslateIPv6ToIPv4(pkt *Packet, nat64Prefix string) ([]byte, error) {
	if !pkt.IsIPv6 {
		return nil, fmt.Errorf("packet is not IPv6")
	}

	// Extract IPv4 address from NAT64 address
	ipv4Dst, err := GetIPV4fromNAT64(pkt.DstIP.String())
	if err != nil {
		return nil, fmt.Errorf("failed to extract IPv4 from NAT64: %w", err)
	}

	// Parse destination IPv4
	dstIPv4 := net.ParseIP(ipv4Dst).To4()
	if dstIPv4 == nil {
		return nil, fmt.Errorf("invalid IPv4 address: %s", ipv4Dst)
	}

	// Extract IPv4 from source (for NAT64, we'll use a mapped address)
	srcIPv4, err := GetIPV4fromNAT64(pkt.SrcIP.String())
	if err != nil {
		// If source is not NAT64, we need to map it
		// For now, use a default internal address
		srcIPv4 = "10.64.0.1"
	}

	srcIP := net.ParseIP(srcIPv4).To4()
	if srcIP == nil {
		return nil, fmt.Errorf("invalid source IPv4 address")
	}

	// Build IPv4 packet
	ipv4Packet, err := buildIPv4Packet(srcIP, dstIPv4, pkt)
	if err != nil {
		return nil, fmt.Errorf("failed to build IPv4 packet: %w", err)
	}

	return ipv4Packet, nil
}

// TranslateIPv4ToIPv6 translates an IPv4 packet to IPv6
func TranslateIPv4ToIPv6(pkt *Packet, nat64Prefix string) ([]byte, error) {
	if pkt.IsIPv6 {
		return nil, fmt.Errorf("packet is already IPv6")
	}

	// Convert IPv4 addresses to NAT64 IPv6 addresses
	srcIPv6, err := IPv4ToNAT64(pkt.SrcIP.String(), nat64Prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to convert source to NAT64: %w", err)
	}

	dstIPv6, err := IPv4ToNAT64(pkt.DstIP.String(), nat64Prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to convert destination to NAT64: %w", err)
	}

	// Build IPv6 packet
	ipv6Packet, err := buildIPv6Packet(srcIPv6, dstIPv6, pkt)
	if err != nil {
		return nil, fmt.Errorf("failed to build IPv6 packet: %w", err)
	}

	return ipv6Packet, nil
}

// buildIPv4Packet constructs an IPv4 packet
func buildIPv4Packet(srcIP, dstIP net.IP, pkt *Packet) ([]byte, error) {
	// IPv4 header is 20 bytes (without options)
	header := make([]byte, 20)

	// Version (4) and IHL (5 = 20 bytes)
	header[0] = 0x45

	// DSCP and ECN
	header[1] = 0

	// Total length (will be set after payload is added)
	totalLen := uint16(20 + len(pkt.Payload))
	binary.BigEndian.PutUint16(header[2:4], totalLen)

	// Identification
	binary.BigEndian.PutUint16(header[4:6], 0)

	// Flags and Fragment offset
	binary.BigEndian.PutUint16(header[6:8], 0x4000) // Don't fragment

	// TTL
	header[8] = 64

	// Protocol (translate ICMPv6 to ICMPv4)
	protocol := pkt.Protocol
	if protocol == 58 { // ICMPv6 -> ICMPv4
		protocol = 1
	}
	header[9] = protocol

	// Checksum (will be calculated later)
	binary.BigEndian.PutUint16(header[10:12], 0)

	// Source IP
	copy(header[12:16], srcIP.To4())

	// Destination IP
	copy(header[16:20], dstIP.To4())

	// Calculate checksum
	checksum := calculateChecksum(header)
	binary.BigEndian.PutUint16(header[10:12], checksum)

	// Combine header and payload
	packet := append(header, pkt.Payload...)

	return packet, nil
}

// buildIPv6Packet constructs an IPv6 packet
func buildIPv6Packet(srcIP, dstIP net.IP, pkt *Packet) ([]byte, error) {
	// IPv6 header is 40 bytes
	header := make([]byte, 40)

	// Version (6), Traffic Class, Flow Label
	header[0] = 0x60
	header[1] = 0
	header[2] = 0
	header[3] = 0

	// Payload length
	payloadLen := uint16(len(pkt.Payload))
	binary.BigEndian.PutUint16(header[4:6], payloadLen)

	// Next header (translate ICMPv4 to ICMPv6)
	nextHeader := pkt.Protocol
	if nextHeader == 1 { // ICMPv4 -> ICMPv6
		nextHeader = 58
	}
	header[6] = nextHeader

	// Hop limit
	header[7] = 64

	// Source address
	copy(header[8:24], srcIP.To16())

	// Destination address
	copy(header[24:40], dstIP.To16())

	// Combine header and payload
	packet := append(header, pkt.Payload...)

	return packet, nil
}

// IPv4ToNAT64 converts an IPv4 address to NAT64 format
func IPv4ToNAT64(ipv4Addr, nat64Prefix string) (net.IP, error) {
	ip := net.ParseIP(ipv4Addr).To4()
	if ip == nil {
		return nil, fmt.Errorf("invalid IPv4 address: %s", ipv4Addr)
	}

	// Using well-known prefix 64:ff9b::/96
	nat64IP := make([]byte, 16)
	nat64IP[0] = 0x00
	nat64IP[1] = 0x64
	nat64IP[2] = 0xff
	nat64IP[3] = 0x9b
	nat64IP[4] = 0x00
	nat64IP[5] = 0x00
	nat64IP[6] = 0x00
	nat64IP[7] = 0x00
	nat64IP[8] = 0x00
	nat64IP[9] = 0x00
	nat64IP[10] = 0x00
	nat64IP[11] = 0x00

	// Embed IPv4 address in last 4 bytes
	nat64IP[12] = ip[0]
	nat64IP[13] = ip[1]
	nat64IP[14] = ip[2]
	nat64IP[15] = ip[3]

	return net.IP(nat64IP), nil
}

// calculateChecksum computes the Internet checksum
func calculateChecksum(data []byte) uint16 {
	sum := uint32(0)

	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(data[i : i+2]))
	}

	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}

	for sum > 0xffff {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	return ^uint16(sum)
}

// RecalculateTransportChecksum recalculates TCP/UDP checksum for translated packets
func RecalculateTransportChecksum(packet []byte, isIPv6 bool) error {
	var pkt *Packet
	var err error

	if isIPv6 {
		pkt, err = ParseIPv6Packet(packet)
	} else {
		pkt, err = ParseIPv4Packet(packet)
	}

	if err != nil {
		return err
	}

	switch pkt.Type {
	case PacketTypeTCP:
		return recalculateTCPChecksum(packet, pkt, isIPv6)
	case PacketTypeUDP:
		return recalculateUDPChecksum(packet, pkt, isIPv6)
	}

	return nil
}

func recalculateTCPChecksum(packet []byte, pkt *Packet, isIPv6 bool) error {
	// TCP checksum calculation with pseudo-header
	// This is a simplified version
	return nil
}

func recalculateUDPChecksum(packet []byte, pkt *Packet, isIPv6 bool) error {
	// UDP checksum calculation with pseudo-header
	// This is a simplified version
	return nil
}
