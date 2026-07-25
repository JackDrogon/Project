package main

import (
	catalogcli "github.com/JackDrogon/project/internal/cli/catalog"
	completioncli "github.com/JackDrogon/project/internal/cli/completion"
	configcli "github.com/JackDrogon/project/internal/cli/config"
	scaffoldcli "github.com/JackDrogon/project/internal/cli/scaffold"
	versioncli "github.com/JackDrogon/project/internal/cli/version"
	"github.com/spf13/cobra"
)

// commandKey names a top-level subcommand so the tree, the config-command
// lookup, and the tests all agree on one spelling.
type commandKey string

const (
	commandKeyNew        commandKey = "new"
	commandKeyInit       commandKey = "init"
	commandKeyList       commandKey = "list"
	commandKeyInspect    commandKey = "inspect"
	commandKeyConfig     commandKey = "config"
	commandKeyVersion    commandKey = "version"
	commandKeyCompletion commandKey = "completion"
)

// subcommands builds the top-level commands. Adding a command means adding one
// line here: there is no registry, no init(), and no ordering table. Cobra
// sorts commands for help output, so this order is declaration order only.
func subcommands(deps dependencies) []*cobra.Command {
	scaffoldDeps := scaffoldcli.Dependencies{NewCreator: deps.newCreator, NewService: deps.newCreateService}
	catalogDeps := catalogcli.Dependencies{NewService: deps.newCatalogService}

	return []*cobra.Command{
		scaffoldcli.NewNewCommand(scaffoldDeps),
		scaffoldcli.NewInitCommand(scaffoldDeps),
		catalogcli.NewListCommand(catalogDeps),
		catalogcli.NewInspectCommand(catalogDeps),
		configcli.NewCommand(configcli.Dependencies{NewService: deps.newConfigService}),
		versioncli.NewCommand(versioncli.Dependencies{NewService: deps.newVersionService}),
		completioncli.NewCommand(completioncli.Dependencies{NewService: deps.newCompletionService}),
	}
}
