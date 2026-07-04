package protocoltoml

import (
	"strings"
	"testing"
)

func TestConfigV1_DecodeValidSectionsAndPresence(t *testing.T) {
	content := []byte(`version = 1

[new]
lang = "go"
project_name = "demo"
module = "example.com/demo"
git_mode = "init+commit"
signoff = false

[new.inputs]
author = "alice"

[init]
lang = "go"
target_dir = "."
git_mode = "none"
signoff = false

[init.inputs]
go_version = "1.25"

[list]
format = "toml"
compact = false
detail = false
table = false
sort = "name"
min_governance = "basic"
required_assets = ["ci"]

[inspect]
lang = "go"
format = "text"
compact = false
mode = "render"

[version]
verbose = false

[completion]
shell = "bash"
`)

	cfg, err := DecodeConfig(content, "config.toml")
	if err != nil {
		t.Fatalf("DecodeConfig() error = %v", err)
	}

	if cfg.Version != ConfigVersion {
		t.Fatalf("cfg.Version = %d, want %d", cfg.Version, ConfigVersion)
	}
	if cfg.New == nil || cfg.Init == nil || cfg.List == nil || cfg.Inspect == nil || cfg.VersionCmd == nil || cfg.Completion == nil {
		t.Fatalf("cfg sections = %#v, want all command sections present", cfg)
	}

	if cfg.New.Signoff == nil || *cfg.New.Signoff {
		t.Fatalf("cfg.New.Signoff = %#v, want explicit false", cfg.New.Signoff)
	}
	if cfg.Init.Signoff == nil || *cfg.Init.Signoff {
		t.Fatalf("cfg.Init.Signoff = %#v, want explicit false", cfg.Init.Signoff)
	}
	if cfg.List.Format == nil || *cfg.List.Format != "toml" {
		t.Fatalf("cfg.List.Format = %#v, want explicit toml", cfg.List.Format)
	}
	if cfg.List.Detail == nil || *cfg.List.Detail {
		t.Fatalf("cfg.List.Detail = %#v, want explicit false", cfg.List.Detail)
	}
	if len(cfg.List.RequiredAsset) != 1 || cfg.List.RequiredAsset[0] != "ci" {
		t.Fatalf("cfg.List.RequiredAsset = %#v, want [ci]", cfg.List.RequiredAsset)
	}
	if cfg.Inspect.Format == nil || *cfg.Inspect.Format != "text" {
		t.Fatalf("cfg.Inspect.Format = %#v, want explicit text", cfg.Inspect.Format)
	}
	if cfg.Inspect.Compact == nil || *cfg.Inspect.Compact {
		t.Fatalf("cfg.Inspect.Compact = %#v, want explicit false", cfg.Inspect.Compact)
	}
	if cfg.VersionCmd.Verbose == nil || *cfg.VersionCmd.Verbose {
		t.Fatalf("cfg.VersionCmd.Verbose = %#v, want explicit false", cfg.VersionCmd.Verbose)
	}

	if cfg.New.Module == nil || *cfg.New.Module != "example.com/demo" {
		t.Fatalf("cfg.New.Module = %#v, want explicit module", cfg.New.Module)
	}
	if cfg.Init.Module != nil {
		t.Fatalf("cfg.Init.Module = %#v, want nil when unset", cfg.Init.Module)
	}

	if got := cfg.New.Inputs["author"]; got != "alice" {
		t.Fatalf("cfg.New.Inputs[author] = %q, want %q", got, "alice")
	}
	if got := cfg.Init.Inputs["go_version"]; got != "1.25" {
		t.Fatalf("cfg.Init.Inputs[go_version] = %q, want %q", got, "1.25")
	}
}

func TestConfigV1_TrailingComments(t *testing.T) {
	t.Run("version line and section header accept trailing comments", func(t *testing.T) {
		content := []byte("version = 1 # config schema version\n\n[new] # scaffold defaults\nlang = \"go\"\n")

		cfg, err := DecodeConfig(content, "config.toml")
		if err != nil {
			t.Fatalf("DecodeConfig() error = %v", err)
		}
		if cfg.Version != ConfigVersion {
			t.Fatalf("cfg.Version = %d, want %d", cfg.Version, ConfigVersion)
		}
		if cfg.New == nil || cfg.New.Lang == nil || *cfg.New.Lang != "go" {
			t.Fatalf("cfg.New = %#v, want lang go", cfg.New)
		}
	})

	t.Run("commented-out version is not a declaration", func(t *testing.T) {
		_, err := DecodeConfig([]byte("# version = 1\n[new]\nlang = \"go\"\n"), "config.toml")
		if err == nil || !strings.Contains(err.Error(), "must declare version") {
			t.Fatalf("DecodeConfig() error = %v, want must declare version", err)
		}
	})
}

func TestConfigV1_RejectsUnknownFieldsLegacyJSONAndInvalidEnums(t *testing.T) {
	t.Run("unknown field", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[new]\nunknown = true\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
	})

	t.Run("removed new keys rejected", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[new]\ntarget_dir = \"demo\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}

		_, err = DecodeConfig([]byte("version = 1\n[new]\nforce = true\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}

		_, err = DecodeConfig([]byte("version = 1\n[new]\ndry_run = true\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
	})

	t.Run("removed init keys rejected", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[init]\ndry_run = false\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
	})

	t.Run("removed list and inspect keys rejected", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[list]\ntoml = true\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}

		_, err = DecodeConfig([]byte("version = 1\n[list]\nhas_repo_asset = [\"ci\"]\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}

		_, err = DecodeConfig([]byte("version = 1\n[inspect]\ntoml = true\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
	})

	t.Run("section version does not satisfy top level version", func(t *testing.T) {
		_, err := DecodeConfig([]byte("[new]\nversion = 1\nlang = \"go\"\nproject_name = \"demo\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "must declare version") {
			t.Fatalf("DecodeConfig() error = %v, want missing top-level version error", err)
		}
	})

	t.Run("legacy json", func(t *testing.T) {
		_, err := DecodeConfig([]byte(`{"version":1,"new":{"lang":"go"}}`), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "legacy JSON") {
			t.Fatalf("DecodeConfig() error = %v, want legacy JSON rejection", err)
		}
	})

	t.Run("invalid new git mode", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[new]\ngit_mode = \"broken\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "new.git_mode") {
			t.Fatalf("DecodeConfig() error = %v, want new.git_mode validation error", err)
		}
	})

	t.Run("invalid list sort", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[list]\nsort = \"size\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "list.sort") {
			t.Fatalf("DecodeConfig() error = %v, want list.sort validation error", err)
		}
	})

	t.Run("invalid list format", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[list]\nformat = \"yaml\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "list.format") {
			t.Fatalf("DecodeConfig() error = %v, want list.format validation error", err)
		}
	})

	t.Run("invalid inspect mode", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[inspect]\nmode = \"bogus\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "inspect.mode") {
			t.Fatalf("DecodeConfig() error = %v, want inspect.mode validation error", err)
		}
	})

	t.Run("invalid completion shell", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[completion]\nshell = \"cmd\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "completion.shell") {
			t.Fatalf("DecodeConfig() error = %v, want completion.shell validation error", err)
		}
	})

	t.Run("invalid inspect format", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[inspect]\nformat = \"json\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "inspect.format") {
			t.Fatalf("DecodeConfig() error = %v, want inspect.format validation error", err)
		}
	})
}

func TestConfigV1_RejectsReservedCreateInputKeys(t *testing.T) {
	t.Run("new inputs", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[new.inputs]\nlang = \"go\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "new.inputs") || !strings.Contains(err.Error(), "reserved key") {
			t.Fatalf("DecodeConfig() error = %v, want reserved key rejection in new.inputs", err)
		}
	})

	t.Run("init inputs", func(t *testing.T) {
		_, err := DecodeConfig([]byte("version = 1\n[init.inputs]\ntarget_dir = \"demo\"\n"), "config.toml")
		if err == nil {
			t.Fatal("DecodeConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "init.inputs") || !strings.Contains(err.Error(), "reserved key") {
			t.Fatalf("DecodeConfig() error = %v, want reserved key rejection in init.inputs", err)
		}
	})
}
