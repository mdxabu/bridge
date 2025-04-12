package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfig(t *testing.T) {
	// Create a temporary test config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	testConfig := `
network:
  ipv6_container:
    container_id: ipv6-client-container-id
    enabled_ipv6: true
    enabled_ipv4: false
  ipv4_container:
    container_id: ipv4-server-container-id
    enabled_ipv6: false
    enabled_ipv4: true
  nat64_prefix: 64:ff9b::/96
  ipv4_range: 192.0.2.0/24
logging:
  level: info
metrics:
  enabled: true
  listen: :9100
interface: eth0
`

	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	config, err := ParseConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Test getters
	if got := config.GetIPv6ContainerID(); got != "ipv6-client-container-id" {
		t.Errorf("GetIPv6ContainerID() = %s, want %s", got, "ipv6-client-container-id")
	}

	if got := config.GetIPv4ContainerID(); got != "ipv4-server-container-id" {
		t.Errorf("GetIPv4ContainerID() = %s, want %s", got, "ipv4-server-container-id")
	}

	if got := config.GetNAT64Prefix(); got != "64:ff9b::/96" {
		t.Errorf("GetNAT64Prefix() = %s, want %s", got, "64:ff9b::/96")
	}

	if got := config.GetIPv4Range(); got != "192.0.2.0/24" {
		t.Errorf("GetIPv4Range() = %s, want %s", got, "192.0.2.0/24")
	}

	if got := config.GetLogLevel(); got != "info" {
		t.Errorf("GetLogLevel() = %s, want %s", got, "info")
	}

	if got := config.IsMetricsEnabled(); got != true {
		t.Errorf("IsMetricsEnabled() = %v, want %v", got, true)
	}

	if got := config.GetMetricsListen(); got != ":9100" {
		t.Errorf("GetMetricsListen() = %s, want %s", got, ":9100")
	}

	if got := config.GetInterface(); got != "eth0" {
		t.Errorf("GetInterface() = %s, want %s", got, "eth0")
	}
}
