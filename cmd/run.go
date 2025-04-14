package cmd

import (
	"net"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/gateway/translator"
	"github.com/mdxabu/bridge/internal/logger"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the translation process",
	Long:  `Run the translation process to convert IPv6 to IPv4 and forward to the IPv4 container.`,
	Run: func(cmd *cobra.Command, args []string) {

		cfg, err := config.ParseConfig()
		if err != nil {
			logger.Error("Failed to parse configuration: %v", err)
			return
		}
		
		ip, err := cfg.GetNAT64IP()
		if err != nil {
			logger.Error("Failed to get NAT64 IP: %v", err)
		}

		ipv6Addr := net.ParseIP(ip)
		if ipv6Addr == nil {
			logger.Error("Failed to parse NAT64 IP: %s", ip)
			return
		}
		logger.Info("IPv6 Address: %s", ipv6Addr)
		logger.Info("is NAT64 Address: %v", translator.IsNAT64Address(ipv6Addr.String()))
		logger.Info("Starting the translation process...")
		
	
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
