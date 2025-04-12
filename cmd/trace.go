package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Trace the translation process",
	Long: `Trace the translation process to see how packets are being translated between IPv4 and IPv6.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("trace called")
	},
}

func init() {
	rootCmd.AddCommand(traceCmd)
}
