package protocoltoml

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

func TestReplayV2_ReadWriteRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "replay.toml")
	want := Replay{
		Version:  ReplayVersion,
		Mode:     "new",
		Template: ReplayTemplate{Lang: "go"},
		Project:  ReplayProject{Name: "demo", TargetDir: "demo", ModulePath: "example.com/demo"},
		Git:      ReplayGit{Mode: domain.GitModeInitCommit, Signoff: true},
		Options:  ReplayOptions{Force: true},
		Inputs: map[string]string{
			"author":      "alice",
			"module_path": "example.com/demo",
		},
	}

	if err := WriteReplay(path, want); err != nil {
		t.Fatalf("WriteReplay() error = %v", err)
	}

	got, err := ReadReplay(path)
	if err != nil {
		t.Fatalf("ReadReplay() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadReplay() = %#v, want %#v", got, want)
	}
}

func TestReplayV2_RejectsLegacyJSON(t *testing.T) {
	_, err := DecodeReplay([]byte(`{"schema_version":1,"command":"new"}`), "replay.toml")
	if err == nil {
		t.Fatal("DecodeReplay() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "legacy JSON") {
		t.Fatalf("DecodeReplay() error = %v, want legacy JSON rejection", err)
	}
}

func TestReplayV2_RejectsUnknownFieldsAndInvalidValues(t *testing.T) {
	t.Run("unknown field", func(t *testing.T) {
		content := []byte("version = 2\nmode = \"new\"\nunknown = true\n\n[template]\nlang = \"go\"\n\n[project]\nname = \"demo\"\ntarget_dir = \"demo\"\n\n[git]\nmode = \"none\"\nsignoff = false\n\n[options]\nforce = false\n\n[inputs]\n")
		_, err := DecodeReplay(content, "replay.toml")
		if err == nil {
			t.Fatal("DecodeReplay() expected error, got nil")
		}
	})

	t.Run("invalid mode", func(t *testing.T) {
		content := []byte("version = 2\nmode = \"list\"\n\n[template]\nlang = \"go\"\n\n[project]\nname = \"demo\"\ntarget_dir = \"demo\"\n\n[git]\nmode = \"none\"\nsignoff = false\n\n[options]\nforce = false\n\n[inputs]\n")
		_, err := DecodeReplay(content, "replay.toml")
		if err == nil {
			t.Fatal("DecodeReplay() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported mode") {
			t.Fatalf("DecodeReplay() error = %v, want mode error", err)
		}
	})
}

func TestReplayV2_WritesTOMLContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "replay.toml")
	replay := Replay{
		Mode:     "init",
		Template: ReplayTemplate{Lang: "go"},
		Project:  ReplayProject{Name: "demo", TargetDir: ".", ModulePath: "example.com/demo"},
		Git:      ReplayGit{Mode: domain.GitModeInitOnly, Signoff: false},
		Options:  ReplayOptions{Force: false},
		Inputs:   map[string]string{"author": "alice"},
	}

	if err := WriteReplay(path, replay); err != nil {
		t.Fatalf("WriteReplay() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(content)
	for _, fragment := range []string{"version = 2", "mode = 'init'", "lang = 'go'", "target_dir = '.'", "mode = 'init-only'"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("replay TOML = %q, want contains %q", got, fragment)
		}
	}
}
