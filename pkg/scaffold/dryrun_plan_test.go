package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBuildDryRunPlan_ClassifiesCreateRenderAndCopyActions(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.json":        {Data: []byte(`{"schema_version":1,"name":"go","description":"Go starter","inputs":[{"name":"module_path","template_var":"ModulePath"},{"name":"go_version","template_var":"GoVersion"},{"name":"author","template_var":"Author"},{"name":"year","template_var":"Year"}]}`)},
		"go/cmd":                                    {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
		"go/README.md":                              {Data: []byte("# {{.ProjectName}}\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
	}

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		t.Fatalf("git runner should not be called while building dry-run plan")
		return nil
	})

	plan, err := creator.BuildDryRunPlan(Options{
		Lang:        "go",
		ProjectName: "Demo",
		ModulePath:  "example.com/demo",
	})
	if err != nil {
		t.Fatalf("BuildDryRunPlan() error = %v", err)
	}

	if plan.Template != "go" {
		t.Fatalf("plan.Template = %q, want %q", plan.Template, "go")
	}
	if plan.Description != "Go starter" {
		t.Fatalf("plan.Description = %q, want %q", plan.Description, "Go starter")
	}
	if plan.TargetDir != "Demo" {
		t.Fatalf("plan.TargetDir = %q, want %q", plan.TargetDir, "Demo")
	}

	if len(plan.ResolvedInputs) != 4 {
		t.Fatalf("len(plan.ResolvedInputs) = %d, want 4", len(plan.ResolvedInputs))
	}

	if got, want := plan.ResolvedInputs[0], (DryRunResolvedInput{Name: "module_path", TemplateVar: "ModulePath", Value: "example.com/demo"}); got != want {
		t.Fatalf("plan.ResolvedInputs[0] = %#v, want %#v", got, want)
	}
	if got := plan.ResolvedInputs[1]; got.Name != "go_version" || got.TemplateVar != "GoVersion" || got.Value == "" {
		t.Fatalf("plan.ResolvedInputs[1] = %#v, want non-empty GoVersion with name/template_var preserved", got)
	}
	if got := plan.ResolvedInputs[2]; got.Name != "author" || got.TemplateVar != "Author" || got.Value == "" {
		t.Fatalf("plan.ResolvedInputs[2] = %#v, want non-empty Author with name/template_var preserved", got)
	}
	if got := plan.ResolvedInputs[3]; got.Name != "year" || got.TemplateVar != "Year" || got.Value == "" {
		t.Fatalf("plan.ResolvedInputs[3] = %#v, want non-empty Year with name/template_var preserved", got)
	}

	if len(plan.Actions) != 5 {
		t.Fatalf("len(plan.Actions) = %d, want 5", len(plan.Actions))
	}

	gotActions := []DryRunAction{
		{Kind: plan.Actions[0].Kind, Source: plan.Actions[0].Source, Target: plan.Actions[0].Target},
		{Kind: plan.Actions[1].Kind, Source: plan.Actions[1].Source, Target: plan.Actions[1].Target},
		{Kind: plan.Actions[2].Kind, Source: plan.Actions[2].Source, Target: plan.Actions[2].Target},
		{Kind: plan.Actions[3].Kind, Source: plan.Actions[3].Source, Target: plan.Actions[3].Target},
		{Kind: plan.Actions[4].Kind, Source: plan.Actions[4].Source, Target: plan.Actions[4].Target},
	}
	wantActions := []DryRunAction{
		{Kind: DryRunActionCopyFile, Source: "go/README.md", Target: filepath.Join("Demo", "README.md")},
		{Kind: DryRunActionCreateDir, Target: filepath.Join("Demo", "cmd")},
		{Kind: DryRunActionCreateDir, Target: filepath.Join("Demo", "cmd", "demo")},
		{Kind: DryRunActionRenderFile, Source: "go/cmd/{{.ProjectNameLower}}/main.go.tmpl", Target: filepath.Join("Demo", "cmd", "demo", "main.go")},
		{Kind: DryRunActionRenderFile, Source: "go/go.mod.tmpl", Target: filepath.Join("Demo", "go.mod")},
	}
	if !reflect.DeepEqual(gotActions, wantActions) {
		t.Fatalf("plan.Actions = %#v, want %#v", gotActions, wantActions)
	}
}

func TestBuildDryRunPlan_PreservesNoWriteAndNoGitGuarantees(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.json": {Data: []byte(`{"schema_version":1,"name":"go","description":"Go starter","inputs":[{"name":"module_path","template_var":"ModulePath"}]}`)},
		"go/main.go.tmpl":                    {Data: []byte("package main\n")},
	}

	originalRemoveAll := osRemoveAll
	removeAllCalled := false
	osRemoveAll = func(path string) error {
		removeAllCalled = true
		return nil
	}
	t.Cleanup(func() {
		osRemoveAll = originalRemoveAll
	})

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		t.Fatalf("git runner should not be called while building dry-run plan")
		return nil
	})

	tmp := withTempWorkingDir(t)
	targetDir := filepath.Join(tmp, "demo")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", targetDir, err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "keep.txt"), []byte("keep"), 0644); err != nil {
		t.Fatalf("WriteFile(keep.txt) error = %v", err)
	}

	plan, err := creator.BuildDryRunPlan(Options{
		Lang:        "go",
		ProjectName: "demo",
		TargetDir:   targetDir,
		ModulePath:  "example.com/demo",
		DryRun:      true,
		Force:       true,
	})
	if err != nil {
		t.Fatalf("BuildDryRunPlan() error = %v", err)
	}

	if plan.TargetDir != targetDir {
		t.Fatalf("plan.TargetDir = %q, want %q", plan.TargetDir, targetDir)
	}
	if removeAllCalled {
		t.Fatal("BuildDryRunPlan() should not remove existing directories")
	}
	if _, err := os.Stat(filepath.Join(targetDir, "keep.txt")); err != nil {
		t.Fatalf("keep.txt should remain after plan build, stat error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "main.go")); !os.IsNotExist(err) {
		t.Fatalf("BuildDryRunPlan() should not create rendered files, stat err = %v", err)
	}
	if len(plan.Actions) != 1 || plan.Actions[0].Kind != DryRunActionRenderFile {
		t.Fatalf("plan.Actions = %#v, want one render action", plan.Actions)
	}
}

func TestBuildDryRunPlan_FailsOnInvalidTemplatesAndRenderedPaths(t *testing.T) {
	t.Run("invalid template content", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go/bad.txt.tmpl": {Data: []byte("{{.ProjectName")},
		}

		creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
			t.Fatalf("git runner should not be called while building dry-run plan")
			return nil
		})

		withTempWorkingDir(t)
		_, err := creator.BuildDryRunPlan(Options{Lang: "go", ProjectName: "demo", DryRun: true})
		if err == nil {
			t.Fatal("BuildDryRunPlan() expected invalid template error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to render template") {
			t.Fatalf("BuildDryRunPlan() error = %v, want rendered template failure", err)
		}
	})

	t.Run("invalid rendered path", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go/{{.ModulePath}}.txt": {Data: []byte("content")},
		}

		creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
			t.Fatalf("git runner should not be called while building dry-run plan")
			return nil
		})

		withTempWorkingDir(t)
		_, err := creator.BuildDryRunPlan(Options{Lang: "go", ProjectName: "demo", ModulePath: "acme/demo", DryRun: true})
		if err == nil {
			t.Fatal("BuildDryRunPlan() expected rendered path error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to render template path") {
			t.Fatalf("BuildDryRunPlan() error = %v, want wrapped path rendering error", err)
		}
	})
}
