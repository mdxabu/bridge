package cmd

import (
	"fmt"
	"os/exec"

	"github.com/mdxabu/bridge/internal/logger"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Docker networks for Bridge",
	Long:  `Create and configure Docker networks required for IPv6 and IPv4 containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Setting up Docker networks for Bridge...")

		// Create IPv6 network
		logger.Info("Creating IPv6-only Docker network...")
		err := createIPv6Network()
		if err != nil {
			logger.Error("Failed to create IPv6 network: %v", err)
		} else {
			logger.Success("IPv6 network 'bridge-ipv6' created successfully")
		}

		// Create IPv4 network
		logger.Info("Creating IPv4-only Docker network...")
		err = createIPv4Network()
		if err != nil {
			logger.Error("Failed to create IPv4 network: %v", err)
		} else {
			logger.Success("IPv4 network 'bridge-ipv4' created successfully")
		}

		logger.Info("\nNetworks created successfully!")
		logger.Info("Next steps:")
		logger.Info("  1. Run 'bridge start' to start the NAT64 translator")
		logger.Info("  2. Connect your IPv6-only containers to 'bridge-ipv6'")
		logger.Info("  3. Connect your IPv4-only containers to 'bridge-ipv4'")
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove Docker networks created by Bridge",
	Long:  `Remove the Docker networks and clean up Bridge resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Cleaning up Docker networks...")

		// Remove IPv6 network
		err := removeNetwork("bridge-ipv6")
		if err != nil {
			logger.Warn("Failed to remove IPv6 network: %v", err)
		} else {
			logger.Success("IPv6 network removed")
		}

		// Remove IPv4 network
		err = removeNetwork("bridge-ipv4")
		if err != nil {
			logger.Warn("Failed to remove IPv4 network: %v", err)
		} else {
			logger.Success("IPv4 network removed")
		}

		logger.Success("Cleanup complete")
	},
}

func createIPv6Network() error {
	cmd := exec.Command("docker", "network", "create",
		"--ipv6",
		"--subnet=fd00:64::/64",
		"--driver=bridge",
		"bridge-ipv6")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(output), err)
	}

	return nil
}

func createIPv4Network() error {
	cmd := exec.Command("docker", "network", "create",
		"--subnet=10.64.0.0/16",
		"--driver=bridge",
		"bridge-ipv4")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(output), err)
	}

	return nil
}

func removeNetwork(name string) error {
	cmd := exec.Command("docker", "network", "rm", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(output), err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(cleanupCmd)
}
