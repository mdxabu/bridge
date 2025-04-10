package cmd

import (
	"fmt"
	"os"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the application configuration",
	Long:  `Allows you to show or validate the application configuration.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Displays the currently loaded configuration",
	Long:  `Prints the currently loaded configuration in YAML format.`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.ConfigFileUsed() == "" {
			fmt.Println("No configuration file loaded.")
			return
		}

		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		configMap := viper.AllSettings()
		fmt.Printf("%+v\n", configMap)
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates a configuration file",
	Long:  `Validates the syntax of a specified configuration file without starting the service.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := args[0]
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Error reading config file '%s': %s\n", configFile, err)
			os.Exit(1)
		}
		fmt.Printf("Configuration file '%s' is valid.\n", configFile)

		var cfg config.Config
		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Printf("Warning: Could not unmarshal config for deeper validation: %s\n", err)
		} else {
			fmt.Println("Basic configuration structure seems valid.")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
}