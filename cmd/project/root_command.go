package main

import (
	"fmt"
	"io"
	"os"

	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

var exitFunc = os.Exit
var stderrWriter io.Writer = os.Stderr

// newRootCmd builds the command tree with all subcommands registered explicitly.
func newRootCmd(creator *appcreate.Creator) *cobra.Command {
	deps := commandDependencies{creator: creator}
	rootCmd := &cobra.Command{
		Use:   "project",
		Short: "project is a tool to create new project",
	}

	addRegisteredCommands(rootCmd, deps)

	return rootCmd
}

// Execute runs the root command.
// If an error occurs during execution, it prints the error to stderr
// and exits the program with status code 1.
func Execute(creator *appcreate.Creator) {
	if err := newRootCmd(creator).Execute(); err != nil {
		_, _ = fmt.Fprintln(stderrWriter, err)
		exitFunc(1)
	}
}
