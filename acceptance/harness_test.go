//go:build acceptance

package acceptance

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// reservedManifestName is never copied into a generated project.
const reservedManifestName = ".project-template-manifest.toml"

// placeholderMarker is the exact prefix of this project's template variables
// (`{{.ProjectName}}`). Escaped output such as `{{ .Version }}` for GoReleaser or
// `{{ldflags}}` for just keeps a space or a letter after the braces, so finding
// this marker in a generated file always means rendering was skipped.
const placeholderMarker = "{{."

// projectBin is the CLI binary built once for the whole acceptance run.
var projectBin string

func TestMain(m *testing.M) {
	binDir, err := os.MkdirTemp("", "project-acceptance-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	code := buildAndRun(m, binDir)
	_ = os.RemoveAll(binDir)
	os.Exit(code)
}

// buildAndRun exists so the temp directory is still removed on a build failure;
// os.Exit in TestMain would skip any deferred cleanup.
func buildAndRun(m *testing.M, binDir string) int {
	projectBin = filepath.Join(binDir, "project")

	build := exec.Command("go", "build", "-o", projectBin, "./cmd/project")
	build.Dir = ".."
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build project binary: %v\n%s", err, out)
		return 1
	}

	return m.Run()
}

// scaffold runs the built CLI inside a fresh working directory and returns it.
func scaffold(t *testing.T, env []string, args ...string) string {
	t.Helper()

	workDir := t.TempDir()
	cmd := exec.Command(projectBin, args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), env...)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("project %s error = %v\n%s", strings.Join(args, " "), err, out)
	}

	return workDir
}

func requireTool(t *testing.T, name string) {
	t.Helper()

	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s is not installed: %v", name, err)
	}
}

func runTool(t *testing.T, dir string, env []string, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s error = %v\n%s", name, strings.Join(args, " "), err, out)
	}

	return string(out)
}

func requireFiles(t *testing.T, root string, paths []string) {
	t.Helper()

	for _, path := range paths {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("generated project is missing %q: %v", path, err)
		}
	}
}

// requireExecutable asserts only that an executable bit survived. embed.FS
// reports every entry as 0444, so a set bit proves the generated permission
// metadata was applied. Exact modes are not compared because the runner's umask
// affects newly created files.
func requireExecutable(t *testing.T, root, path string) {
	t.Helper()

	full := filepath.Join(root, path)
	info, err := os.Stat(full)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("%q mode = %v, want an executable bit set", path, info.Mode().Perm())
	}
}

func requireFileContains(t *testing.T, root, path, want string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s = %q, want it to contain %q", path, string(data), want)
	}
}

// requireCleanRender walks the generated tree and fails when the reserved
// manifest leaked into the output or when anything still carries an unrendered
// placeholder, in a file name, a directory name, or file content.
func requireCleanRender(t *testing.T, root string) {
	t.Helper()

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.Name() == reservedManifestName {
			return fmt.Errorf("reserved manifest was copied into the project: %s", path)
		}
		if strings.Contains(entry.Name(), placeholderMarker) {
			return fmt.Errorf("unrendered placeholder in path: %s", path)
		}
		if entry.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(data, []byte(placeholderMarker)) {
			return fmt.Errorf("unrendered placeholder inside file: %s", path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("generated tree failed render checks: %v", err)
	}
}
