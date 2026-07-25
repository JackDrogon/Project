package config

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/spf13/cobra"
)

func TestConfigCmd_ShowsActiveConfigSummary(t *testing.T) {
	active := loadActiveConfigForTest(t, "version = 1\n\n[new]\n\n[version]\nverbose = true\n")

	var buf bytes.Buffer
	root := newSingleCommandRootWithContext(NewCommand(Dependencies{NewService: appconfig.NewService}), active, appconfig.Context{})
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
		"  path: " + active.Path + "\n",
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
	tempHome := t.TempDir()
	configHome := filepath.Join(tempHome, ".config")
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	var buf bytes.Buffer
	root := newSingleCommandRootWithContext(NewCommand(Dependencies{NewService: appconfig.NewService}), appconfig.ActiveConfig{Source: appconfig.SourceNone}, appconfig.Context{})
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"config"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	checks := []string{
		"  source: none\n",
		"  path: " + filepath.Join(configHome, "project", "config.toml") + "\n",
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

func TestConfigCmd_ShowsResolvedExplicitMissingPathAndLoadError(t *testing.T) {
	explicitPath := filepath.Join(t.TempDir(), "missing.toml")
	loadErr := os.ErrNotExist

	var buf bytes.Buffer
	root := newSingleCommandRootWithContext(
		NewCommand(Dependencies{NewService: appconfig.NewService}),
		appconfig.ActiveConfig{Source: appconfig.SourceNone},
		appconfig.Context{ExplicitPath: explicitPath},
	)
	root.SetContext(appconfig.WithLoadError(root.Context(), loadErr))
	for _, sub := range root.Commands() {
		sub.SetContext(root.Context())
	}
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"config"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	checks := []string{
		"  source: none\n",
		"  path: " + explicitPath + "\n",
		"  load_error: file does not exist\n",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("output = %q, want contains %q", output, check)
		}
	}
}

func TestConfigPathCmd_PrintsResolvedPath(t *testing.T) {
	tempHome := t.TempDir()
	configHome := filepath.Join(tempHome, ".config")
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	var buf bytes.Buffer
	root := newSingleCommandRootWithContext(NewCommand(Dependencies{NewService: appconfig.NewService}), appconfig.ActiveConfig{Source: appconfig.SourceNone}, appconfig.Context{})
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"config", "path"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := filepath.Join(configHome, "project", "config.toml") + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestConfigInitCmd_CreatesSeedConfigFile(t *testing.T) {
	tempHome := t.TempDir()
	configHome := filepath.Join(tempHome, ".config")
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	var buf bytes.Buffer
	root := newSingleCommandRootWithContext(NewCommand(Dependencies{NewService: appconfig.NewService}), appconfig.ActiveConfig{Source: appconfig.SourceNone}, appconfig.Context{})
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"config", "init"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	path := filepath.Join(configHome, "project", "config.toml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(content) != "version = 1\n" {
		t.Fatalf("config content = %q, want %q", string(content), "version = 1\n")
	}
	if got := buf.String(); got != "Created config file: "+path+"\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestConfigInitCmd_UsesExplicitConfigPathEvenWhenMissing(t *testing.T) {
	explicitPath := filepath.Join(t.TempDir(), "nested", "custom.toml")

	var buf bytes.Buffer
	root := newSingleCommandRootWithContext(
		NewCommand(Dependencies{NewService: appconfig.NewService}),
		appconfig.ActiveConfig{Source: appconfig.SourceNone},
		appconfig.Context{ExplicitPath: explicitPath},
	)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"config", "init"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content, err := os.ReadFile(explicitPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", explicitPath, err)
	}
	if string(content) != "version = 1\n" {
		t.Fatalf("config content = %q, want %q", string(content), "version = 1\n")
	}
}

func TestConfigValidateCmd_SucceedsForLoadedConfig(t *testing.T) {
	active := loadActiveConfigForTest(t, "version = 1\n")

	var buf bytes.Buffer
	root := newSingleCommandRootWithContext(NewCommand(Dependencies{NewService: appconfig.NewService}), active, appconfig.Context{ExplicitPath: active.Path})
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"config", "validate"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := buf.String(); got != "Config is valid: "+active.Path+"\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestConfigValidateCmd_FailsWhenConfigDoesNotExist(t *testing.T) {
	tempHome := t.TempDir()
	configHome := filepath.Join(tempHome, ".config")
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	root := newSingleCommandRootWithContext(NewCommand(Dependencies{NewService: appconfig.NewService}), appconfig.ActiveConfig{Source: appconfig.SourceNone}, appconfig.Context{})
	root.SetArgs([]string{"config", "validate"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want missing config error")
	}
	want := filepath.Join(configHome, "project", "config.toml")
	if !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("Execute() error = %v, want missing config path", err)
	}
}

func TestConfigValidateCmd_FailsWithLoadError(t *testing.T) {
	explicitPath := filepath.Join(t.TempDir(), "broken.toml")
	loadErr := fmt.Errorf("failed to decode config file %s: broken field", explicitPath)

	root := newSingleCommandRootWithContext(
		NewCommand(Dependencies{NewService: appconfig.NewService}),
		appconfig.ActiveConfig{Source: appconfig.SourceNone},
		appconfig.Context{ExplicitPath: explicitPath},
	)
	root.SetContext(appconfig.WithLoadError(root.Context(), loadErr))
	for _, sub := range root.Commands() {
		sub.SetContext(root.Context())
	}
	root.SetArgs([]string{"config", "validate"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want validation failure")
	}
	if !strings.Contains(err.Error(), explicitPath) || !strings.Contains(err.Error(), "config validation failed") {
		t.Fatalf("Execute() error = %v, want wrapped validation error", err)
	}
}

func newSingleCommandRootWithContext(command *cobra.Command, active appconfig.ActiveConfig, loadCtx appconfig.Context) *cobra.Command {
	ctx := appconfig.WithLoadContext(appconfig.WithActiveConfig(context.Background(), active), loadCtx)
	root := &cobra.Command{Use: "project"}
	root.SetContext(ctx)
	command.SetContext(ctx)
	root.AddCommand(command)
	return root
}

// loadActiveConfigForTest loads a TOML config document through the app
// config loader from a real temp file.
func loadActiveConfigForTest(t *testing.T, content string) appconfig.ActiveConfig {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	active, err := appconfig.NewService().LoadActiveConfig(appconfig.Context{ExplicitPath: path})
	if err != nil {
		t.Fatalf("LoadActiveConfig() error = %v", err)
	}

	return active
}
