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
	replayPath := writeReplayForServiceTest(t, protocoltoml.Replay{
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
		Flags: Flags{
			ReplayPath: replayPath,
			SetValues:  []string{"author=from-cli"},
		},
	})
	if err != nil {
		t.Fatalf("BuildNewOptions() error = %v", err)
	}

	if opts.Lang != "cpp" {
		t.Fatalf("opts.Lang = %q, want %q", opts.Lang, "cpp")
	}
	if opts.ProjectName != "replayed-demo" {
		t.Fatalf("opts.ProjectName = %q, want %q", opts.ProjectName, "replayed-demo")
	}
	if opts.TargetDir != "replayed-demo" {
		t.Fatalf("opts.TargetDir = %q, want %q", opts.TargetDir, "replayed-demo")
	}
	if opts.ModulePath != "example.com/replayed-demo" {
		t.Fatalf("opts.ModulePath = %q, want %q", opts.ModulePath, "example.com/replayed-demo")
	}
	if opts.GitMode != domain.GitModeInitOnly {
		t.Fatalf("opts.GitMode = %q, want %q", opts.GitMode, domain.GitModeInitOnly)
	}
	if !opts.Signoff {
		t.Fatal("opts.Signoff = false, want true")
	}
	if !opts.Force {
		t.Fatal("opts.Force = false, want true")
	}
	if got := opts.TemplateInputValues["author"]; got != "from-cli" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want %q", got, "from-cli")
	}
	if _, exists := opts.TemplateInputValues["module_path"]; exists {
		t.Fatal("opts.TemplateInputValues[module_path] should be resolved outside template inputs")
	}
}

func TestServiceBuildNewOptions_ExplicitOverridesBeatReplay(t *testing.T) {
	svc := NewService()
	replayPath := writeReplayForServiceTest(t, protocoltoml.Replay{
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
		Changed: Changed{
			Lang:    true,
			Module:  true,
			Signoff: true,
			Git:     true,
			Force:   true,
		},
		Force:  true,
		Arg:    "cli-demo",
		HasArg: true,
	})
	if err != nil {
		t.Fatalf("BuildNewOptions() error = %v", err)
	}

	if opts.Lang != "go" {
		t.Fatalf("opts.Lang = %q, want %q", opts.Lang, "go")
	}
	if opts.ProjectName != "cli-demo" {
		t.Fatalf("opts.ProjectName = %q, want %q", opts.ProjectName, "cli-demo")
	}
	if opts.TargetDir != "cli-demo" {
		t.Fatalf("opts.TargetDir = %q, want %q", opts.TargetDir, "cli-demo")
	}
	if opts.ModulePath != "example.com/from-cli" {
		t.Fatalf("opts.ModulePath = %q, want %q", opts.ModulePath, "example.com/from-cli")
	}
	if opts.GitMode != domain.GitModeNone {
		t.Fatalf("opts.GitMode = %q, want %q", opts.GitMode, domain.GitModeNone)
	}
	if opts.Signoff {
		t.Fatal("opts.Signoff = true, want false")
	}
	if !opts.Force {
		t.Fatal("opts.Force = false, want true")
	}
	if got := opts.TemplateInputValues["author"]; got != "from-cli" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want %q", got, "from-cli")
	}
}

func TestServiceBuildInitOptions_ExplicitOverridesReplayTargetAndOptions(t *testing.T) {
	svc := NewService()
	replayPath := writeReplayForServiceTest(t, protocoltoml.Replay{
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
		Changed: Changed{
			Lang:    true,
			Module:  true,
			Signoff: true,
			Git:     true,
		},
		Arg:    targetDir,
		HasArg: true,
	})
	if err != nil {
		t.Fatalf("BuildInitOptions() error = %v", err)
	}

	if opts.Lang != "go" {
		t.Fatalf("opts.Lang = %q, want %q", opts.Lang, "go")
	}
	if opts.ProjectName != "demo" {
		t.Fatalf("opts.ProjectName = %q, want %q", opts.ProjectName, "demo")
	}
	if opts.TargetDir != targetDir {
		t.Fatalf("opts.TargetDir = %q, want %q", opts.TargetDir, targetDir)
	}
	if opts.ModulePath != "example.com/from-cli" {
		t.Fatalf("opts.ModulePath = %q, want %q", opts.ModulePath, "example.com/from-cli")
	}
	if opts.GitMode != domain.GitModeNone {
		t.Fatalf("opts.GitMode = %q, want %q", opts.GitMode, domain.GitModeNone)
	}
	if opts.Signoff {
		t.Fatal("opts.Signoff = true, want false")
	}
	if !opts.AllowExistingEmptyDir {
		t.Fatal("opts.AllowExistingEmptyDir = false, want true")
	}
	if got := opts.TemplateInputValues["author"]; got != "from-cli" {
		t.Fatalf("opts.TemplateInputValues[author] = %q, want %q", got, "from-cli")
	}
}

func writeReplayForServiceTest(t *testing.T, replay protocoltoml.Replay) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "replay.toml")
	if err := protocoltoml.WriteReplay(path, replay); err != nil {
		t.Fatalf("WriteReplay(%q) error = %v", path, err)
	}

	return path
}
