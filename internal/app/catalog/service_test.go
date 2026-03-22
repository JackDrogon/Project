package catalog

import (
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

type failingFS struct{ err error }

func (f failingFS) Open(string) (fs.File, error)          { return nil, f.err }
func (f failingFS) ReadDir(string) ([]fs.DirEntry, error) { return nil, f.err }

func TestServiceListLangs(t *testing.T) {
	svc := NewService(fstest.MapFS{
		"go/.project-template-manifest.toml":  {Data: []byte("version = 2\nname = \"go\"\n")},
		"cpp/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"cpp\"\n")},
	}, nil)

	got, err := svc.ListLangs()
	if err != nil {
		t.Fatalf("ListLangs() error = %v", err)
	}
	if !reflect.DeepEqual(got, []string{"cpp", "go"}) {
		t.Fatalf("ListLangs() = %v, want [cpp go]", got)
	}
}

func TestServiceListLangs_PropagatesReadErrors(t *testing.T) {
	svc := NewService(failingFS{err: errors.New("boom")}, nil)
	_, err := svc.ListLangs()
	if err == nil {
		t.Fatal("ListLangs() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read templates") {
		t.Fatalf("ListLangs() error = %v, want template read error", err)
	}
}

func TestServiceInspect(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":        {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = true\n")},
		"go/.gitignore":                             {Data: []byte("bin/\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
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
	if got.Mode != InspectModeRender || got.ShownCount != 2 {
		t.Fatalf("Inspect() mode/shown = (%q, %d), want (%q, %d)", got.Mode, got.ShownCount, InspectModeRender, 2)
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
		_, err := svc.Inspect("go", "bogus")
		if err == nil {
			t.Fatal("Inspect() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --mode") {
			t.Fatalf("Inspect() error = %v, want mode error", err)
		}
	})
}

func TestServiceListSummaries(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":  {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n")},
		"go/go.mod.tmpl":                      {Data: []byte("module {{.ModulePath}}\n")},
		"cpp/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"cpp\"\ndescription = \"C++ starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"cpp/README.md.tmpl":                  {Data: []byte("By {{.Author}}\n")},
	}
	svc := NewService(fsys, nil)

	got, err := svc.ListSummaries()
	if err != nil {
		t.Fatalf("ListSummaries() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListSummaries() len = %d, want 2", len(got))
	}
	if got[0].Name != "cpp" || got[1].Name != "go" {
		t.Fatalf("ListSummaries() names = %#v, want sorted [cpp go]", got)
	}
}

func filepathToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
