package cmd

import (
	"net"
	"os"
	"strings"

	"github.com/mdxabu/bridge/internal/gateway/translator"
	"github.com/mdxabu/bridge/internal/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Interface string `yaml:"interface"`
	NAT64IP   string `yaml:"nat64_ip"`
	DestIPpath    string `yaml:"dest_ip_path"`
	DestDomainPath string `yaml:"dest_domain_path"`
	// IPv4      string `yaml:"ipv4"`
}

var bridgeConfig Config

func getIPFromInterface(interfaceName string, wantIPv6 bool) (string, bool) {
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.Error("Failed to get network interfaces: %v", err)
		return "", false
	}

	for _, iface := range ifaces {
		if interfaceName != "" && iface.Name != interfaceName {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipnet.IP

			if wantIPv6 {
				if ip.To16() != nil && ip.To4() == nil && !ip.IsLinkLocalUnicast() {
					return ip.String(), true
				}
			} else {
				if ip.To4() != nil {
					return ip.String(), true
				}
			}
		}
	}
	return "", false
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the bridge configuration",
	Long: `The init command creates a bridgeconfig.yaml file to store
			information for IPv4 and IPv6 translation.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat("bridgeconfig.yaml"); err == nil {
			logger.Warn("Configuration file 'bridgeconfig.yaml' already exists!")
			return
		}

		interfaceName, _ := cmd.Flags().GetString("interface")

		if interfaceName == "" {
			ifaces, err := net.Interfaces()
			if err != nil {
				logger.Fatal("Failed to get network interfaces: %v", err)
			}

			for _, iface := range ifaces {
				if strings.Contains(strings.ToLower(iface.Name), "wi-fi") ||
					strings.Contains(strings.ToLower(iface.Name), "wireless") {
					interfaceName = iface.Name
					logger.Info("Found WiFi interface: %s", interfaceName)
					break
				}
			}
		}

		bridgeConfig.Interface = interfaceName

	
		if ip, found := getIPFromInterface(interfaceName, false); found {
			nat64_ip := translator.GetNAT64Prefix(ip)
			bridgeConfig.NAT64IP = nat64_ip
			logger.Info("Configured the NAT64 Address on your interface IP Address")
		} else {
			logger.Warn("No IP address found for interface: %s", interfaceName)
			bridgeConfig.NAT64IP = ""
		}

		bridgeConfig.DestIPpath = "ipv4.txt"
		bridgeConfig.DestDomainPath = "domains.txt"

		data, err := yaml.Marshal(&bridgeConfig)
		if err != nil {
			logger.Fatal("Failed to serialize configuration: %v", err)
		}

		err = os.WriteFile("bridgeconfig.yaml", data, 0644)
		if err != nil {
			logger.Fatal("Failed to write configuration file: %v", err)
		}

		logger.Completed("Default bridgeconfig.yaml template created successfully!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("interface", "", "Network interface to get IP addresses from (defaults to WiFi interface)")
}
