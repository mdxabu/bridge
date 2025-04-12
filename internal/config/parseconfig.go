package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "bridgeconfig.yaml"

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

type Config struct {
	Network   NetworkConfig `yaml:"network"`
	Logging   LoggingConfig `yaml:"logging"`
	Metrics   MetricsConfig `yaml:"metrics"`
	Interface string        `yaml:"interface"`
}


func ParseConfig() (*Config, error) {
	data, err := os.ReadFile(DefaultConfigPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Network.IPv6Container.ContainerID == "" {
		return errors.New("missing IPv6 container ID")
	}
	if c.Network.IPv4Container.ContainerID == "" {
		return errors.New("missing IPv4 container ID")
	}
	if c.Network.NAT64Prefix == "" {
		return errors.New("missing NAT64 prefix")
	}
	if c.Network.IPv4Range == "" {
		return errors.New("missing IPv4 range")
	}
	return nil
}

func (c *Config) GetIPv6ContainerID() string {
	return c.Network.IPv6Container.ContainerID
}

func (c *Config) GetIPv4ContainerID() string {
	return c.Network.IPv4Container.ContainerID
}

func (c *Config) GetNAT64Prefix() string {
	return c.Network.NAT64Prefix
}

func (c *Config) GetIPv4Range() string {
	return c.Network.IPv4Range
}

func (c *Config) GetLogLevel() string {
	return c.Logging.Level
}

func (c *Config) IsMetricsEnabled() bool {
	return c.Metrics.Enabled
}

func (c *Config) GetMetricsListen() string {
	return c.Metrics.Listen
}

func (c *Config) GetInterface() string {
	return c.Interface
}
