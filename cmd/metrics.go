package cmd

import (
	"fmt"
	"github.com/mdxabu/bridge/internal/metrics"
	"github.com/spf13/cobra"
)

var nat64 bool

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Display metrics about the translation process",
	Long:  `Display metrics about the translation process to monitor its performance and resource usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting Web Metrics Dashboard...")
		metrics.StartWebDashboard(nat64)
	},
}

func init() {
	metricsCmd.Flags().BoolVar(&nat64, "nat64", false, "Enable NAT64 mode")
	rootCmd.AddCommand(metricsCmd)
}
