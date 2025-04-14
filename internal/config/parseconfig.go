package config

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "bridgeconfig.yaml"

var Default_NAT64_prefix = "64:ff9b::"

type BridgeConfig struct {
	Interface string `yaml:"interface"`
	NAT64IP   string `yaml:"nat64_ip"`
	DestIPpath string `yaml:"dest_ip_path"`
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

	if err := validate(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validate(c *BridgeConfig) error {
	if c.Interface == "" {
		return errors.New("missing interface in configuration")
	}
	if c.NAT64IP == "" {
		return errors.New("missing NAT64 IP in configuration")
	}
	return nil
}

func (c *BridgeConfig) GetInterface() string {
	return c.Interface
}

func (c *BridgeConfig) GetNAT64IP() (string,error) {
	if c.NAT64IP == "" {
		return "", errors.New("NAT64 IP is not set")
	}

	return c.NAT64IP, nil
}

func (c *BridgeConfig) GetNAT64Prefix() (string,error) {
	if c.NAT64IP == "" {
		return "", errors.New("NAT64 IP is not set")
	}
	return Default_NAT64_prefix, nil
}

func (c *BridgeConfig) GetDestIPPath() (string,error) {
	if c.DestIPpath == "" {
		return "", errors.New("destination IP path is not set")
	}
	return c.DestIPpath, nil
}


func GetDestIpAddress() (func() (string, bool), error) {
	config, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	
	filepath, err := config.GetDestIPPath()
	if err != nil {
		return nil, err
	}
	
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	var ips []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			ips = append(ips, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	index := 0
	return func() (string, bool) {
		if index >= len(ips) {
			return "", false
		}
		ip := ips[index]
		index++
		return ip, true
	}, nil
}