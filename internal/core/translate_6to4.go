package core

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// TranslateIPv6ToIPv4 performs the translation of an IPv6 packet to IPv4.
// This function should implement the logic described in RFC 6145.
func TranslateIPv6ToIPv4(packet gopacket.Packet, ipv4DstIP net.IP) (gopacket.Packet, error) {
	ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
	if ipv6Layer == nil {
		return packet, fmt.Errorf("no IPv6 layer found")
	}
	ipv6 := ipv6Layer.(*layers.IPv6)

	transportLayer := packet.TransportLayer()
	if transportLayer == nil {
		return packet, fmt.Errorf("no transport layer found in IPv6 packet")
	}

	// Create IPv4 layer
	ipv4 := &layers.IPv4{
		Version:  4,
		IHL:      5, // Minimum header length
		TTL:      ipv6.HopLimit,
		Protocol: GetIPv4ProtocolNumber(ipv6.NextHeader),
		SrcIP:    ExtractIPv4FromIPv6(ipv6.SrcIP, net.ParseIP("64:ff9b::")), // Extract IPv4 source (if applicable)
		DstIP:    ipv4DstIP,
	}

	// Handle transport layer (TCP, UDP, ICMP) and update checksums
	switch t := transportLayer.(type) {
	case *layers.TCP:
		// Update checksum (requires pseudo-header)
		t.SetNetworkLayerForChecksum(ipv4)
	case *layers.UDP:
		// Update checksum (requires pseudo-header)
		t.SetNetworkLayerForChecksum(ipv4)
	case *layers.ICMPv6:
		// Translate ICMPv6 to ICMPv4 if needed (RFC 6145)
		// This is a simplified example
		ipv4ICMP := &layers.ICMPv4{}
		if t.TypeCode.Type() == layers.ICMPv6TypeEchoRequest {
			ipv4ICMP.TypeCode = layers.ICMPv4TypeEchoRequest
			ipv4ICMP.Code = t.Code
		} else if t.TypeCode.Type() == layers.ICMPv6TypeEchoReply {
			ipv4ICMP.TypeCode = layers.ICMPv4TypeEchoReply
			ipv4ICMP.Code = t.Code
		} else if t.TypeCode.Type() == layers.ICMPv6TypeDestinationUnreachable {
			ipv4ICMP.TypeCode = layers.ICMPv4TypeDestinationUnreachable
			// Need to map codes as well
			ipv4ICMP.Code = 0 // Example
		} else if t.TypeCode.Type() == layers.ICMPv6TypePacketTooBig {
			ipv4ICMP.TypeCode = layers.ICMPv4TypeDestinationUnreachable
			ipv4ICMP.Code = 9 // Fragmentation needed and DF set
		} else if t.TypeCode.Type() == layers.ICMPv6TypeTimeExceeded {
			ipv4ICMP.TypeCode = layers.ICMPv4TypeTimeExceeded
			ipv4ICMP.Code = t.Code
		} else {
			return packet, fmt.Errorf("unsupported ICMPv6 type for translation: %v", t.TypeCode)
		}
		ipv4ICMP.SetNetworkLayerForChecksum(ipv4)
		// Prepend ICMPv4 layer
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
		err := gopacket.SerializeLayers(buf, opts, ipv4, ipv4ICMP, gopacket.Payload(t.Payload))
		if err != nil {
			return packet, fmt.Errorf("failed to serialize layers for ICMPv4: %v", err)
		}
		return gopacket.NewPacket(buf.Bytes(), layers.LinkTypeEthernet, gopacket.Default), nil
	default:
		return packet, fmt.Errorf("unsupported transport protocol for IPv6 to IPv4: %T", transportLayer)
	}

	// Serialize layers
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	err := gopacket.SerializeLayers(buf, opts, ipv4, transportLayer, gopacket.Payload(transportLayer.Payload()))
	if err != nil {
		return packet, fmt.Errorf("failed to serialize layers for IPv6 to IPv4: %v", err)
	}

	return gopacket.NewPacket(buf.Bytes(), layers.LinkTypeEthernet, gopacket.Default), nil
}

// ExtractIPv4FromIPv6 extracts the embedded IPv4 address from an IPv6 address (RFC 6052).
// Assuming the well-known prefix 64:ff9b::/96.
func ExtractIPv4FromIPv6(ipv6Addr net.IP, nat64Prefix net.IP) net.IP {
	if len(ipv6Addr) != net.IPv6len || len(nat64Prefix) != net.IPv6len {
		return nil
	}
	if !ipv6Addr[:12].Equal(nat64Prefix[:12]) {
		return nil
	}
	return ipv6Addr[12:]
}

// GetIPv4ProtocolNumber maps IPv6 Next Header to IPv4 Protocol Number.
func GetIPv4ProtocolNumber(nextHeader layers.IPProtocol) uint8 {
	switch nextHeader {
	case layers.IPProtocolTCP:
		return 6
	case layers.IPProtocolUDP:
		return 17
	case layers.IPProtocolICMPv6:
		return 58 // ICMP for IPv6
	default:
		return uint8(nextHeader) // Try to cast if no specific mapping
	}
}

// IsIPv6InNAT64Prefix checks if an IPv6 address falls within the NAT64 prefix.
func IsIPv6InNAT64Prefix(ipv6Addr net.IP, nat64Prefix net.IP) bool {
	if len(ipv6Addr) != net.IPv6len || len(nat64Prefix) != net.IPv6len {
		return false
	}
	return ipv6Addr[:12].Equal(nat64Prefix[:12])
}

// GetTransportPorts extracts source and destination ports from a transport layer.
func GetTransportPorts(transportLayer gopacket.TransportLayer) (srcPort, dstPort uint16) {
	if tcpLayer, ok := transportLayer.(*layers.TCP); ok {
		srcPort = uint16(tcpLayer.SrcPort)
		dstPort = uint16(tcpLayer.DstPort)
	} else if udpLayer, ok := transportLayer.(*layers.UDP); ok {
		srcPort = uint16(udpLayer.SrcPort)
		dstPort = uint16(udpLayer.DstPort)
	}
	return
}