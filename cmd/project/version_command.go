package main

import (
	"fmt"

	"github.com/JackDrogon/project/pkg/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.GitTagSha)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
