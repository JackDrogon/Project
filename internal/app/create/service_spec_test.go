package create

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JackDrogon/project/internal/scaffold"
)

func TestServiceBuildNewSpec(t *testing.T) {
	svc := NewService()
	spec, err := svc.BuildNewSpec(NewRequest{
		Flags:   Flags{Lang: "go", GitMode: "none"},
		Changed: Changed{Lang: true, Git: true},
		Arg:     "demo",
		HasArg:  true,
	})
	if err != nil {
		t.Fatalf("BuildNewSpec() error = %v", err)
	}
	if spec.Command != CommandNew || spec.Options.ProjectName != "demo" || spec.Options.TargetDir != "demo" {
		t.Fatalf("spec = %#v, want new/demo", spec)
	}
}

func TestServiceBuildInitSpec(t *testing.T) {
	svc := NewService()
	spec, err := svc.BuildInitSpec(InitRequest{
		Flags:   Flags{Lang: "go", GitMode: "none"},
		Changed: Changed{Lang: true, Git: true},
		Arg:     "nested/demo",
		HasArg:  true,
	})
	if err != nil {
		t.Fatalf("BuildInitSpec() error = %v", err)
	}
	if spec.Command != CommandInit || spec.Options.TargetDir != "nested/demo" || spec.Options.ProjectName != "demo" {
		t.Fatalf("spec = %#v, want init nested/demo", spec)
	}
	if !spec.Options.AllowExistingEmptyDir {
		t.Fatal("spec.Options.AllowExistingEmptyDir = false, want true")
	}
}

func TestServiceExecuteScaffoldSpec(t *testing.T) {
	svc := NewService()
	var out bytes.Buffer
	creator := NewCreator(fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\n")},
		"go/.gitignore":                      {Data: []byte("bin/\n")},
		"go/main.go.tmpl":                    {Data: []byte("package main\n")},
	}, &out)
	spec := ScaffoldSpec{
		Command: CommandNew,
		Flags:   Flags{DryRun: true},
		Options: Options{Lang: "go", ProjectName: "demo", TargetDir: "demo", GitMode: scaffold.GitModeNone, DryRun: true},
	}

	if err := svc.ExecuteScaffoldSpec(t.Context(), creator, spec); err != nil {
		t.Fatalf("ExecuteScaffoldSpec() error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, "Dry-run mode: no files will be created") {
		t.Fatalf("ExecuteScaffoldSpec() output = %q, want dry-run banner", got)
	}
}

func TestServiceExecuteScaffoldSpec_PropagatesCreateError(t *testing.T) {
	svc := NewService()
	creator := NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	spec := ScaffoldSpec{Command: CommandNew, Flags: Flags{}, Options: Options{Lang: "missing", ProjectName: "demo", TargetDir: "demo", GitMode: scaffold.GitModeNone}}

	err := svc.ExecuteScaffoldSpec(t.Context(), creator, spec)
	if err == nil {
		t.Fatal("ExecuteScaffoldSpec() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported language") {
		t.Fatalf("ExecuteScaffoldSpec() error = %v, want unsupported language error", err)
	}
}
