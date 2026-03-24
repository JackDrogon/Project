package main

import (
	versioncli "github.com/JackDrogon/project/internal/cli/version"
	"github.com/spf13/cobra"
)

func buildVersionCommand(commandDependencies) *cobra.Command {
	return versioncli.NewCommand(versioncli.Dependencies{NewService: newVersionService})
}

func init() {
	registerOrderedCommand(commandKeyVersion, commandOrderVersion, buildVersionCommand)
}
