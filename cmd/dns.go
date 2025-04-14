package cmd

import (
	"github.com/mdxabu/bridge/internal/dns64"
	"github.com/mdxabu/bridge/internal/logger"
	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Resolve domains and synthesize IPv6 addresses",
	Long: `Resolve domain names to their IPv4 and IPv6 addresses,
and synthesize IPv6 addresses from IPv4 addresses using 
the configured NAT64 prefix.

This command reads domain names from the configured domain file
and displays their IPv4, IPv6, and synthesized addresses.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("DNS64 resolution starting...")
		dns64.Start()
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)
}
