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
		"go/.project-template-manifest.toml":   {Data: []byte("version = 2\nname = \"go\"\n")},
		"cpp/.project-template-manifest.toml":  {Data: []byte("version = 2\nname = \"cpp\"\n")},
		"rust/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"rust\"\n")},
	}, nil)

	got, err := svc.ListLangs()
	if err != nil {
		t.Fatalf("ListLangs() error = %v", err)
	}
	if !reflect.DeepEqual(got, []string{"cpp", "go", "rust"}) {
		t.Fatalf("ListLangs() = %v, want [cpp go rust]", got)
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
		"go/.project-template-manifest.toml":   {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n")},
		"go/go.mod.tmpl":                       {Data: []byte("module {{.ModulePath}}\n")},
		"cpp/.project-template-manifest.toml":  {Data: []byte("version = 2\nname = \"cpp\"\ndescription = \"C++ starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"cpp/README.md.tmpl":                   {Data: []byte("By {{.Author}}\n")},
		"rust/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"rust\"\ndescription = \"Rust starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n\n[[inputs]]\nkey = \"year\"\ntemplate_var = \"Year\"\nrequired = false\n")},
		"rust/.cargo/config.toml":              {Data: []byte("[alias]\ndocs = \"doc --workspace --all-features --no-deps\"\n")},
		"rust/.editorconfig":                   {Data: []byte("root = true\n")},
		"rust/.env.example":                    {Data: []byte("APP_ENV=development\n")},
		"rust/.github/dependabot.yml":          {Data: []byte("version: 2\n")},
		"rust/.github/workflows/ci.yml":        {Data: []byte("name: ci\n")},
		"rust/CONTRIBUTING.md.tmpl":            {Data: []byte("# Contributing to {{.ProjectName}}\n")},
		"rust/Cargo.toml.tmpl":                 {Data: []byte("[package]\nname = \"{{.ProjectName}}\"\nauthors = [\"{{.Author}}\"]\n")},
		"rust/README.md.tmpl":                  {Data: []byte("# {{.ProjectName}}\nGenerated in {{.Year}}\n")},
		"rust/SECURITY.md.tmpl":                {Data: []byte("# Security Policy\n{{.ProjectName}}\n")},
		"rust/clippy.toml":                     {Data: []byte("allow-dbg-in-tests = true\n")},
		"rust/dprint.json":                     {Data: []byte("{\n  \"includes\": [\"**/*.toml\"]\n}\n")},
		"rust/justfile.tmpl":                   {Data: []byte("build:\n    cargo build\n")},
		"rust/rustfmt.toml":                    {Data: []byte("wrap_comments = true\n")},
		"rust/src/lib.rs.tmpl":                 {Data: []byte("pub fn greet(name: &str) -> String { format!(\"Hello, {name}!\") }\n")},
		"rust/src/main.rs.tmpl":                {Data: []byte("fn app_name() -> &'static str { \"{{.ProjectName}}\" }\n")},
		"rust/typos.toml":                      {Data: []byte("[files]\nextend-exclude = [\"target\"]\n")},
	}
	svc := NewService(fsys, nil)

	got, err := svc.ListSummaries()
	if err != nil {
		t.Fatalf("ListSummaries() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("ListSummaries() len = %d, want 3", len(got))
	}
	if !reflect.DeepEqual([]string{got[0].Name, got[1].Name, got[2].Name}, []string{"cpp", "go", "rust"}) {
		t.Fatalf("ListSummaries() names = %#v, want sorted [cpp go rust]", got)
	}
}

func filepathToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
