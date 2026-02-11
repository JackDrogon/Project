package main

import (
	"fmt"
	"os"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

// newRootCmd builds the command tree with all subcommands registered explicitly.
func newRootCmd(creator *scaffold.Creator) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "project",
		Short: "project is a tool to create new project",
	}

	rootCmd.AddCommand(
		newNewCmd(creator),
		newListCmd(creator),
		newVersionCmd(),
		newCompletionCmd(),
	)

	return rootCmd
}

// Execute runs the root command.
// If an error occurs during execution, it prints the error to stderr
// and exits the program with status code 1.
func Execute(creator *scaffold.Creator) {
	if err := newRootCmd(creator).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
