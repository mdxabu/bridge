package core

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// TranslateIPv4ToIPv6 performs the translation of an IPv4 packet to IPv6.
// This function should implement the logic described in RFC 6145.
func TranslateIPv4ToIPv6(packet gopacket.Packet, ipv6DstIP net.IP) (gopacket.Packet, error) {
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ipv4Layer == nil {
		return packet, fmt.Errorf("no IPv4 layer found")
	}
	ipv4 := ipv4Layer.(*layers.IPv4)

	transportLayer := packet.TransportLayer()
	if transportLayer == nil {
		return packet, fmt.Errorf("no transport layer found in IPv4 packet")
	}

	// Create IPv6 layer
	ipv6 := &layers.IPv6{
		Version:    6,
		TrafficClass: 0,
		FlowLabel:  0,
		HopLimit:   ipv4.TTL,
		SrcIP:      TranslateIPv4ToIPv6Address(ipv4.SrcIP), // Synthesize IPv6 source
		DstIP:      ipv6DstIP,                               // Use provided IPv6 destination
	}

	// Handle transport layer (TCP, UDP, ICMP) and update checksums
	var payload gopacket.Payload
	switch t := transportLayer.(type) {
	case *layers.TCP:
		// Update checksum (requires pseudo-header)
		t.SetNetworkLayerForChecksum(ipv6)
	case *layers.UDP:
		// Update checksum (requires pseudo-header)
		t.SetNetworkLayerForChecksum(ipv6)
	case *layers.ICMPv4:
		// Translate ICMPv4 to ICMPv6 if needed (RFC 6145)
		// This is a simplified example, full ICMP translation is complex
		ipv6ICMP := &layers.ICMPv6{
			TypeCode: layers.ICMPv6TypeEchoRequest, // Default
		}
		if t.TypeCode == layers.ICMPv4TypeEchoRequest {
			ipv6ICMP.TypeCode = layers.ICMPv6TypeEchoRequest
			ipv6ICMP.Code = t.Code
		} else if t.TypeCode == layers.ICMPv4TypeEchoReply {
			ipv6ICMP.TypeCode = layers.ICMPv6TypeEchoReply
			ipv6ICMP.Code = t.Code
		} else if t.TypeCode == layers.ICMPv4TypeDestinationUnreachable {
			ipv6ICMP.TypeCode = layers.ICMPv6TypeDestinationUnreachable
			// Map codes if needed
			ipv6ICMP.Code = 0 // Example
		} else if t.TypeCode == layers.ICMPv4TypeTimeExceeded {
			ipv6ICMP.TypeCode = layers.ICMPv6TypeTimeExceeded
			ipv6ICMP.Code = t.Code
		} else if t.TypeCode == layers.ICMPv4TypePacketTooBig {
			ipv6ICMP.TypeCode = layers.ICMPv6TypePacketTooBig
			ipv6ICMP.Code = 0
		} else {
			return packet, fmt.Errorf("unsupported ICMPv4 type for translation: %v", t.TypeCode)
		}
		payload = gopacket.Payload(t.Payload) // Carry original ICMP payload
		// ICMPv6 checksum is calculated over the ICMPv6 message itself + pseudo-header
		ipv6ICMP.SetNetworkLayerForChecksum(ipv6)
		// Prepend ICMPv6 layer
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
		err := gopacket.SerializeLayers(buf, opts, ipv6, ipv6ICMP, payload)
		if err != nil {
			return packet, fmt.Errorf("failed to serialize layers for ICMPv6: %v", err)
		}
		return gopacket.NewPacket(buf.Bytes(), layers.LinkTypeEthernet, gopacket.Default), nil
	default:
		return packet, fmt.Errorf("unsupported transport protocol for IPv4 to IPv6: %T", transportLayer)
	}

	// Serialize layers
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	err := gopacket.SerializeLayers(buf, opts, ipv6, transportLayer, gopacket.Payload(transportLayer.Payload()))
	if err != nil {
		return packet, fmt.Errorf("failed to serialize layers for IPv4 to IPv6: %v", err)
	}

	return gopacket.NewPacket(buf.Bytes(), layers.LinkTypeEthernet, gopacket.Default), nil
}

// TranslateIPv4ToIPv6Address synthesizes an IPv6 address from an IPv4 address (RFC 6052).
// Assuming the well-known prefix 64:ff9b::/96.
func TranslateIPv4ToIPv6Address(ipv4Addr net.IP) net.IP {
	nat64Prefix := net.ParseIP("64:ff9b::")
	if nat64Prefix == nil {
		// This should not happen if the prefix is hardcoded correctly
		return nil
	}
	ipv6Addr := make(net.IP, net.IPv6len)
	copy(ipv6Addr[:12], nat64Prefix.To16()[:12]) // Copy the /96 prefix
	copy(ipv6Addr[12:], ipv4Addr.To4())        // Embed the 32-bit IPv4 address
	return ipv6Addr
}