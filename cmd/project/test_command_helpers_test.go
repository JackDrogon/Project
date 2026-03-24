package main

import (
	"testing"

	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

func requireSubcommand(t *testing.T, creator *appcreate.Creator, key commandKey) *cobra.Command {
	t.Helper()

	for _, provider := range registeredCommandProviders() {
		if provider.key() == key {
			return provider.buildCommand(commandDependencies{creator: creator})
		}
	}

	t.Fatalf("subcommand %q not found", key)
	return nil
}
