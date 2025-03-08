/*
Copyright © 2025 Mohamed Abdullah <110121104039@aalimec.ac.in>
*/
package cmd

import (
	"os"
	"github.com/spf13/cobra"
)



var rootCmd = &cobra.Command{
	Use:   "bridge",
	Short: "",
	Long: ``,
	
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


