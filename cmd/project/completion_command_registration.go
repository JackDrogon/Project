package main

import (
	completioncli "github.com/JackDrogon/project/internal/cli/completion"
	"github.com/spf13/cobra"
)

func buildCompletionCommand(commandDependencies) *cobra.Command {
	return completioncli.NewCommand()
}

func init() {
	registerOrderedCommand(commandKeyCompletion, commandOrderCompletion, buildCompletionCommand)
}
