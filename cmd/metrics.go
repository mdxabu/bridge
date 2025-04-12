package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Display metrics about the translation process",
	Long: `Display metrics about the translation process to monitor its performance and resource usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("metrics called")
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}
