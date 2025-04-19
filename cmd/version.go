package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var (
	version   = "0.1.0"
	commit    = getGitCommit()
	buildDate = getBuildDate()
)

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return string(out[:len(out)-1])
}

func getBuildDate() string {
	return time.Now().UTC().Format(time.RFC3339)
}

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
