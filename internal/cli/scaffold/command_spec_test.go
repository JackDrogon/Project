package scaffold

import (
	"bytes"
	"testing"
	"testing/fstest"

	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func TestNewScaffoldCommandSpecBuilder(t *testing.T) {
	service := appcreate.NewService()
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := NewNewCommand(newTestDependencies(creator))
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "demo"})
	if err := cmd.ParseFlags([]string{"--lang", "go", "--git", "none"}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	flags := scaffoldCommandFlags{lang: "go", gitMode: "none"}
	force := false
	builder := newScaffoldCommandSpecBuilder{flags: &flags, force: &force}
	spec, err := builder.Build(service, cmd, newTestDependencies(creator), []string{"demo"})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if spec.Command != appcreate.CommandNew || spec.Options.ProjectName != "demo" {
		t.Fatalf("spec = %#v, want new/demo", spec)
	}
}

func TestInitScaffoldCommandSpecBuilder(t *testing.T) {
	service := appcreate.NewService()
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := NewInitCommand(newTestDependencies(creator))
	if err := cmd.ParseFlags([]string{"--lang", "go", "--git", "none"}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	flags := scaffoldCommandFlags{lang: "go", gitMode: "none"}
	builder := initScaffoldCommandSpecBuilder{flags: &flags}
	spec, err := builder.Build(service, cmd, newTestDependencies(creator), []string{"nested/demo"})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if spec.Command != appcreate.CommandInit || spec.Options.ProjectName != "demo" || spec.Options.TargetDir != "nested/demo" {
		t.Fatalf("spec = %#v, want init nested/demo", spec)
	}
}
