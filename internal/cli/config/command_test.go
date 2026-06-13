package config

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/spf13/cobra"
)

func TestConfigCmd_ShowsActiveConfigSummary(t *testing.T) {
	verbose := true
	active := appconfig.ActiveConfig{
		Source: appconfig.SourceExplicit,
		Path:   "/tmp/project/config.toml",
		Config: &protocoltoml.Config{
			Version:    protocoltoml.ConfigVersion,
			New:        &protocoltoml.ConfigNewSection{},
			VersionCmd: &protocoltoml.ConfigVersionCmd{Verbose: &verbose},
		},
	}

	var buf bytes.Buffer
	root := newSingleCommandRootWithConfig(NewCommand(), active)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"config"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	checks := []string{
		"active config summary:\n",
		"  source: explicit-config\n",
		"  path: /tmp/project/config.toml\n",
		"  loaded: true\n",
		"  version: 1\n",
		"  sections: new, version\n",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("output = %q, want contains %q", output, check)
		}
	}
}

func TestConfigCmd_ShowsHintWhenNoConfigIsActive(t *testing.T) {
	var buf bytes.Buffer
	root := newSingleCommandRootWithConfig(NewCommand(), appconfig.ActiveConfig{Source: appconfig.SourceNone})
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"config"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	checks := []string{
		"  source: none\n",
		"  path: (none)\n",
		"  loaded: false\n",
		"  version: (none)\n",
		"  sections: (none)\n",
		"  hint: use --config <path> or create config.toml in the default user config directory\n",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("output = %q, want contains %q", output, check)
		}
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
