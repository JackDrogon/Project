package main

import (
	catalogcli "github.com/JackDrogon/project/internal/cli/catalog"
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

func init() {
	registerCatalogCommands()
}
