package main

import (
	"testing"

	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

func requireSubcommand(t *testing.T, creator *appcreate.Creator, key commandKey) *cobra.Command {
	t.Helper()

	cmd, ok := buildRegisteredCommand(commandDependencies{creator: creator}, key)
	if ok {
		return cmd
	}

	t.Fatalf("subcommand %q not found", key)
	return nil
}
