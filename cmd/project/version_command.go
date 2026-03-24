package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func buildVersionCommand(commandDependencies) *cobra.Command {
	var verbose bool
	service := newVersionService()

	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), service.Verbose())
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), service.Info())
			return err
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed version info")
	return cmd
}

func init() {
	registerOrderedCommand(commandKeyVersion, commandOrderVersion, buildVersionCommand)
}
