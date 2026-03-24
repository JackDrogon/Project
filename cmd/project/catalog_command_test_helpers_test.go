package main

import (
	"bytes"
	"io/fs"
	"testing"
	"testing/fstest"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
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

var _ = bytes.Buffer{}
