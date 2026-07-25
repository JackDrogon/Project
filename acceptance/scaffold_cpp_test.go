//go:build acceptance

package acceptance

import (
	"path/filepath"
	"testing"
)

func TestAcceptance_CppTemplateConfiguresAndBuilds(t *testing.T) {
	requireTool(t, "cmake")

	workDir := scaffold(t, nil, "new", "--lang", "cpp", "--git", "none", "acceptcpp")
	projectDir := filepath.Join(workDir, "acceptcpp")

	requireFiles(t, projectDir, []string{
		"CMakeLists.txt",
		"src/main.cc",
		"include/.gitkeep",
		"dev-tools/apply-format",
		"dev-tools/check-style.py",
		"dev-tools/cpplint.py",
		"dev-tools/git-pre-commit-format",
		"CPPLINT.cfg",
		"CHANGELOG",
		".github/workflows/ci.yml",
		".github/dependabot.yml",
		".gitignore",
		".editorconfig",
		"justfile",
		"README.md",
		"CONTRIBUTING.md",
		"typos.toml",
	})

	requireCleanRender(t, projectDir)
	requireFileContains(t, projectDir, "CMakeLists.txt", "project(acceptcpp")

	// embed.FS reports every entry as 0444, so these executable bits can only
	// come from the generated permission metadata surviving a real scaffold.
	for _, tool := range []string{
		"dev-tools/apply-format",
		"dev-tools/check-style.py",
		"dev-tools/cpplint.py",
		"dev-tools/git-pre-commit-format",
	} {
		requireExecutable(t, projectDir, tool)
	}

	runTool(t, projectDir, nil, "cmake", "-S", ".", "-B", "build")
	runTool(t, projectDir, nil, "cmake", "--build", "build")
}
