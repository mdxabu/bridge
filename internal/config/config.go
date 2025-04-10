package config

type Config struct {
	InterfaceName string `yaml:"interface_name"`
	NAT64Prefix   string `yaml:"nat64_prefix"`
	LogLevel      string `yaml:"log_level"`
	StateTimeout  int    `yaml:"state_timeout"`
}