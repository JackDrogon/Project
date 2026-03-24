package catalog

import (
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestServiceInspect(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":        {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = true\n")},
		"go/.editorconfig":                          {Data: []byte("root = true\n")},
		"go/.github/workflows/ci.yml":               {Data: []byte("name: CI\n")},
		"go/.github/dependabot.yml":                 {Data: []byte("version: 2\n")},
		"go/.gitignore":                             {Data: []byte("bin/\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
		"go/typos.toml":                             {Data: []byte("[files]\nextend-exclude = [\"bin\"]\n")},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: fs.ModeDir | 0o755},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
	}
	svc := NewService(fsys, nil)

	got, err := svc.Inspect("go", InspectModeRender)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if got.Name != "go" || got.Description != "Go starter" || got.ManifestVersion != 2 {
		t.Fatalf("Inspect() summary = %#v, want manifest metadata preserved", got)
	}
	if got.Mode != InspectModeRender || got.ShownCount() != 2 {
		t.Fatalf("Inspect() mode/shown = (%q, %d), want (%q, %d)", got.Mode, got.ShownCount(), InspectModeRender, 2)
	}
	if !reflect.DeepEqual(inputNames(got.Inputs), []string{"module_path", "go_version"}) {
		t.Fatalf("Inspect() inputs = %#v, want module_path/go_version", got.Inputs)
	}
	if len(got.Files) != 2 || got.Files[0].Source != filepathToSlash("cmd/{{.ProjectNameLower}}/main.go.tmpl") || got.Files[1].Source != "go.mod.tmpl" {
		t.Fatalf("Inspect() files = %#v, want only rendered files", got.Files)
	}
	if !reflect.DeepEqual(got.Variables, []string{"ModulePath", "ProjectNameLower"}) {
		t.Fatalf("Inspect() variables = %v, want ModulePath and ProjectNameLower", got.Variables)
	}
}

func TestServiceInspect_Errors(t *testing.T) {
	svc := NewService(fstest.MapFS{}, nil)

	t.Run("unsupported language", func(t *testing.T) {
		_, err := svc.Inspect("missing", InspectModeAll)
		if err == nil {
			t.Fatal("Inspect() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported language") {
			t.Fatalf("Inspect() error = %v, want unsupported language error", err)
		}
	})

	t.Run("invalid mode", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\n")},
		}
		svc := NewService(fsys, nil)
		_, err := svc.Inspect("go", InspectMode("bogus"))
		if err == nil {
			t.Fatal("Inspect() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --mode") {
			t.Fatalf("Inspect() error = %v, want mode error", err)
		}
	})
}

func TestServiceQueryInspection(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":        {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = true\n")},
		"go/.gitignore":                             {Data: []byte("bin/\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: fs.ModeDir | 0o755},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
	}
	svc := NewService(fsys, nil)

	t.Run("render mode query", func(t *testing.T) {
		got, err := svc.QueryInspection(InspectionQuery{Lang: "go", Mode: InspectModeRender})
		if err != nil {
			t.Fatalf("QueryInspection() error = %v", err)
		}
		if got.Mode != InspectModeRender || got.ShownCount() != 2 {
			t.Fatalf("QueryInspection() mode/shown = (%q, %d), want (%q, %d)", got.Mode, got.ShownCount(), InspectModeRender, 2)
		}
	})

	t.Run("query validation", func(t *testing.T) {
		_, err := svc.QueryInspection(InspectionQuery{Lang: "go", Mode: InspectMode("bogus")})
		if err == nil {
			t.Fatal("QueryInspection() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --mode") {
			t.Fatalf("QueryInspection() error = %v, want mode validation error", err)
		}
	})
}
