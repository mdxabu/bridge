package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "bridgeconfig.yaml"

type BridgeConfig struct {
	Interface    string `yaml:"interface"`
	NAT64Prefix  string `yaml:"nat64_prefix"`
	NAT64Gateway string `yaml:"nat64_gateway"`
	APIPort      int    `yaml:"api_port"`
}

func ParseConfig() (*BridgeConfig, error) {
	data, err := os.ReadFile(DefaultConfigPath)
	if err != nil {
		return nil, err
	}

	var config BridgeConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if config.NAT64Prefix == "" {
		config.NAT64Prefix = "64:ff9b::/96"
	}
	if config.NAT64Gateway == "" {
		config.NAT64Gateway = "64:ff9b::1"
	}
	if config.APIPort == 0 {
		config.APIPort = 8080
	}

	return &config, nil
}

func (c *BridgeConfig) GetInterface() string {
	return c.Interface
}

func (c *BridgeConfig) GetNAT64Prefix() string {
	return c.NAT64Prefix
}

func (c *BridgeConfig) GetNAT64Gateway() string {
	return c.NAT64Gateway
}

func (c *BridgeConfig) GetAPIPort() int {
	return c.APIPort
}

func CreateDefaultConfig() error {
	config := BridgeConfig{
		Interface:    "",
		NAT64Prefix:  "64:ff9b::/96",
		NAT64Gateway: "64:ff9b::1",
		APIPort:      8080,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	return os.WriteFile(DefaultConfigPath, data, 0644)
}
