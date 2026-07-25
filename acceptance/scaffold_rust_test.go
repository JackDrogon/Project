//go:build acceptance

package acceptance

import (
	"path/filepath"
	"testing"
)

func TestAcceptance_RustTemplateChecks(t *testing.T) {
	requireTool(t, "cargo")

	workDir := scaffold(t, nil, "new", "--lang", "rust", "--git", "none", "acceptrust")
	projectDir := filepath.Join(workDir, "acceptrust")

	requireFiles(t, projectDir, []string{
		"Cargo.toml",
		"src/main.rs",
		"src/lib.rs",
		".cargo/config.toml",
		"rustfmt.toml",
		"clippy.toml",
		"dprint.json",
		".github/workflows/ci.yml",
		".github/dependabot.yml",
		".gitignore",
		".editorconfig",
		".env.example",
		"justfile",
		"README.md",
		"CONTRIBUTING.md",
		"SECURITY.md",
		"typos.toml",
	})

	requireCleanRender(t, projectDir)
	requireFileContains(t, projectDir, "Cargo.toml", `name = "acceptrust"`)
	requireExecutable(t, projectDir, "justfile")

	// The template declares no dependencies, so --offline keeps cargo away from
	// the network. CARGO_TARGET_DIR stays inside the per-test temp directory.
	cargoEnv := []string{"CARGO_TARGET_DIR=" + filepath.Join(t.TempDir(), "target")}
	runTool(t, projectDir, cargoEnv, "cargo", "check", "--offline")
}
