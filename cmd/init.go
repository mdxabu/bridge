package cmd

import (
	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/logger"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the bridge configuration",
	Long:  `Create a default configuration file for the Bridge NAT64 translator.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Initializing bridge configuration...")

		err := config.CreateDefaultConfig()
		if err != nil {
			logger.Error("Failed to create configuration: %v", err)
			return
		}

		logger.Success("Configuration file created: bridgeconfig.yaml")
		logger.Info("Default settings:")
		logger.Info("  NAT64 Prefix: 64:ff9b::/96")
		logger.Info("  Gateway IP: 64:ff9b::1")
		logger.Info("  API Port: 8080")
		logger.Info("\nEdit bridgeconfig.yaml to customize settings")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
