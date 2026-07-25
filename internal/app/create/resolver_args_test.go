package create

import (
	"os"
	"strings"
	"testing"
)

func TestResolveNewProjectArgs_GoModuleVersionHeuristic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		arg             string
		wantProjectName string
		wantTargetDir   string
		wantModulePath  string
	}{
		{
			name:            "plain module path uses final segment",
			arg:             "github.com/acme/agent-village",
			wantProjectName: "agent-village",
			wantTargetDir:   "agent-village",
			wantModulePath:  "github.com/acme/agent-village",
		},
		{
			name:            "major version suffix uses repository segment",
			arg:             "github.com/acme/agent-village/v2",
			wantProjectName: "agent-village",
			wantTargetDir:   "agent-village",
			wantModulePath:  "github.com/acme/agent-village/v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectName, targetDir, modulePath, err := resolveNewProjectArgs("go", "", tt.arg)
			if err != nil {
				t.Fatalf("resolveNewProjectArgs() error = %v", err)
			}
			if projectName != tt.wantProjectName {
				t.Fatalf("projectName = %q, want %q", projectName, tt.wantProjectName)
			}
			if targetDir != tt.wantTargetDir {
				t.Fatalf("targetDir = %q, want %q", targetDir, tt.wantTargetDir)
			}
			if modulePath != tt.wantModulePath {
				t.Fatalf("modulePath = %q, want %q", modulePath, tt.wantModulePath)
			}
		})
	}
}

func TestResolveNewProjectArgs_FallbacksAndErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		lang            string
		module          string
		arg             string
		wantProjectName string
		wantTargetDir   string
		wantModulePath  string
		wantErr         string
	}{
		{
			name:            "non-go language uses literal arg",
			lang:            "cpp",
			arg:             "github.com/acme/agent-village",
			wantProjectName: "github.com/acme/agent-village",
			wantTargetDir:   "github.com/acme/agent-village",
			wantModulePath:  "",
		},
		{
			name:            "rust also uses literal arg",
			lang:            "rust",
			arg:             "github.com/acme/agent-village",
			wantProjectName: "github.com/acme/agent-village",
			wantTargetDir:   "github.com/acme/agent-village",
			wantModulePath:  "",
		},
		{
			name:            "explicit module bypasses go module heuristic",
			lang:            "go",
			module:          "example.com/custom/module",
			arg:             "github.com/acme/agent-village",
			wantProjectName: "github.com/acme/agent-village",
			wantTargetDir:   "github.com/acme/agent-village",
			wantModulePath:  "example.com/custom/module",
		},
		{
			name:    "derived repository name must also be a valid project name",
			lang:    "go",
			arg:     "github.com/acme/9agent",
			wantErr: "project name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectName, targetDir, modulePath, err := resolveNewProjectArgs(tt.lang, tt.module, tt.arg)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("resolveNewProjectArgs() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("resolveNewProjectArgs() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("resolveNewProjectArgs() error = %v", err)
			}
			if projectName != tt.wantProjectName {
				t.Fatalf("projectName = %q, want %q", projectName, tt.wantProjectName)
			}
			if targetDir != tt.wantTargetDir {
				t.Fatalf("targetDir = %q, want %q", targetDir, tt.wantTargetDir)
			}
			if modulePath != tt.wantModulePath {
				t.Fatalf("modulePath = %q, want %q", modulePath, tt.wantModulePath)
			}
		})
	}
}

// Not parallel: it removes the process working directory, which is global state.
func TestProjectNameFromTargetDir_FailsWhenCWDIsMissing(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir(%q) error = %v", tmp, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
	if err := os.RemoveAll(tmp); err != nil {
		t.Fatalf("RemoveAll(%q) error = %v", tmp, err)
	}

	if _, err := projectNameFromTargetDir("."); err == nil {
		t.Fatal("projectNameFromTargetDir() expected error, got nil")
	} else if !strings.Contains(err.Error(), "failed to resolve target directory") {
		t.Fatalf("projectNameFromTargetDir() error = %v, want resolution error", err)
	}
}
