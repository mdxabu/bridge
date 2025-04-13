package utils

import (
	"errors"
	"net"
)

func ExtractAddresses(packet []byte) (string, string, error) {
	if len(packet) < 40 { 
		return "", "", errors.New("packet too short to extract addresses")
	}

	srcAddr := net.IP(packet[8:24]).String()  
	destAddr := net.IP(packet[24:40]).String() 

	return srcAddr, destAddr, nil
}
