package main

import (
	"fmt"

	"github.com/JackDrogon/project/pkg/version"

	"github.com/spf13/cobra"
)

// newVersionCmd creates the "version" subcommand that prints the build version.
func newVersionCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run: func(cmd *cobra.Command, args []string) {
			if verbose {
				fmt.Println(version.Verbose())
			} else {
				fmt.Println(version.Info())
			}
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed version info")
	return cmd
}
