package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appconfig "github.com/JackDrogon/project/internal/app/config"
	appversion "github.com/JackDrogon/project/internal/app/version"
	"github.com/spf13/cobra"
)

func TestVersionCmd_ConfigVerboseDefaultAppliesAndFlagWins(t *testing.T) {
	useVersionServiceFactoryWith(t, func() *appversion.Service {
		return appversion.NewService(stubVersionProvider{
			info:    "short-version",
			verbose: "Tag:      short-version\nDirty:    false",
		})
	})

	verbose := true
	active := appconfig.ActiveConfig{
		Source: appconfig.SourceUserConfig,
		Config: &protocoltoml.Config{
			Version: protocoltoml.ConfigVersion,
			VersionCmd: &protocoltoml.ConfigVersionCmd{
				Verbose: &verbose,
			},
		},
	}

	t.Run("config default applies", func(t *testing.T) {
		var buf bytes.Buffer
		root := newSingleCommandRootWithConfig(buildVersionCommand(commandDependencies{}), active)
		root.SetOut(&buf)
		root.SetErr(&buf)
		root.SetArgs([]string{"version"})

		if err := root.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "Tag:") {
			t.Fatalf("output = %q, want verbose output from config default", buf.String())
		}
	})

	t.Run("explicit flag wins", func(t *testing.T) {
		var buf bytes.Buffer
		root := newSingleCommandRootWithConfig(buildVersionCommand(commandDependencies{}), active)
		root.SetOut(&buf)
		root.SetErr(&buf)
		root.SetArgs([]string{"version", "--verbose=false"})

		if err := root.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		if got != "short-version" {
			t.Fatalf("output = %q, want short version output when flag explicitly disables verbose", got)
		}
	})
}

func TestCompletionCmd_ConfigShellDefaultAppliesAndArgWins(t *testing.T) {
	zsh := "zsh"
	active := appconfig.ActiveConfig{
		Source: appconfig.SourceUserConfig,
		Config: &protocoltoml.Config{
			Version: protocoltoml.ConfigVersion,
			Completion: &protocoltoml.ConfigCompletion{
				Shell: &zsh,
			},
		},
	}

	t.Run("config shell default applies", func(t *testing.T) {
		var buf bytes.Buffer
		root := newSingleCommandRootWithConfig(buildCompletionCommand(commandDependencies{}), active)
		root.SetOut(&buf)
		root.SetErr(&buf)
		root.SetArgs([]string{"completion"})

		if err := root.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !strings.Contains(buf.String(), "#compdef project") {
			t.Fatalf("output = %q, want zsh completion generated from config default", buf.String())
		}
	})

	t.Run("explicit arg wins", func(t *testing.T) {
		var buf bytes.Buffer
		root := newSingleCommandRootWithConfig(buildCompletionCommand(commandDependencies{}), active)
		root.SetOut(&buf)
		root.SetErr(&buf)
		root.SetArgs([]string{"completion", "bash"})

		if err := root.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !strings.Contains(buf.String(), "__start_project") {
			t.Fatalf("output = %q, want bash completion when positional arg is explicit", buf.String())
		}
	})
}

func TestCompletionCmd_InvalidConfigShellFailsBeforeGeneration(t *testing.T) {
	invalid := "cmd"
	active := appconfig.ActiveConfig{
		Source: appconfig.SourceUserConfig,
		Config: &protocoltoml.Config{
			Version: protocoltoml.ConfigVersion,
			Completion: &protocoltoml.ConfigCompletion{
				Shell: &invalid,
			},
		},
	}

	var buf bytes.Buffer
	root := newSingleCommandRootWithConfig(buildCompletionCommand(commandDependencies{}), active)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"completion"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid argument \"cmd\"") {
		t.Fatalf("Execute() error = %v, want invalid argument error for config shell", err)
	}
	if strings.Contains(buf.String(), "__start_project") || strings.Contains(buf.String(), "#compdef project") {
		t.Fatalf("output = %q, want no completion generation output on invalid config shell", buf.String())
	}
}

func newSingleCommandRootWithConfig(command *cobra.Command, active appconfig.ActiveConfig) *cobra.Command {
	ctx := appconfig.WithActiveConfig(context.Background(), active)
	root := &cobra.Command{Use: "project"}
	root.SetContext(ctx)
	command.SetContext(ctx)
	root.AddCommand(command)
	return root
}
