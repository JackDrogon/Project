package main

import (
	scaffoldcli "github.com/JackDrogon/project/internal/cli/scaffold"
	"github.com/spf13/cobra"
)

func buildNewCommand(deps commandDependencies) *cobra.Command {
	return scaffoldcli.NewNewCommand(scaffoldcli.Dependencies{Creator: deps.creator, NewService: newCreateService})
}

func buildInitCommand(deps commandDependencies) *cobra.Command {
	return scaffoldcli.NewInitCommand(scaffoldcli.Dependencies{Creator: deps.creator, NewService: newCreateService})
}

func init() {
	registerOrderedCommand(commandKeyNew, commandOrderNew, buildNewCommand)
	registerOrderedCommand(commandKeyInit, commandOrderInit, buildInitCommand)
}
