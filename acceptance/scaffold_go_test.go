//go:build acceptance

package acceptance

import (
	"path/filepath"
	"testing"
)

func TestAcceptance_GoTemplateBuildsAndTests(t *testing.T) {
	requireTool(t, "go")

	workDir := scaffold(t, nil,
		"new", "--lang", "go", "--git", "none",
		"--module", "example.com/acceptgo", "acceptgo",
	)
	projectDir := filepath.Join(workDir, "acceptgo")

	requireFiles(t, projectDir, []string{
		"go.mod",
		// cmd/acceptgo comes from the templated path segment
		// `cmd/{{.ProjectNameLower}}`, which only real templates exercise.
		"cmd/acceptgo/main.go",
		"internal/app/app.go",
		"internal/app/app_test.go",
		"internal/version/version.go",
		"internal/version/version_test.go",
		".golangci.yml",
		".goreleaser.yml",
		".github/workflows/ci.yml",
		".github/dependabot.yml",
		".gitignore",
		".editorconfig",
		"justfile",
		"README.md",
		"CONTRIBUTING.md",
		"CODE_OF_CONDUCT.md",
		"typos.toml",
		"codecov.yml",
	})

	requireCleanRender(t, projectDir)
	requireFileContains(t, projectDir, "go.mod", "module example.com/acceptgo")
	requireFileContains(t, projectDir, "cmd/acceptgo/main.go", `"example.com/acceptgo/internal/app"`)
	requireExecutable(t, projectDir, "justfile")

	// GOTOOLCHAIN=local keeps the generated go directive from triggering a
	// toolchain download, so the check stays offline and deterministic.
	goEnv := []string{"GOFLAGS=-mod=mod", "GOTOOLCHAIN=local"}
	runTool(t, projectDir, goEnv, "go", "build", "./...")
	runTool(t, projectDir, goEnv, "go", "test", "./...")
}
