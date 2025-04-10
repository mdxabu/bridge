package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/core/gateway"
	"github.com/mdxabu/bridge/internal/logger" 
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the NAT64 gateway service",
	Long: `Starts the NAT64 gateway service, reading configuration,
setting up network listeners, and beginning the translation process.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting the NAT64 gateway...")

		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
			viper.AddConfigPath(".")         
			viper.AddConfigPath("./configs") 
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

		log := logger.New(logger.InfoLevel)
		if cfg.LogLevel == "debug" {
			logger.SetDefaultLogLevel(logger.DebugLevel)
		} else if cfg.LogLevel == "error" {
			logger.SetDefaultLogLevel(logger.ErrorLevel)
		}

		log.Info("Configuration loaded successfully")
		log.Debug("Loaded configuration: %+v", cfg)

		nat64Gateway, err := gateway.New(cfg, log)
		if err != nil {
			log.Error("Failed to initialize gateway: %v", err)
			return
		}

		log.Info("NAT64 gateway initialized. Starting...")
		if err := nat64Gateway.Start(); err != nil {
			log.Error("Failed to start gateway: %v", err)
			os.Exit(1) 
		}

		fmt.Println("NAT64 gateway started. Press Ctrl+C to stop.")
		select {}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
