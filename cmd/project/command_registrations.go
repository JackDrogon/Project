package main

import (
	catalogcli "github.com/JackDrogon/project/internal/cli/catalog"
	completioncli "github.com/JackDrogon/project/internal/cli/completion"
	configcli "github.com/JackDrogon/project/internal/cli/config"
	scaffoldcli "github.com/JackDrogon/project/internal/cli/scaffold"
	versioncli "github.com/JackDrogon/project/internal/cli/version"
	"github.com/spf13/cobra"
)

func registerCatalogCommands() {
	registerOrderedCommand(commandKeyList, commandOrderList, func(commandDependencies) *cobra.Command {
		return catalogcli.NewListCommand(catalogcli.Dependencies{NewService: newCatalogService})
	})
	registerOrderedCommand(commandKeyInspect, commandOrderInspect, func(commandDependencies) *cobra.Command {
		return catalogcli.NewInspectCommand(catalogcli.Dependencies{NewService: newCatalogService})
	})
}

func buildNewCommand(deps commandDependencies) *cobra.Command {
	return scaffoldcli.NewNewCommand(scaffoldcli.Dependencies{Creator: deps.creator, NewService: newCreateService})
}

func buildInitCommand(deps commandDependencies) *cobra.Command {
	return scaffoldcli.NewInitCommand(scaffoldcli.Dependencies{Creator: deps.creator, NewService: newCreateService})
}

func buildVersionCommand(commandDependencies) *cobra.Command {
	return versioncli.NewCommand(versioncli.Dependencies{NewService: newVersionService})
}

func buildConfigCommand(commandDependencies) *cobra.Command {
	return configcli.NewCommand()
}

func buildCompletionCommand(commandDependencies) *cobra.Command {
	return completioncli.NewCommand()
}

func init() {
	registerCatalogCommands()
	registerOrderedCommand(commandKeyNew, commandOrderNew, buildNewCommand)
	registerOrderedCommand(commandKeyInit, commandOrderInit, buildInitCommand)
	registerOrderedCommand(commandKeyConfig, commandOrderConfig, buildConfigCommand)
	registerOrderedCommand(commandKeyVersion, commandOrderVersion, buildVersionCommand)
	registerOrderedCommand(commandKeyCompletion, commandOrderCompletion, buildCompletionCommand)
}
