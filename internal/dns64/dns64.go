package dns64

import (
	"fmt"
	"net"
	"strings"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/logger"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

func Start() {

	nextDomain, err := config.GetDestDomain()
	if err != nil {
		logger.Error("Failed to get destination domains: %v", err)
		return
	}

	fmt.Println(strings.Repeat("─", 100))
	fmt.Printf("%-30s %-20s %-45s\n", "DOMAIN", "IPV4 ADDRESS", "SYNTHESIZED IPV6 ADDRESS")
	fmt.Println(strings.Repeat("─", 100))

	domainCount := 0

	for {
		domain, ok := nextDomain()
		if !ok {
			break
		}

		ResolveDomain(domain)
		domainCount++
	}

	fmt.Println(strings.Repeat("─", 100))

	if domainCount > 0 {
		logger.Info("Completed resolving %d domains", domainCount)
	} else {
		logger.Warn("No domains were resolved")
	}
}

func ResolveDomain(domain string) {
	logger.Debug("Resolving domain: %s", domain)

	ips, err := net.LookupIP(domain)
	if err != nil {
		logger.Error("Failed to look up IP addresses for %s: %v", domain, err)
		printDomainRow(domain, "Resolution failed", "")
		return
	}

	var ipv4List []string
	var ipv6List []string

	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4List = append(ipv4List, ip.String())
		} else {
			ipv6List = append(ipv6List, ip.String())
		}
	}

	conf, err := config.ParseConfig()
	if err != nil {
		logger.Error("Failed to parse configuration: %v", err)
		return
	}

	nat64Prefix, err := conf.GetNAT64Prefix()
	if err != nil {
		logger.Error("Failed to get NAT64 prefix: %v", err)
		return
	}

	if len(ipv4List) == 0 {
		printDomainRow(domain, "No IPv4 address", "Cannot synthesize")
		return
	}

	for i, ipv4 := range ipv4List {
		synth, err := SynthesizeIPv6(ipv4, nat64Prefix)
		if err != nil {
			logger.Error("Failed to synthesize IPv6 for %s: %v", ipv4, err)
			synth = "Error synthesizing"
		}

		if i == 0 {
			printDomainRow(domain, ipv4, synth)
		} else {
			printDomainRow("", ipv4, synth)
		}
	}

	if len(ipv6List) > 0 {
		logger.Debug("Native IPv6 addresses for %s:", domain)
		for _, ip := range ipv6List {
			logger.Debug("  %s", ip)
		}
	} else {
		logger.Debug("No native IPv6 addresses for %s", domain)
	}
}

func printDomainRow(domain, ipv4, ipv6 string) {
	domainColor := colorBlue
	ipv4Color := colorReset
	ipv6Color := colorGreen

	if ipv4 == "No IPv4 address" || ipv4 == "Resolution failed" {
		ipv4Color = colorRed
	}

	if ipv6 == "Cannot synthesize" || ipv6 == "Error synthesizing" {
		ipv6Color = colorRed
	}

	fmt.Printf("%s%-30s%s %s%-20s%s %s%-45s%s\n",
		domainColor, domain, colorReset,
		ipv4Color, ipv4, colorReset,
		ipv6Color, ipv6, colorReset)
}

func SynthesizeIPv6(ipv4Str, prefixStr string) (string, error) {
	ipv4 := net.ParseIP(ipv4Str).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("invalid IPv4 address: %s", ipv4Str)
	}

	if prefixStr == "" {
		prefixStr = "64:ff9b::"
	}

	if !strings.Contains(prefixStr, "::") {
		prefixStr = prefixStr + "::"
	}

	prefixStr = strings.Split(prefixStr, "/")[0]

	prefix := net.ParseIP(prefixStr)
	if prefix == nil {
		return "", fmt.Errorf("invalid prefix: %s", prefixStr)
	}

	ipv6 := make(net.IP, 16)
	copy(ipv6, prefix.To16())

	copy(ipv6[12:], ipv4)

	return ipv6.String(), nil
}
