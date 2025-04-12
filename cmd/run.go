package cmd

import (
	"github.com/mdxabu/bridge/internal/config"
	"github.com/spf13/cobra"
	"github.com/mdxabu/bridge/internal/gateway"
	"github.com/mdxabu/bridge/internal/logger"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the translation process",
	Long:  `Run the translation process to convert IPv6 to IPv4 and forward to the IPv4 container.`,
	Run: func(cmd *cobra.Command, args []string) {
		
		cfg, err := config.ParseConfig()
		if err != nil {
			logger.Fatal("Failed to parse config file: %v", err)
		}
		logger.Info("Starting the translation process...")
		gateway.Start(cfg)
	
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
