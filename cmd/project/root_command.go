package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "project",
	Short: "project is a tool to create new project",
}

// Execute runs the root command
// If an error occurs during execution, it prints the error to standard output
// and exits the program with status code 1
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
