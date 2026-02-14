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
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), version.Verbose())
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), version.Info())
			return err
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed version info")
	return cmd
}
