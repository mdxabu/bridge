package cmd

import (
	"os"

	"github.com/spf13/cobra"
)



var rootCmd = &cobra.Command{
	Use:   "bridge",
	Short: "Bridge helps you to communicate with IpV6 to IpV4",
	Long: `Bridge is a tool that helps you to communicate with IpV6 to IpV4.`,
	
}


func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


