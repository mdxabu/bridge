/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of the application",
	Long:  `Version of the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bridge-cli-" + version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
