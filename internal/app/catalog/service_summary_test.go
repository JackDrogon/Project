package catalog

import (
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestServiceListSummaries(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":   {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n")},
		"go/.editorconfig":                     {Data: []byte("root = true\n")},
		"go/.github/dependabot.yml":            {Data: []byte("version: 2\n")},
		"go/.github/workflows/ci.yml":          {Data: []byte("name: CI\n")},
		"go/go.mod.tmpl":                       {Data: []byte("module {{.ModulePath}}\n")},
		"go/typos.toml":                        {Data: []byte("[files]\nextend-exclude = [\"bin\"]\n")},
		"cpp/.project-template-manifest.toml":  {Data: []byte("version = 2\nname = \"cpp\"\ndescription = \"C++ starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"cpp/.editorconfig":                    {Data: []byte("root = true\n")},
		"cpp/.github/dependabot.yml":           {Data: []byte("version: 2\n")},
		"cpp/.github/workflows/ci.yml":         {Data: []byte("name: ci\n")},
		"cpp/.gitignore":                       {Data: []byte("/build/\n")},
		"cpp/CONTRIBUTING.md.tmpl":             {Data: []byte("# Contributing to {{.ProjectName}}\n")},
		"cpp/README.md.tmpl":                   {Data: []byte("By {{.Author}}\n")},
		"cpp/typos.toml":                       {Data: []byte("[files]\nextend-exclude = [\"build\"]\n")},
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

	got, err := svc.listSummaries()
	if err != nil {
		t.Fatalf("listSummaries() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("listSummaries() len = %d, want 3", len(got))
	}
	if !reflect.DeepEqual([]string{got[0].Name, got[1].Name, got[2].Name}, []string{"cpp", "go", "rust"}) {
		t.Fatalf("listSummaries() names = %#v, want sorted [cpp go rust]", got)
	}
}

func TestServiceQuerySummaries(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":   {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n")},
		"go/.editorconfig":                     {Data: []byte("root = true\n")},
		"go/.github/dependabot.yml":            {Data: []byte("version: 2\n")},
		"go/.github/workflows/ci.yml":          {Data: []byte("name: CI\n")},
		"go/.gitignore":                        {Data: []byte("bin/\n")},
		"go/go.mod.tmpl":                       {Data: []byte("module {{.ModulePath}}\n")},
		"go/typos.toml":                        {Data: []byte("[files]\nextend-exclude = [\"bin\"]\n")},
		"cpp/.project-template-manifest.toml":  {Data: []byte("version = 2\nname = \"cpp\"\ndescription = \"C++ starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"cpp/.editorconfig":                    {Data: []byte("root = true\n")},
		"cpp/.github/dependabot.yml":           {Data: []byte("version: 2\n")},
		"cpp/.github/workflows/ci.yml":         {Data: []byte("name: ci\n")},
		"cpp/.gitignore":                       {Data: []byte("/build/\n")},
		"cpp/CONTRIBUTING.md.tmpl":             {Data: []byte("# Contributing to {{.ProjectName}}\n")},
		"cpp/README.md.tmpl":                   {Data: []byte("By {{.Author}}\n")},
		"cpp/typos.toml":                       {Data: []byte("[files]\nextend-exclude = [\"build\"]\n")},
		"rust/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"rust\"\ndescription = \"Rust starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n\n[[inputs]]\nkey = \"year\"\ntemplate_var = \"Year\"\nrequired = false\n")},
		"rust/.cargo/config.toml":              {Data: []byte("[alias]\ndocs = \"doc --workspace --all-features --no-deps\"\n")},
		"rust/.editorconfig":                   {Data: []byte("root = true\n")},
		"rust/.env.example":                    {Data: []byte("APP_ENV=development\n")},
		"rust/.github/dependabot.yml":          {Data: []byte("version: 2\n")},
		"rust/.github/workflows/ci.yml":        {Data: []byte("name: ci\n")},
		"rust/.gitignore":                      {Data: []byte("/target/\n")},
		"rust/CONTRIBUTING.md.tmpl":            {Data: []byte("# Contributing to {{.ProjectName}}\n")},
		"rust/Cargo.toml.tmpl":                 {Data: []byte("[package]\nname = \"{{.ProjectName}}\"\n")},
		"rust/README.md.tmpl":                  {Data: []byte("# {{.ProjectName}}\n")},
		"rust/SECURITY.md.tmpl":                {Data: []byte("# Security Policy\n")},
		"rust/clippy.toml":                     {Data: []byte("allow-dbg-in-tests = true\n")},
		"rust/dprint.json":                     {Data: []byte("{}\n")},
		"rust/justfile.tmpl":                   {Data: []byte("build:\n    cargo build\n")},
		"rust/rustfmt.toml":                    {Data: []byte("wrap_comments = true\n")},
		"rust/src/lib.rs.tmpl":                 {Data: []byte("pub fn greet() {}\n")},
		"rust/src/main.rs.tmpl":                {Data: []byte("fn main() {}\n")},
		"rust/typos.toml":                      {Data: []byte("[files]\nextend-exclude = [\"target\"]\n")},
	}
	svc := NewService(fsys, nil)

	t.Run("rich governance filter", func(t *testing.T) {
		got, err := svc.QuerySummaries(SummaryQuery{MinGovernance: "rich"})
		if err != nil {
			t.Fatalf("QuerySummaries() error = %v", err)
		}
		if len(got) != 1 || got[0].Name != "rust" {
			t.Fatalf("QuerySummaries() = %#v, want only rust", got)
		}
	})

	t.Run("repo asset filter", func(t *testing.T) {
		got, err := svc.QuerySummaries(SummaryQuery{RequiredAssets: []string{"security"}})
		if err != nil {
			t.Fatalf("QuerySummaries() error = %v", err)
		}
		if len(got) != 1 || got[0].Name != "rust" {
			t.Fatalf("QuerySummaries() = %#v, want only rust", got)
		}
	})

	t.Run("governance sort", func(t *testing.T) {
		got, err := svc.QuerySummaries(SummaryQuery{SortBy: SummarySortGovernance})
		if err != nil {
			t.Fatalf("QuerySummaries() error = %v", err)
		}
		if !reflect.DeepEqual([]string{got[0].Name, got[1].Name, got[2].Name}, []string{"rust", "cpp", "go"}) {
			t.Fatalf("QuerySummaries() names = %#v, want [rust cpp go]", []string{got[0].Name, got[1].Name, got[2].Name})
		}
	})

	t.Run("invalid query rejected", func(t *testing.T) {
		_, err := svc.QuerySummaries(SummaryQuery{MinGovernance: "elite"})
		if err == nil {
			t.Fatal("QuerySummaries() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --min-governance") {
			t.Fatalf("QuerySummaries() error = %v, want governance validation error", err)
		}
	})
}

func TestServiceWithSwappedPolicies(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":        {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n")},
		"go/.gitignore":                             {Data: []byte("bin/\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: fs.ModeDir | 0o755},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
	}

	originalRepoAssets := activeRepoAssets
	activeRepoAssets = newRepoAssetRegistry(map[string]string{"custom": "go.mod.tmpl"})
	t.Cleanup(func() { activeRepoAssets = originalRepoAssets })

	originalGovernanceTier := governanceTier
	governanceTier = func(Inspection) string { return "custom-tier" }
	t.Cleanup(func() { governanceTier = originalGovernanceTier })

	svc := NewService(fsys, nil)

	t.Run("custom repo assets affect inspection", func(t *testing.T) {
		got, err := svc.QueryInspection(InspectionQuery{Lang: "go", Mode: InspectModeAll})
		if err != nil {
			t.Fatalf("QueryInspection() error = %v", err)
		}
		if !reflect.DeepEqual(got.RepoAssets, []string{"custom"}) {
			t.Fatalf("QueryInspection() repo assets = %#v, want [custom]", got.RepoAssets)
		}
		if len(got.RepoFiles()) != 1 || got.RepoFiles()[0].Source != "go.mod.tmpl" {
			t.Fatalf("QueryInspection() repo files = %#v, want go.mod as repo file", got.RepoFiles())
		}
	})

	t.Run("custom governance affects summaries", func(t *testing.T) {
		got, err := svc.QuerySummaries(DefaultSummaryQuery())
		if err != nil {
			t.Fatalf("QuerySummaries() error = %v", err)
		}
		if len(got) != 1 || got[0].GovernanceTier != "custom-tier" {
			t.Fatalf("QuerySummaries() = %#v, want governance tier custom-tier", got)
		}
	})
}
