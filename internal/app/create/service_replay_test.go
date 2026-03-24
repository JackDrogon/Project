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
