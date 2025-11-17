package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/logger"
	"github.com/mdxabu/bridge/internal/tun"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the NAT64 bridge",
	Long:  `Start the NAT64 bridge to translate packets between IPv6 and IPv4 networks.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Starting NAT64 Bridge...")

		// Load configuration
		cfg, err := config.ParseConfig()
		if err != nil {
			logger.Error("Failed to parse configuration: %v", err)
			return
		}

		nat64Prefix := cfg.GetNAT64Prefix()
		nat64Gateway := cfg.GetNAT64Gateway()

		// Create bridge
		bridge, err := tun.NewBridge(nat64Prefix)
		if err != nil {
			logger.Error("Failed to create bridge: %v", err)
			return
		}

		// Create TUN interfaces
		logger.Info("Creating TUN interfaces...")

		err = bridge.CreateTUNInterface("tun-ipv6", true)
		if err != nil {
			logger.Error("Failed to create IPv6 TUN interface: %v", err)
			logger.Warn("Note: TUN interface creation requires root/admin privileges")
			return
		}

		err = bridge.CreateTUNInterface("tun-ipv4", false)
		if err != nil {
			logger.Error("Failed to create IPv4 TUN interface: %v", err)
			return
		}

		// Start the bridge
		err = bridge.Start()
		if err != nil {
			logger.Error("Failed to start bridge: %v", err)
			return
		}

		logger.Success("NAT64 Bridge is running")
		logger.Info("NAT64 Prefix: %s", nat64Prefix)
		logger.Info("NAT64 Gateway IP: %s", nat64Gateway)
		logger.Info("Press Ctrl+C to stop")

		// Wait for interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down...")
		bridge.Stop()
		logger.Success("Bridge stopped successfully")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
