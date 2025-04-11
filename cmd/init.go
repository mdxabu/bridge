package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mdxabu/bridge/internal/logger"
)

type BridgeConfig struct {
	IPv4ContainerID string `yaml:"ipv4_container_id"`
	IPv6ContainerID string `yaml:"ipv6_container_id"`
	Description     string `yaml:"description"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the bridge configuration",
	Long: `The init command creates a bridgeconfig.yaml file to store
information about the Docker containers used for IPv4 and IPv6 translation.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create a default configuration template instead of asking for input
		config := BridgeConfig{
			IPv4ContainerID: "ipv4-container-id",
			IPv6ContainerID: "ipv6-container-id",
			Description:     "Default bridge configuration for Docker containers",
		}

		data, err := yaml.Marshal(&config)
		if err != nil {
			logger.Fatal("Failed to serialize configuration: %v", err)
		}

		err = os.WriteFile("bridgeconfig.yaml", data, 0644)
		if err != nil {
			logger.Fatal("Failed to write configuration file: %v", err)
		}

		logger.Info("Default bridgeconfig.yaml template created successfully!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
