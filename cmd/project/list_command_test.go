package main

import (
	"bytes"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func useCatalogServiceFactory(t *testing.T, factory func() *appcatalog.Service) {
	t.Helper()
	oldFactory := newCatalogService
	newCatalogService = factory
	t.Cleanup(func() {
		newCatalogService = oldFactory
	})
}

func newCommandTestCatalogService() *appcatalog.Service {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":        {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = true\n")},
		"cpp/.project-template-manifest.toml":       {Data: []byte("version = 2\nname = \"cpp\"\ndescription = \"C++ starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"rust/.project-template-manifest.toml":      {Data: []byte("version = 2\nname = \"rust\"\ndescription = \"Cargo-based Rust starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n\n[[inputs]]\nkey = \"year\"\ntemplate_var = \"Year\"\nrequired = false\n")},
		"go/.editorconfig":                          {Data: []byte("root = true\n")},
		"go/.github":                                {Mode: fs.ModeDir | 0o755},
		"go/.github/dependabot.yml":                 {Data: []byte("version: 2\n")},
		"go/.github/workflows":                      {Mode: fs.ModeDir | 0o755},
		"go/.github/workflows/ci.yml":               {Data: []byte("name: CI\n")},
		"go/.gitignore":                             {Data: []byte("bin/\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
		"go/typos.toml":                             {Data: []byte("[files]\nextend-exclude = [\"bin\"]\n")},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: fs.ModeDir | 0o755},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
		"cpp/.editorconfig":                         {Data: []byte("root = true\n")},
		"cpp/.github":                               {Mode: fs.ModeDir | 0o755},
		"cpp/.github/dependabot.yml":                {Data: []byte("version: 2\n")},
		"cpp/.github/workflows":                     {Mode: fs.ModeDir | 0o755},
		"cpp/.github/workflows/ci.yml":              {Data: []byte("name: ci\n")},
		"cpp/.gitignore":                            {Data: []byte("/build/\n")},
		"cpp/CONTRIBUTING.md.tmpl":                  {Data: []byte("# Contributing to {{.ProjectName}}\n")},
		"cpp/src/main.cc.tmpl":                      {Data: []byte("// {{.ProjectName}}\n")},
		"cpp/README.md.tmpl":                        {Data: []byte("By {{.Author}}\n")},
		"cpp/typos.toml":                            {Data: []byte("[files]\nextend-exclude = [\"build\"]\n")},
		"rust/.cargo/config.toml":                   {Data: []byte("[alias]\ndocs = \"doc --workspace --all-features --no-deps\"\n")},
		"rust/.cargo":                               {Mode: fs.ModeDir | 0o755},
		"rust/.editorconfig":                        {Data: []byte("root = true\n")},
		"rust/.env.example":                         {Data: []byte("APP_ENV=development\n")},
		"rust/.github":                              {Mode: fs.ModeDir | 0o755},
		"rust/.github/dependabot.yml":               {Data: []byte("version: 2\n")},
		"rust/.github/workflows":                    {Mode: fs.ModeDir | 0o755},
		"rust/.github/workflows/ci.yml":             {Data: []byte("name: ci\n")},
		"rust/.gitignore":                           {Data: []byte("/target/\n")},
		"rust/CONTRIBUTING.md.tmpl":                 {Data: []byte("# Contributing to {{.ProjectName}}\n")},
		"rust/Cargo.toml.tmpl":                      {Data: []byte("[package]\nname = \"{{.ProjectName}}\"\nauthors = [\"{{.Author}}\"]\n")},
		"rust/README.md.tmpl":                       {Data: []byte("# {{.ProjectName}}\nGenerated in {{.Year}}\n")},
		"rust/SECURITY.md.tmpl":                     {Data: []byte("# Security Policy\n{{.ProjectName}}\n")},
		"rust/clippy.toml":                          {Data: []byte("allow-dbg-in-tests = true\n")},
		"rust/dprint.json":                          {Data: []byte("{\n  \"includes\": [\"**/*.toml\"]\n}\n")},
		"rust/justfile.tmpl":                        {Data: []byte("build:\n    cargo build\ncheck:\n    cargo test\n")},
		"rust/rustfmt.toml":                         {Data: []byte("wrap_comments = true\n")},
		"rust/src":                                  {Mode: fs.ModeDir | 0o755},
		"rust/src/lib.rs.tmpl":                      {Data: []byte("pub fn greet(name: &str) -> String { format!(\"Hello, {name}!\") }\n#[cfg(test)]\nmod tests { use super::greet; #[test] fn greet_includes_name() { assert_eq!(greet(\"{{.ProjectName}}\"), \"Hello, {{.ProjectName}}!\"); } }\n")},
		"rust/src/main.rs.tmpl":                     {Data: []byte("fn app_name() -> &'static str { \"{{.ProjectName}}\" }\nfn main() { println!(\"Hello, {}!\", app_name()); }\n")},
		"rust/typos.toml":                           {Data: []byte("[files]\nextend-exclude = [\"target\"]\n")},
	}

	return appcatalog.NewService(fsys, nil)
}

func TestSelectedOutputFormat(t *testing.T) {
	if got := selectedOutputFormat(false); got != outputFormatText {
		t.Fatalf("selectedOutputFormat(false) = %q, want %q", got, outputFormatText)
	}
	if got := selectedOutputFormat(true); got != outputFormatTOML {
		t.Fatalf("selectedOutputFormat(true) = %q, want %q", got, outputFormatTOML)
	}
}

func TestListCmdOutputs(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})

	tests := []struct {
		name        string
		args        []string
		wantContain []string
		wantEither  [][2]string
	}{
		{name: "text", wantContain: []string{"cpp", "go", "rust"}},
		{name: "toml", args: []string{"--toml"}, wantContain: []string{"languages = ", "cpp", "go", "rust"}},
		{name: "detail text", args: []string{"--detail"}, wantContain: []string{`desc="C++ starter"`, `manifest=v2`, `inputs=[author]`, `files=8`, `vars=[Author, ProjectName]`, `repo=[ci, contributing, dependabot, editorconfig, gitignore, typos]`, `repo_files=6`, `governance=standard`}},
		{name: "detail compact text", args: []string{"--detail", "--compact"}, wantContain: []string{"cpp [standard]", "counts: files=8 templates=3 repo_files=6 manifest=v2", "repo: ci, contributing, dependabot, editorconfig, gitignore, typos"}},
		{name: "detail table text", args: []string{"--detail", "--table"}, wantContain: []string{"NAME", "GOVERNANCE", "REPO_FILES", "cpp", "go", "rust", "ci,dependabot,editorconfig,gitignore,typos"}},
		{name: "detail min governance filter", args: []string{"--detail", "--min-governance", "rich"}, wantContain: []string{"rust	desc=", "governance=rich"}},
		{name: "detail repo asset filter", args: []string{"--detail", "--has-repo-asset", "security"}, wantContain: []string{"rust	desc=", "security"}},
		{name: "detail governance sort", args: []string{"--detail", "--sort", "governance"}, wantContain: []string{"rust	desc=", "governance=rich", "cpp	desc=", "go	desc="}},
		{name: "detail repo-files sort", args: []string{"--detail", "--sort", "repo-files"}, wantContain: []string{"rust	desc=", "repo_files=8", "cpp	desc=", "repo_files=6", "go	desc=", "repo_files=5"}},
		{name: "detail toml", args: []string{"--detail", "--toml"}, wantContain: []string{"[[templates]]", "manifest_version = 2", "file_count = 8", "template_count = 3", "Cargo-based Rust starter", "file_count = 17", "template_count = 7", "Go starter", "file_count = 7", "template_count = 2", "repo_assets", "repo_file_count", "governance_tier"}, wantEither: [][2]string{{"name = 'cpp'", `name = "cpp"`}, {"input_names = ['author']", `input_names = ["author"]`}, {"name = 'rust'", `name = "rust"`}, {"input_names = ['author', 'year']", `input_names = ["author", "year"]`}, {"name = 'go'", `name = "go"`}, {"input_names = ['module_path', 'go_version']", `input_names = ["module_path", "go_version"]`}, {"repo_assets = ['ci', 'dependabot', 'editorconfig', 'gitignore', 'typos']", `repo_assets = ["ci", "dependabot", "editorconfig", "gitignore", "typos"]`}, {"repo_file_count = 5", `repo_file_count = 5`}, {"governance_tier = 'rich'", `governance_tier = "rich"`}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := requireSubcommand(t, creator, commandKeyList)
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			got := buf.String()
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Fatalf("output = %q, want contains %q", got, want)
				}
			}
			for _, want := range tt.wantEither {
				if !strings.Contains(got, want[0]) && !strings.Contains(got, want[1]) {
					t.Fatalf("output = %q, want contains either %q or %q", got, want[0], want[1])
				}
			}
		})
	}
}

func TestListCmd_SortBehavior(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})

	t.Run("governance order", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--detail", "--sort", "governance"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := buf.String()
		rustIndex := strings.Index(got, "rust\tdesc=")
		cppIndex := strings.Index(got, "cpp\tdesc=")
		goIndex := strings.Index(got, "go\tdesc=")
		if rustIndex == -1 || cppIndex == -1 || goIndex == -1 {
			t.Fatalf("output = %q, want rust/cpp/go detail rows", got)
		}
		if !(rustIndex < cppIndex && cppIndex < goIndex) {
			t.Fatalf("governance order = %q, want rust before cpp before go", got)
		}
	})

	t.Run("repo-files order", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--detail", "--sort", "repo-files"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := buf.String()
		rustIndex := strings.Index(got, "rust\tdesc=")
		cppIndex := strings.Index(got, "cpp\tdesc=")
		goIndex := strings.Index(got, "go\tdesc=")
		if rustIndex == -1 || cppIndex == -1 || goIndex == -1 {
			t.Fatalf("output = %q, want rust/cpp/go detail rows", got)
		}
		if !(rustIndex < cppIndex && cppIndex < goIndex) {
			t.Fatalf("repo-files order = %q, want rust before cpp before go", got)
		}
	})

	t.Run("sort without detail errors", func(t *testing.T) {
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetArgs([]string{"--sort", "governance"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "requires --detail") {
			t.Fatalf("Execute() error = %v, want --detail requirement", err)
		}
	})

	t.Run("invalid sort errors", func(t *testing.T) {
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetArgs([]string{"--detail", "--sort", "bogus"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --sort") {
			t.Fatalf("Execute() error = %v, want invalid sort error", err)
		}
	})

	t.Run("min governance filter", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--detail", "--min-governance", "rich"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := buf.String()
		if !strings.Contains(got, "rust\tdesc=") || strings.Contains(got, "cpp\tdesc=") || strings.Contains(got, "go\tdesc=") {
			t.Fatalf("min-governance output = %q, want only rust detail row", got)
		}
	})

	t.Run("repo asset filter", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--detail", "--has-repo-asset", "security"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := buf.String()
		if !strings.Contains(got, "rust\tdesc=") || strings.Contains(got, "cpp\tdesc=") || strings.Contains(got, "go\tdesc=") {
			t.Fatalf("repo-asset output = %q, want only rust detail row", got)
		}
	})

	t.Run("filters without detail errors", func(t *testing.T) {
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetArgs([]string{"--min-governance", "standard"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "filters require --detail") {
			t.Fatalf("Execute() error = %v, want detail filter requirement", err)
		}
	})

	t.Run("invalid governance filter errors", func(t *testing.T) {
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetArgs([]string{"--detail", "--min-governance", "elite"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --min-governance") {
			t.Fatalf("Execute() error = %v, want invalid governance filter error", err)
		}
	})

	t.Run("invalid repo asset filter errors", func(t *testing.T) {
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetArgs([]string{"--detail", "--has-repo-asset", "license"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --has-repo-asset") {
			t.Fatalf("Execute() error = %v, want invalid repo asset error", err)
		}
	})

	t.Run("compact toml rejected", func(t *testing.T) {
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetArgs([]string{"--detail", "--compact", "--toml"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "compact output is only supported for text format") {
			t.Fatalf("Execute() error = %v, want compact/toml error", err)
		}
	})

	t.Run("table without detail errors", func(t *testing.T) {
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetArgs([]string{"--table"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "--table requires --detail") {
			t.Fatalf("Execute() error = %v, want table detail requirement", err)
		}
	})

	t.Run("table toml rejected", func(t *testing.T) {
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetArgs([]string{"--detail", "--table", "--toml"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "table output is only supported for text format") {
			t.Fatalf("Execute() error = %v, want table/toml error", err)
		}
	})

	t.Run("table compact rejected", func(t *testing.T) {
		cmd := requireSubcommand(t, creator, commandKeyList)
		cmd.SetArgs([]string{"--detail", "--table", "--compact"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "--table cannot be combined with --compact") {
			t.Fatalf("Execute() error = %v, want table/compact error", err)
		}
	})
}

func TestListCmd_PropagatesCatalogErrors(t *testing.T) {
	useCatalogServiceFactory(t, func() *appcatalog.Service {
		return appcatalog.NewService(failingCatalogFS{err: errors.New("boom")}, nil)
	})
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})

	tests := []struct {
		name string
		args []string
	}{
		{name: "langs", args: nil},
		{name: "summaries", args: []string{"--detail"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := requireSubcommand(t, creator, commandKeyList)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() expected error, got nil")
			}
			if !strings.Contains(err.Error(), "failed to read templates") {
				t.Fatalf("Execute() error = %v, want template read error", err)
			}
		})
	}
}

func TestInspectCmdOutputs(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})

	tests := []struct {
		name        string
		args        []string
		wantContain []string
		wantEither  [][2]string
		wantErr     string
	}{
		{name: "text", args: []string{"go"}, wantContain: []string{"name: go", "manifest_version: 2", "repo_assets: ci, dependabot, editorconfig, gitignore, typos", "mode: all", "shown: 7", "repo_files:", "language_files:"}},
		{name: "compact text", args: []string{"go", "--compact"}, wantContain: []string{"go — Go starter", "manifest: v2 | mode: all | shown: 7 | files: 7 | templates: 2", "repo_files:", "language_files:"}},
		{name: "toml render", args: []string{"go", "--toml", "--mode", "render"}, wantContain: []string{"manifest_version = 2", "mode = ", "shown_count = 2", "[[files]]", "[[language_files]]", "repo_files = []", "source = "}, wantEither: [][2]string{{"name = 'go'", `name = "go"`}, {"mode = 'render'", `mode = "render"`}, {"source = 'go.mod.tmpl'", `source = "go.mod.tmpl"`}}},
		{name: "compact toml rejected", args: []string{"go", "--compact", "--toml"}, wantErr: "compact output is only supported for text format"},
		{name: "invalid mode", args: []string{"go", "--mode", "bogus"}, wantErr: "invalid --mode"},
		{name: "unsupported language", args: []string{"missing"}, wantErr: "unsupported language"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := requireSubcommand(t, creator, commandKeyInspect)
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("Execute() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Execute() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			got := buf.String()
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Fatalf("output = %q, want contains %q", got, want)
				}
			}
			for _, want := range tt.wantEither {
				if !strings.Contains(got, want[0]) && !strings.Contains(got, want[1]) {
					t.Fatalf("output = %q, want contains either %q or %q", got, want[0], want[1])
				}
			}
		})
	}
}
