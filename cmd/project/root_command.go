package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newRootCmd builds the command tree with all subcommands registered explicitly.
func newRootCmd(templateFS embed.FS) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "project",
		Short: "project is a tool to create new project",
	}

	rootCmd.AddCommand(
		newNewCmd(templateFS),
		newListCmd(templateFS),
		newVersionCmd(),
		newCompletionCmd(),
	)

	return rootCmd
}

// Execute runs the root command.
// If an error occurs during execution, it prints the error to stderr
// and exits the program with status code 1.
func Execute(templateFS embed.FS) {
	if err := newRootCmd(templateFS).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
