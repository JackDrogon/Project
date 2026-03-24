package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	appcreate "github.com/JackDrogon/project/internal/app/create"
	appversion "github.com/JackDrogon/project/internal/app/version"
	"github.com/spf13/cobra"
)

func assertCLIContractGolden(t *testing.T, name, got string) {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	path := filepath.Join(filepath.Dir(file), "testdata", "cli_contracts", name)
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	normalizedGot := normalizeCLIContractText(got)
	normalizedWant := normalizeCLIContractText(string(want))
	if normalizedGot != normalizedWant {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, normalizedGot, normalizedWant)
	}
}

func normalizeCLIContractText(s string) string {
	normalized := strings.ReplaceAll(s, "\r\n", "\n")
	if os.PathSeparator != '/' {
		normalized = strings.ReplaceAll(normalized, string(os.PathSeparator), "/")
	}
	return normalized
}

func normalizeVersionContractOutput(tag, got string) string {
	normalized := normalizeCLIContractText(strings.TrimSpace(got))
	if strings.HasPrefix(normalized, tag) {
		return tag + "<build-meta>\n"
	}
	return normalized + "\n"
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = oldStderr
	})

	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	return string(data)
}

func newCLIContractDryRunCreator(out *bytes.Buffer) *appcreate.Creator {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":        {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = false\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n\n[[inputs]]\nkey = \"year\"\ntemplate_var = \"Year\"\nrequired = false\n")},
		"go/README.md":                              {Data: []byte("# README\n")},
		"go/cmd":                                    {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
	}

	return appcreate.NewCreator(fsys, out)
}

func renderCLICommandTree(cmd *cobra.Command) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\t%s\n", cmd.Use, cmd.Short)
	for _, sub := range cmd.Commands() {
		if sub.Name() == "help" {
			continue
		}
		fmt.Fprintf(&b, "%s\t%s\n", sub.Use, sub.Short)
	}
	return b.String()
}

func TestCLIContract_CommandTree(t *testing.T) {
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	assertCLIContractGolden(t, "command-tree.txt", renderCLICommandTree(newRootCmd(creator)))
}

func TestCLIContract_NewDryRunText(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	var out bytes.Buffer
	creator := newCLIContractDryRunCreator(&out)
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--git", "none", "--module", "example.com/demo", "--set", "go_version=1.25", "--set", "author=cli-contract", "--set", "year=2042", "demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create destination directory, stat err = %v", err)
	}

	assertCLIContractGolden(t, "new-dry-run.txt", out.String())
}

func TestCLIContract_InitDryRunText(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	var out bytes.Buffer
	creator := newCLIContractDryRunCreator(&out)
	cmd := requireSubcommand(t, creator, commandKeyInit)
	targetDir := filepath.Join("nested", "demo")
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--git", "none", "--module", "example.com/demo", "--set", "go_version=1.25", "--set", "author=cli-contract", "--set", "year=2042", targetDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, targetDir)); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create target directory, stat err = %v", err)
	}

	assertCLIContractGolden(t, "init-dry-run.txt", out.String())
}

func TestCLIContract_ListText(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	var buf bytes.Buffer
	cmd := requireSubcommand(t, appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{}), commandKeyList)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "list-text.txt", buf.String())
}

func TestCLIContract_ListDetailText(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	var buf bytes.Buffer
	cmd := requireSubcommand(t, appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{}), commandKeyList)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--detail"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "list-detail-text.txt", buf.String())
}

func TestCLIContract_ListDetailCompactText(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	var buf bytes.Buffer
	cmd := requireSubcommand(t, appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{}), commandKeyList)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--detail", "--compact"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "list-detail-compact-text.txt", buf.String())
}

func TestCLIContract_ListDetailTableText(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	var buf bytes.Buffer
	cmd := requireSubcommand(t, appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{}), commandKeyList)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--detail", "--table"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "list-detail-table-text.txt", buf.String())
}

func TestCLIContract_ListDetailGovernanceSortText(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	var buf bytes.Buffer
	cmd := requireSubcommand(t, appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{}), commandKeyList)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--detail", "--sort", "governance"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "list-detail-governance-text.txt", buf.String())
}

func TestCLIContract_ListDetailRichFilterText(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	var buf bytes.Buffer
	cmd := requireSubcommand(t, appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{}), commandKeyList)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--detail", "--min-governance", "rich"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "list-detail-rich-text.txt", buf.String())
}

func TestCLIContract_InspectText(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	var buf bytes.Buffer
	cmd := requireSubcommand(t, appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{}), commandKeyInspect)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"go"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "inspect-all-text.txt", buf.String())
}

func TestCLIContract_InspectCompactText(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	var buf bytes.Buffer
	cmd := requireSubcommand(t, appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{}), commandKeyInspect)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"go", "--compact"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "inspect-all-compact-text.txt", buf.String())
}

func TestCLIContract_InspectRenderText(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	var buf bytes.Buffer
	cmd := requireSubcommand(t, appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{}), commandKeyInspect)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"go", "--mode", "render"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "inspect-render-text.txt", buf.String())
}

func TestCLIContract_VersionDefault(t *testing.T) {
	oldFactory := newVersionService
	newVersionService = func() *appversion.Service {
		return appversion.NewService(stubVersionProvider{
			info:    "cli-contract-tag:abcdef0",
			verbose: "Tag:      cli-contract-tag\nRevision: abcdef0\nDirty:    false",
		})
	}
	t.Cleanup(func() {
		newVersionService = oldFactory
	})

	var buf bytes.Buffer
	cmd := newVersionCmd()
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertCLIContractGolden(t, "version-default.txt", normalizeVersionContractOutput("cli-contract-tag", buf.String()))
}

func TestCLIContract_ExecuteExitPath(t *testing.T) {
	oldArgs := os.Args
	oldExit := exitFunc
	oldStderr := stderrWriter
	os.Args = []string{"project", "new"}
	exitFunc = func(code int) {
		panic(exitPanic{code: code})
	}
	t.Cleanup(func() {
		os.Args = oldArgs
		exitFunc = oldExit
		stderrWriter = oldStderr
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	var exitCode int
	stderr := captureStderr(t, func() {
		stderrWriter = os.Stderr
		defer func() {
			r := recover()
			panicValue, ok := r.(exitPanic)
			if !ok {
				t.Fatalf("recover() = %#v, want exit panic", r)
			}
			exitCode = panicValue.code
		}()

		Execute(creator)
	})

	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	assertCLIContractGolden(t, "execute-error.txt", stderr)
}
