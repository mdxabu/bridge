package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "bridge",
	Short: "A Stateful NAT64 Gateway",
	Long: `bridge is a Stateful NAT64 gateway that enables communication
between IPv6-only clients and IPv4-only servers.`,
	Run: func(cmd *cobra.Command, args []string) {
		asciiart := `
██████╗ ██████╗ ██╗██████╗  ██████╗ ███████╗
██╔══██╗██╔══██╗██║██╔══██╗██╔════╝ ██╔════╝
██████╔╝██████╔╝██║██║  ██║██║  ███╗█████╗  
██╔══██╗██╔══██╗██║██║  ██║██║   ██║██╔══╝  
██████╔╝██║  ██║██║██████╔╝╚██████╔╝███████╗
╚═════╝ ╚═╝  ╚═╝╚═╝╚═════╝  ╚═════╝ ╚══════╝
`
		fmt.Println(asciiart)
		fmt.Println("Welcome to the bridge CLI!, please use the --help flag to see available commands.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
}