package main

import (
	"fmt"

	"github.com/JackDrogon/project/pkg/version"

	"github.com/spf13/cobra"
)

// newVersionCmd creates the "version" subcommand that prints the build version.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GitTagSha)
		},
	}
}
