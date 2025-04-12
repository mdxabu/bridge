package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mdxabu/bridge/internal/logger"
)

type ContainerConfig struct {
	ContainerID string `yaml:"container_id"`
	EnabledIPv6 bool   `yaml:"enabled_ipv6"`
	EnabledIPv4 bool   `yaml:"enabled_ipv4"`
}

type NetworkConfig struct {
	IPv6Container ContainerConfig `yaml:"ipv6_container"`
	IPv4Container ContainerConfig `yaml:"ipv4_container"`
	NAT64Prefix   string          `yaml:"nat64_prefix"`
	IPv4Range     string          `yaml:"ipv4_range"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

type BridgeConfig struct {
	Network   NetworkConfig `yaml:"network"`
	Logging   LoggingConfig `yaml:"logging"`
	Metrics   MetricsConfig `yaml:"metrics"`
	Interface string        `yaml:"interface"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the bridge configuration",
	Long: `The init command creates a bridgeconfig.yaml file to store
information about the Docker containers used for IPv4 and IPv6 translation.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := BridgeConfig{
			Network: NetworkConfig{
				IPv6Container: ContainerConfig{
					ContainerID: "ipv6-client-container-id",
					EnabledIPv6: true,
					EnabledIPv4: false,
				},
				IPv4Container: ContainerConfig{
					ContainerID: "ipv4-server-container-id",
					EnabledIPv6: false,
					EnabledIPv4: true,
				},
				NAT64Prefix: "64:ff9b::/96",
				IPv4Range:   "192.0.2.0/24",
			},
			Logging: LoggingConfig{
				Level: "info",
			},
			Metrics: MetricsConfig{
				Enabled: true,
				Listen:  ":9100",
			},
			Interface: "eth0",
		}

		// Check if the file already exists
		if _, err := os.Stat("bridgeconfig.yaml"); err == nil {
			logger.Warn("Configuration file 'bridgeconfig.yaml' already exists!")
			return
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
