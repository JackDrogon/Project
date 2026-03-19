package scaffold

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JackDrogon/project/pkg/templates"
)

func TestEmbeddedGoTemplateInspect(t *testing.T) {
	creator := NewCreator(templates.FS, &bytes.Buffer{})

	details, err := creator.InspectLang("go")
	if err != nil {
		t.Fatalf("InspectLang() error = %v", err)
	}

	if details.FileCount != 12 {
		t.Fatalf("details.FileCount = %d, want %d", details.FileCount, 12)
	}
	if details.TemplateCount != 9 {
		t.Fatalf("details.TemplateCount = %d, want %d", details.TemplateCount, 9)
	}

	wantVars := []string{"ModulePath", "ProjectName"}
	if !reflect.DeepEqual(details.Variables, wantVars) {
		t.Fatalf("details.Variables = %v, want %v", details.Variables, wantVars)
	}

	wantFiles := []TemplateFile{
		{Source: ".github/workflows/ci.yml", Output: ".github/workflows/ci.yml", IsTemplate: false},
		{Source: ".gitignore", Output: ".gitignore", IsTemplate: false},
		{Source: ".golangci.yml", Output: ".golangci.yml", IsTemplate: false},
		{Source: ".goreleaser.yml.tmpl", Output: ".goreleaser.yml", IsTemplate: true},
		{Source: "README.md.tmpl", Output: "README.md", IsTemplate: true},
		{Source: "cmd/app/main.go.tmpl", Output: "cmd/app/main.go", IsTemplate: true},
		{Source: "go.mod.tmpl", Output: "go.mod", IsTemplate: true},
		{Source: "internal/app/app.go.tmpl", Output: "internal/app/app.go", IsTemplate: true},
		{Source: "internal/app/app_test.go.tmpl", Output: "internal/app/app_test.go", IsTemplate: true},
		{Source: "internal/version/version.go.tmpl", Output: "internal/version/version.go", IsTemplate: true},
		{Source: "internal/version/version_test.go.tmpl", Output: "internal/version/version_test.go", IsTemplate: true},
		{Source: "justfile.tmpl", Output: "justfile", IsTemplate: true},
	}

	if !reflect.DeepEqual(details.Files, wantFiles) {
		t.Fatalf("details.Files = %#v, want %#v", details.Files, wantFiles)
	}
}

func TestEmbeddedGoTemplateCreate(t *testing.T) {
	var out bytes.Buffer
	creator := NewCreator(templates.FS, &out)
	tmp := withTempWorkingDir(t)

	if err := creator.Create(Options{Lang: "go", ProjectName: "demo", ModulePath: "example.com/demo", NoGit: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	checks := map[string]string{
		"go.mod":                           "module example.com/demo",
		"cmd/app/main.go":                  "example.com/demo/internal/app",
		"internal/app/app.go":              "flag.NewFlagSet(\"demo\"",
		"internal/app/app_test.go":         "Run([]string{\"--version\"}",
		"internal/version/version.go":      "var Version = \"dev\"",
		"internal/version/version_test.go": "func TestInfo",
		"README.md":                        "production-ready Go CLI starter",
		"justfile":                         "goreleaser release --snapshot --clean",
		".golangci.yml":                    "staticcheck",
		".github/workflows/ci.yml":         "go test ./...",
		".goreleaser.yml":                  "main: ./cmd/app",
	}

	for relPath, want := range checks {
		content, err := os.ReadFile(filepath.Join(tmp, "demo", relPath))
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", relPath, err)
		}
		if !strings.Contains(string(content), want) {
			t.Fatalf("%s content = %q, want contains %q", relPath, string(content), want)
		}
	}

	if !strings.Contains(out.String(), "Skipping git initialization") {
		t.Fatalf("output = %q, want git skip message", out.String())
	}
}

func TestEmbeddedGoTemplateJustfileWorksWithoutGit(t *testing.T) {
	if _, err := exec.LookPath("just"); err != nil {
		t.Skip("just is not installed")
	}

	creator := NewCreator(templates.FS, &bytes.Buffer{})
	tmp := withTempWorkingDir(t)

	if err := creator.Create(Options{Lang: "go", ProjectName: "demo", ModulePath: "example.com/demo", NoGit: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	cmd := exec.Command("just", "--list")
	cmd.Dir = filepath.Join(tmp, "demo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("just --list error = %v\n%s", err, string(output))
	}
	if !strings.Contains(string(output), "build") || !strings.Contains(string(output), "pre-commit") || !strings.Contains(string(output), "loc") || !strings.Contains(string(output), "todos") {
		t.Fatalf("just --list output = %q, want core and maintenance recipes", string(output))
	}
}

func TestEmbeddedGoTemplateJustfileIgnoresParentGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	if _, err := exec.LookPath("just"); err != nil {
		t.Skip("just is not installed")
	}

	creator := NewCreator(templates.FS, &bytes.Buffer{})
	tmp := withTempWorkingDir(t)
	parent := filepath.Join(tmp, "workspace")
	if err := os.MkdirAll(parent, 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", parent, err)
	}

	gitInit := exec.Command("git", "init")
	gitInit.Dir = parent
	if output, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("git init error = %v\n%s", err, string(output))
	}
	configName := exec.Command("git", "config", "user.name", "Template Tester")
	configName.Dir = parent
	if output, err := configName.CombinedOutput(); err != nil {
		t.Fatalf("git config user.name error = %v\n%s", err, string(output))
	}
	configEmail := exec.Command("git", "config", "user.email", "template@example.com")
	configEmail.Dir = parent
	if output, err := configEmail.CombinedOutput(); err != nil {
		t.Fatalf("git config user.email error = %v\n%s", err, string(output))
	}
	if err := os.WriteFile(filepath.Join(parent, "README.md"), []byte("parent repo\n"), 0644); err != nil {
		t.Fatalf("WriteFile(README.md) error = %v", err)
	}
	commit := exec.Command("git", "add", ".")
	commit.Dir = parent
	if output, err := commit.CombinedOutput(); err != nil {
		t.Fatalf("git add error = %v\n%s", err, string(output))
	}
	commit = exec.Command("git", "commit", "-m", "init")
	commit.Dir = parent
	if output, err := commit.CombinedOutput(); err != nil {
		t.Fatalf("git commit error = %v\n%s", err, string(output))
	}
	tag := exec.Command("git", "tag", "v9.9.9")
	tag.Dir = parent
	if output, err := tag.CombinedOutput(); err != nil {
		t.Fatalf("git tag error = %v\n%s", err, string(output))
	}

	if err := creator.Create(Options{Lang: "go", ProjectName: "demo", ModulePath: "example.com/demo", TargetDir: filepath.Join(parent, "demo"), NoGit: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	build := exec.Command("just", "build")
	build.Dir = filepath.Join(parent, "demo")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("just build error = %v\n%s", err, string(output))
	}
	version := exec.Command(filepath.Join(parent, "demo", "bin", "demo"), "--version")
	version.Dir = filepath.Join(parent, "demo")
	output, err := version.CombinedOutput()
	if err != nil {
		t.Fatalf("built binary --version error = %v\n%s", err, string(output))
	}
	if strings.TrimSpace(string(output)) != "dev" {
		t.Fatalf("version output = %q, want %q", strings.TrimSpace(string(output)), "dev")
	}
}
