package translator

import (
	"fmt"
	"net"
)

var nat64_prefix = "64:ff9b::"

func GetNAT64Prefix(ipv4Str string) string {
	ip := net.ParseIP(ipv4Str).To4()
	if ip == nil {
		return ""
	}

	hexIP := fmt.Sprintf("%x:%x", uint16(ip[0])<<8|uint16(ip[1]), uint16(ip[2])<<8|uint16(ip[3]))

	nat64Address := nat64_prefix + hexIP

	return nat64Address
}
