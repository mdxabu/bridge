package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version   = "0.1.0"
	commit    = "unknown"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of the bridge-cli application",
	Long:  `Version of the bridge-cli application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Bridge CLI version information:")
		fmt.Printf("- Version:    %s\n", version)
		fmt.Printf("- Commit:     %s\n", commit)
		fmt.Printf("- Built on:   %s\n", buildDate)
		fmt.Printf("- Go version: %s\n", runtime.Version())
		fmt.Printf("- OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
