package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	appconfig "github.com/JackDrogon/project/internal/app/config"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	appversion "github.com/JackDrogon/project/internal/app/version"
	"github.com/spf13/cobra"
)

func TestNewRootCmd(t *testing.T) {
	fsys := fstest.MapFS{
		"go/Makefile": {Data: []byte("build:")},
	}

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})

	// Basic smoke test: verify the command tree can be built without panicking
	cmd := newRootCmd(creator)
	if cmd.Use != "project" {
		t.Errorf("root command Use = %q, want %q", cmd.Use, "project")
	}

	// Verify expected subcommands are registered
	subCmds := cmd.Commands()
	wantCmds := map[string]bool{"new": false, "init": false, "list": false, "inspect": false, "config": false, "version": false, "completion": false}
	for _, sub := range subCmds {
		if _, ok := wantCmds[sub.Name()]; ok {
			wantCmds[sub.Name()] = true
		}
	}
	for name, found := range wantCmds {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestNewRootCmd_RegistersPersistentConfigFlags(t *testing.T) {
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newRootCmd(creator)

	configFlag := cmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Fatal("PersistentFlags().Lookup(" + "config" + ") = nil")
	}
	if configFlag.Usage != "Path to CLI config file" {
		t.Fatalf("config flag usage = %q, want %q", configFlag.Usage, "Path to CLI config file")
	}

	explainFlag := cmd.PersistentFlags().Lookup("explain-config")
	if explainFlag == nil {
		t.Fatal("PersistentFlags().Lookup(" + "explain-config" + ") = nil")
	}
	if explainFlag.Usage != "Explain resolved config sources on stderr" {
		t.Fatalf("explain-config flag usage = %q, want %q", explainFlag.Usage, "Explain resolved config sources on stderr")
	}

	var newCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Name() == string(commandKeyNew) {
			newCmd = subCmd
			break
		}
	}
	if newCmd == nil {
		t.Fatalf("root command missing %q subcommand", commandKeyNew)
	}
	usage := newCmd.UsageString()
	if !strings.Contains(usage, "--config string") {
		t.Fatalf("new command usage = %q, want contains %q", usage, "--config string")
	}
	if !strings.Contains(usage, "--explain-config") {
		t.Fatalf("new command usage = %q, want contains %q", usage, "--explain-config")
	}
}

func TestNewRootCmd_LoadsActiveConfigIntoCommandContext(t *testing.T) {
	oldConfigService := newConfigService
	newConfigService = func() *appconfig.Service {
		return appconfig.NewServiceWithDeps(appconfig.Dependencies{
			UserConfigDir: func() (string, error) { return "/tmp/config-home", nil },
			Join: func(elem ...string) string {
				return "/tmp/config-home/project/config.toml"
			},
			ReadFile: func(path string) ([]byte, error) {
				return []byte("version = 1\n[version]\nverbose = true\n"), nil
			},
		})
	}
	t.Cleanup(func() {
		newConfigService = oldConfigService
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newRootCmd(creator)

	var versionCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Name() == string(commandKeyVersion) {
			versionCmd = subCmd
			break
		}
	}
	if versionCmd == nil {
		t.Fatalf("root command missing %q subcommand", commandKeyVersion)
	}

	if err := cmd.PersistentPreRunE(versionCmd, nil); err != nil {
		t.Fatalf("PersistentPreRunE() error = %v", err)
	}

	active, ok := appconfig.ActiveConfigFromContext(versionCmd.Context())
	if !ok {
		t.Fatal("ActiveConfigFromContext(versionCmd.Context()) ok = false, want true")
	}
	if active.Source != appconfig.SourceUserConfig {
		t.Fatalf("active.Source = %q, want %q", active.Source, appconfig.SourceUserConfig)
	}
	if active.Path != "/tmp/config-home/project/config.toml" {
		t.Fatalf("active.Path = %q, want %q", active.Path, "/tmp/config-home/project/config.toml")
	}
}

func TestConfigSubcommands_AllowMissingExplicitConfigPath(t *testing.T) {
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})

	t.Run("config path prints explicit missing path", func(t *testing.T) {
		var out bytes.Buffer
		cmd := newRootCmd(creator)
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		missingPath := filepath.Join(t.TempDir(), "missing-config.toml")
		cmd.SetArgs([]string{"--config", missingPath, "config", "path"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if got := out.String(); got != missingPath+"\n" {
			t.Fatalf("output = %q, want %q", got, missingPath+"\n")
		}
	})

	t.Run("config init creates explicit missing path", func(t *testing.T) {
		var out bytes.Buffer
		cmd := newRootCmd(creator)
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		configPath := filepath.Join(t.TempDir(), "nested", "config.toml")
		cmd.SetArgs([]string{"--config", configPath, "config", "init"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", configPath, err)
		}
		if string(content) != "version = 1\n" {
			t.Fatalf("config content = %q, want %q", string(content), "version = 1\n")
		}
		if got := out.String(); got != "Created config file: "+configPath+"\n" {
			t.Fatalf("output = %q", got)
		}
	})

	t.Run("config summary shows explicit missing path", func(t *testing.T) {
		var out bytes.Buffer
		cmd := newRootCmd(creator)
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		missingPath := filepath.Join(t.TempDir(), "missing-config.toml")
		cmd.SetArgs([]string{"--config", missingPath, "config"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(out.String(), "  path: "+missingPath+"\n") {
			t.Fatalf("output = %q, want resolved explicit path", out.String())
		}
	})

	t.Run("config validate reports malformed explicit config", func(t *testing.T) {
		var out bytes.Buffer
		cmd := newRootCmd(creator)
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		configPath := filepath.Join(t.TempDir(), "broken.toml")
		if err := os.WriteFile(configPath, []byte("version = 1\n[completion]\nshell = 42\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", configPath, err)
		}
		cmd.SetArgs([]string{"--config", configPath, "config", "validate"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want validation failure")
		}
		if !strings.Contains(err.Error(), "config validation failed") || !strings.Contains(err.Error(), configPath) {
			t.Fatalf("Execute() error = %v, want wrapped validate error for %q", err, configPath)
		}
	})
}

func TestRootHelp_ListsPersistentConfigFlags(t *testing.T) {
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newRootCmd(creator)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	helpText := stdout.String()
	if !strings.Contains(helpText, "Flags:\n") {
		t.Fatalf("stdout = %q, want root flags section", helpText)
	}
	if !strings.Contains(helpText, "--config string") {
		t.Fatalf("stdout = %q, want config flag", helpText)
	}
	if !strings.Contains(helpText, "--explain-config") {
		t.Fatalf("stdout = %q, want explain-config flag", helpText)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty stderr for root help", stderr.String())
	}
}

func TestVersionCmd_DiscoveredUserConfigAppliesVerboseDefault(t *testing.T) {
	useVersionServiceFactoryWith(t, func() *appversion.Service {
		return appversion.NewService(stubVersionProvider{
			info:    "short-version",
			verbose: "Tag:      short-version\nDirty:    false",
		})
	})

	tempConfigHome := t.TempDir()
	configDir := filepath.Join(tempConfigHome, "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", configDir, err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("version = 1\n[version]\nverbose = true\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	oldConfigService := newConfigService
	newConfigService = func() *appconfig.Service {
		return appconfig.NewServiceWithDeps(appconfig.Dependencies{
			UserConfigDir: func() (string, error) { return tempConfigHome, nil },
		})
	}
	t.Cleanup(func() {
		newConfigService = oldConfigService
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newRootCmd(creator)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Tag:") {
		t.Fatalf("output = %q, want verbose version output from discovered user config", stdout.String())
	}
}

func TestVersionCmd_ExplicitConfigSwitchesProfile(t *testing.T) {
	useVersionServiceFactoryWith(t, func() *appversion.Service {
		return appversion.NewService(stubVersionProvider{
			info:    "short-version",
			verbose: "Tag:      short-version\nDirty:    false",
		})
	})

	tempConfigHome := t.TempDir()
	configDir := filepath.Join(tempConfigHome, "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", configDir, err)
	}
	userConfigPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(userConfigPath, []byte("version = 1\n[version]\nverbose = false\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", userConfigPath, err)
	}
	explicitConfigPath := filepath.Join(t.TempDir(), "explicit.toml")
	if err := os.WriteFile(explicitConfigPath, []byte("version = 1\n[version]\nverbose = true\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", explicitConfigPath, err)
	}

	oldConfigService := newConfigService
	newConfigService = func() *appconfig.Service {
		return appconfig.NewServiceWithDeps(appconfig.Dependencies{
			UserConfigDir: func() (string, error) { return tempConfigHome, nil },
		})
	}
	t.Cleanup(func() {
		newConfigService = oldConfigService
	})

	t.Run("discovered user config stays active by default", func(t *testing.T) {
		creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
		cmd := newRootCmd(creator)
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)
		cmd.SetArgs([]string{"version"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if strings.Contains(stdout.String(), "Tag:") {
			t.Fatalf("output = %q, want short version output from discovered config profile", stdout.String())
		}
		if strings.TrimSpace(stdout.String()) != "short-version" {
			t.Fatalf("output = %q, want short-version", stdout.String())
		}
	})

	t.Run("explicit config path overrides discovered profile", func(t *testing.T) {
		creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
		cmd := newRootCmd(creator)
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)
		cmd.SetArgs([]string{"--config", explicitConfigPath, "version"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(stdout.String(), "Tag:") {
			t.Fatalf("output = %q, want verbose version output from explicit config profile", stdout.String())
		}
	})
}

func TestInspectCmd_DiscoveredUserConfigProvidesLangFallback(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)

	tempConfigHome := t.TempDir()
	configDir := filepath.Join(tempConfigHome, "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", configDir, err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("version = 1\n[inspect]\nlang = \"go\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	oldConfigService := newConfigService
	newConfigService = func() *appconfig.Service {
		return appconfig.NewServiceWithDeps(appconfig.Dependencies{
			UserConfigDir: func() (string, error) { return tempConfigHome, nil },
		})
	}
	t.Cleanup(func() {
		newConfigService = oldConfigService
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newRootCmd(creator)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"inspect"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "name: go\n") {
		t.Fatalf("output = %q, want inspect output resolved from discovered config lang", stdout.String())
	}
}

func TestCompletionCmd_ExplicitConfigProvidesShellFallback(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("version = 1\n[completion]\nshell = \"zsh\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newRootCmd(creator)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "completion"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "#compdef project") {
		t.Fatalf("output = %q, want zsh completion generated from explicit config shell", stdout.String())
	}
}
