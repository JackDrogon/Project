package create

import (
	"path/filepath"
	"testing"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("NewService() = nil")
	}
}

func TestServiceBuildNewOptions_UsesReplayWhenArgOmitted(t *testing.T) {
	svc := NewService()
	replayPath := writeReplayForCreateServiceTest(t, protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(CommandNew),
		Template: protocoltoml.ReplayTemplate{Lang: "cpp"},
		Project: protocoltoml.ReplayProject{
			Name:       "replayed-demo",
			TargetDir:  "replayed-demo",
			ModulePath: "example.com/replayed-demo",
		},
		Git:     protocoltoml.ReplayGit{Mode: domain.GitModeInitOnly, Signoff: true},
		Options: protocoltoml.ReplayOptions{Force: true},
		Inputs: map[string]string{
			"module_path": "example.com/replayed-demo",
			"author":      "from-replay",
		},
	})

	opts, err := svc.BuildNewOptions(NewRequest{
		Flags: Flags{ReplayPath: replayPath, SetValues: []string{"author=from-cli"}},
	})
	if err != nil {
		t.Fatalf("BuildNewOptions() error = %v", err)
	}

	if opts.Lang != "cpp" || opts.ProjectName != "replayed-demo" || opts.TargetDir != "replayed-demo" {
		t.Fatalf("opts = %#v, want replay name/target/lang", opts)
	}
	if opts.ModulePath != "example.com/replayed-demo" {
		t.Fatalf("opts.ModulePath = %q, want replay module", opts.ModulePath)
	}
	if opts.GitMode != domain.GitModeInitOnly || !opts.Signoff || !opts.Force {
		t.Fatalf("opts = %#v, want replay git/signoff/force", opts)
	}
	if got := opts.TemplateInputValues["author"]; got != "from-cli" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want from-cli", got)
	}
	if _, exists := opts.TemplateInputValues["module_path"]; exists {
		t.Fatal("opts.TemplateInputValues[module_path] should be resolved outside template inputs")
	}
}

func TestServiceBuildNewOptions_ExplicitOverridesBeatReplay(t *testing.T) {
	svc := NewService()
	replayPath := writeReplayForCreateServiceTest(t, protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(CommandNew),
		Template: protocoltoml.ReplayTemplate{Lang: "cpp"},
		Project: protocoltoml.ReplayProject{
			Name:       "replay-name",
			TargetDir:  "replay-dir",
			ModulePath: "example.com/from-replay",
		},
		Git:     protocoltoml.ReplayGit{Mode: domain.GitModeInitOnly, Signoff: true},
		Options: protocoltoml.ReplayOptions{Force: false},
		Inputs: map[string]string{
			"module_path": "example.com/from-replay",
			"author":      "from-replay",
		},
	})

	opts, err := svc.BuildNewOptions(NewRequest{
		Flags: Flags{
			Lang:       "go",
			Module:     "example.com/from-cli",
			Signoff:    false,
			GitMode:    "none",
			ReplayPath: replayPath,
			SetValues:  []string{"author=from-cli"},
		},
		Changed: Changed{Lang: true, Module: true, Signoff: true, Git: true, Force: true},
		Force:   true,
		Arg:     "cli-demo",
		HasArg:  true,
	})
	if err != nil {
		t.Fatalf("BuildNewOptions() error = %v", err)
	}

	if opts.Lang != "go" || opts.ProjectName != "cli-demo" || opts.TargetDir != "cli-demo" {
		t.Fatalf("opts = %#v, want CLI lang/project/target", opts)
	}
	if opts.ModulePath != "example.com/from-cli" || opts.GitMode != domain.GitModeNone {
		t.Fatalf("opts = %#v, want CLI module/git", opts)
	}
	if opts.Signoff || !opts.Force {
		t.Fatalf("opts = %#v, want signoff=false force=true", opts)
	}
	if got := opts.TemplateInputValues["author"]; got != "from-cli" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want from-cli", got)
	}
}

func TestServiceBuildNewOptions_ConfigDefaultsApplyWhenFlagsAndArgAreOmitted(t *testing.T) {
	svc := NewService()

	opts, err := svc.BuildNewOptions(NewRequest{
		Flags: Flags{ActiveConfig: activeConfigForCreateServiceTest(protocoltoml.Config{
			Version: protocoltoml.ConfigVersion,
			New: &protocoltoml.ConfigNewSection{
				Lang:        stringPtr("go"),
				ProjectName: stringPtr("demo"),
				Module:      stringPtr("example.com/demo"),
				GitMode:     stringPtr("none"),
				Signoff:     boolPtr(true),
				Inputs: map[string]string{
					"author":     "from-config",
					"go_version": "1.25",
				},
			},
		})},
	})
	if err != nil {
		t.Fatalf("BuildNewOptions() error = %v", err)
	}

	if opts.Lang != "go" || opts.ProjectName != "demo" || opts.TargetDir != "demo" {
		t.Fatalf("opts = %#v, want config lang/project/target", opts)
	}
	if opts.ModulePath != "example.com/demo" || opts.GitMode != domain.GitModeNone {
		t.Fatalf("opts = %#v, want config module/git", opts)
	}
	if !opts.Signoff {
		t.Fatalf("opts.Signoff = %v, want true", opts.Signoff)
	}
	if got := opts.TemplateInputValues["author"]; got != "from-config" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want from-config", got)
	}
	if got := opts.TemplateInputValues["go_version"]; got != "1.25" {
		t.Fatalf("opts.TemplateInputValues[go_version] = %q, want 1.25", got)
	}
}

func TestServiceBuildNewOptions_ExplicitCLIBeatsConfig(t *testing.T) {
	svc := NewService()

	opts, err := svc.BuildNewOptions(NewRequest{
		Flags: Flags{
			Lang:      "go",
			Module:    "example.com/from-cli",
			GitMode:   "none",
			SetValues: []string{"author=from-cli"},
			ActiveConfig: activeConfigForCreateServiceTest(protocoltoml.Config{
				Version: protocoltoml.ConfigVersion,
				New: &protocoltoml.ConfigNewSection{
					Lang:        stringPtr("cpp"),
					ProjectName: stringPtr("config-demo"),
					Module:      stringPtr("example.com/from-config"),
					GitMode:     stringPtr("init-only"),
					Signoff:     boolPtr(true),
					Inputs:      map[string]string{"author": "from-config"},
				},
			}),
		},
		Changed: Changed{Lang: true, Module: true, Signoff: true, Git: true},
		Arg:     "cli-demo",
		HasArg:  true,
	})
	if err != nil {
		t.Fatalf("BuildNewOptions() error = %v", err)
	}

	if opts.Lang != "go" || opts.ProjectName != "cli-demo" || opts.TargetDir != "cli-demo" {
		t.Fatalf("opts = %#v, want CLI lang/project/target", opts)
	}
	if opts.ModulePath != "example.com/from-cli" || opts.GitMode != domain.GitModeNone {
		t.Fatalf("opts = %#v, want CLI module/git", opts)
	}
	if opts.Signoff {
		t.Fatalf("opts.Signoff = %v, want false", opts.Signoff)
	}
	if got := opts.TemplateInputValues["author"]; got != "from-cli" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want from-cli", got)
	}
}

func TestServiceBuildInitOptions_ExplicitOverridesReplayTargetAndOptions(t *testing.T) {
	svc := NewService()
	replayPath := writeReplayForCreateServiceTest(t, protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(CommandInit),
		Template: protocoltoml.ReplayTemplate{Lang: "cpp"},
		Project: protocoltoml.ReplayProject{
			Name:       "replay-name",
			TargetDir:  "replay-dir",
			ModulePath: "example.com/from-replay",
		},
		Git: protocoltoml.ReplayGit{Mode: domain.GitModeInitOnly, Signoff: true},
		Inputs: map[string]string{
			"module_path": "example.com/from-replay",
			"author":      "from-replay",
		},
	})

	targetDir := filepath.Join("nested", "demo")
	opts, err := svc.BuildInitOptions(InitRequest{
		Flags: Flags{
			Lang:       "go",
			Module:     "example.com/from-cli",
			Signoff:    false,
			GitMode:    "none",
			ReplayPath: replayPath,
			SetValues:  []string{"author=from-cli"},
		},
		Changed: Changed{Lang: true, Module: true, Signoff: true, Git: true},
		Arg:     targetDir,
		HasArg:  true,
	})
	if err != nil {
		t.Fatalf("BuildInitOptions() error = %v", err)
	}

	if opts.Lang != "go" || opts.ProjectName != "demo" || opts.TargetDir != targetDir {
		t.Fatalf("opts = %#v, want init target derived from CLI", opts)
	}
	if opts.ModulePath != "example.com/from-cli" || opts.GitMode != domain.GitModeNone {
		t.Fatalf("opts = %#v, want CLI module/git override", opts)
	}
	if opts.Signoff || !opts.AllowExistingEmptyDir {
		t.Fatalf("opts = %#v, want signoff=false allowExistingEmptyDir=true", opts)
	}
	if got := opts.TemplateInputValues["author"]; got != "from-cli" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want from-cli", got)
	}
}

func TestServiceBuildInitOptions_ConfigTargetAppliesWhenArgIsOmitted(t *testing.T) {
	svc := NewService()
	targetDir := filepath.Join("nested", "demo")

	opts, err := svc.BuildInitOptions(InitRequest{
		Flags: Flags{ActiveConfig: activeConfigForCreateServiceTest(protocoltoml.Config{
			Version: protocoltoml.ConfigVersion,
			Init: &protocoltoml.ConfigInitSection{
				Lang:      stringPtr("go"),
				TargetDir: stringPtr(targetDir),
				Module:    stringPtr("example.com/demo"),
				GitMode:   stringPtr("none"),
				Inputs:    map[string]string{"author": "from-config"},
			},
		})},
	})
	if err != nil {
		t.Fatalf("BuildInitOptions() error = %v", err)
	}

	if opts.Lang != "go" || opts.ProjectName != "demo" || opts.TargetDir != targetDir {
		t.Fatalf("opts = %#v, want config init target", opts)
	}
	if opts.ModulePath != "example.com/demo" || opts.GitMode != domain.GitModeNone {
		t.Fatalf("opts = %#v, want config module/git", opts)
	}
	if got := opts.TemplateInputValues["author"]; got != "from-config" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want from-config", got)
	}
	if !opts.AllowExistingEmptyDir {
		t.Fatal("opts.AllowExistingEmptyDir = false, want true")
	}
}

func TestServiceBuildInitOptions_ExplicitCLIBeatsConfig(t *testing.T) {
	svc := NewService()
	targetDir := filepath.Join("nested", "demo")

	opts, err := svc.BuildInitOptions(InitRequest{
		Flags: Flags{
			Lang:      "go",
			Module:    "example.com/from-cli",
			GitMode:   "none",
			SetValues: []string{"author=from-cli"},
			ActiveConfig: activeConfigForCreateServiceTest(protocoltoml.Config{
				Version: protocoltoml.ConfigVersion,
				Init: &protocoltoml.ConfigInitSection{
					Lang:      stringPtr("cpp"),
					TargetDir: stringPtr("config-dir"),
					Module:    stringPtr("example.com/from-config"),
					GitMode:   stringPtr("init-only"),
					Signoff:   boolPtr(true),
					Inputs:    map[string]string{"author": "from-config"},
				},
			}),
		},
		Changed: Changed{Lang: true, Module: true, Signoff: true, Git: true},
		Arg:     targetDir,
		HasArg:  true,
	})
	if err != nil {
		t.Fatalf("BuildInitOptions() error = %v", err)
	}

	if opts.Lang != "go" || opts.ProjectName != "demo" || opts.TargetDir != targetDir {
		t.Fatalf("opts = %#v, want CLI init target", opts)
	}
	if opts.ModulePath != "example.com/from-cli" || opts.GitMode != domain.GitModeNone {
		t.Fatalf("opts = %#v, want CLI module/git", opts)
	}
	if opts.Signoff {
		t.Fatalf("opts.Signoff = %v, want false", opts.Signoff)
	}
	if got := opts.TemplateInputValues["author"]; got != "from-cli" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want from-cli", got)
	}
}
