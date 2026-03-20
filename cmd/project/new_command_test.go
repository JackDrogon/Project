package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JackDrogon/project/pkg/scaffold"
)

func TestNewCmd_RequiresLang(t *testing.T) {
	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"demo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "required flag") {
		t.Fatalf("Execute() error = %v, want required flag error", err)
	}
}

func TestNewCmd_RejectsInvalidArgCount(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing project name", args: []string{"--lang", "go"}},
		{name: "too many args", args: []string{"--lang", "go", "one", "two"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
			cmd := newNewCmd(creator)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() expected error, got nil")
			}
			if !strings.Contains(err.Error(), "accepts 1 arg") {
				t.Fatalf("Execute() error = %v, want arg count error", err)
			}
		})
	}
}

func TestNewCmd_CreatesProjectFromArgument(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--module", "example.com/demo", "demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(workDir, "demo", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if string(goMod) != "module example.com/demo\n" {
		t.Fatalf("go.mod = %q, want %q", string(goMod), "module example.com/demo\n")
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, "demo", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "demo"`) {
		t.Fatalf("main.go content = %q, want rendered project name", string(mainGo))
	}
}

func TestNewCmd_GoModuleArgumentDerivesProjectNameAndModulePath(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "github.com/JackDrogon/agent-village"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(workDir, "agent-village", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if string(goMod) != "module github.com/JackDrogon/agent-village\n" {
		t.Fatalf("go.mod = %q, want %q", string(goMod), "module github.com/JackDrogon/agent-village\n")
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, "agent-village", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "agent-village"`) {
		t.Fatalf("main.go content = %q, want rendered project name", string(mainGo))
	}
}

func TestNewCmd_GoModuleArgumentWithMajorVersionSuffixUsesRepositoryName(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "github.com/acme/agent-village/v2"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(workDir, "agent-village", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if string(goMod) != "module github.com/acme/agent-village/v2\n" {
		t.Fatalf("go.mod = %q, want %q", string(goMod), "module github.com/acme/agent-village/v2\n")
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, "agent-village", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "agent-village"`) {
		t.Fatalf("main.go content = %q, want rendered repository name", string(mainGo))
	}

	if _, err := os.Stat(filepath.Join(workDir, "v2")); !os.IsNotExist(err) {
		t.Fatalf("version suffix directory should not be created, stat err = %v", err)
	}
}

func TestResolveNewProjectArgs_GoModuleVersionHeuristic(t *testing.T) {
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
