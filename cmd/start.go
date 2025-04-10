// bridge/cmd/start.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/core/gateway"
	"github.com/mdxabu/bridge/internal/logger" // Corrected import path
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the NAT64 gateway service",
	Long: `Starts the NAT64 gateway service, reading configuration,
setting up network listeners, and beginning the translation process.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting the NAT64 gateway...")

		// Initialize configuration
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
			viper.AddConfigPath(".")         // Search current directory
			viper.AddConfigPath("./configs") // Search configs directory
		}

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Println("Error: Configuration file not found. Please provide a config file using --config or place config.yaml in the current or ./configs directory.")
				return
			} else {
				fmt.Printf("Error reading config file: %s\n", err)
				return
			}
		}

		var cfg config.Config
		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Printf("Error unmarshalling config: %s\n", err)
			return
		}

		// Initialize logger
		log := logger.New(cfg.LogLevel)
		log.Info("Configuration loaded successfully")
		log.Debug("Loaded configuration", "config", cfg)

		// Initialize and start the NAT64 gateway
		nat64Gateway, err := gateway.New(cfg, log)
		if err != nil {
			log.Error("Failed to initialize gateway", "error", err)
			return
		}

		log.Info("NAT64 gateway initialized. Starting...")
		if err := nat64Gateway.Start(); err != nil {
			log.Error("Failed to start gateway", "error", err)
			os.Exit(1) // Terminate the application
		}

		fmt.Println("NAT64 gateway started. Press Ctrl+C to stop.")
		select {}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
