package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

type openErrorFS struct {
	fstest.MapFS
	err error
}

func (f openErrorFS) Open(name string) (fs.File, error) {
	return f.MapFS.Open(name)
}

func (f openErrorFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == "lang/sub" {
		return nil, f.err
	}
	return fs.ReadDir(f.MapFS, name)
}

func (f openErrorFS) ReadFile(name string) ([]byte, error) {
	if name == "lang/bad.txt" {
		return nil, f.err
	}
	return fs.ReadFile(f.MapFS, name)
}

type stubDirEntry struct {
	name    string
	mode    fs.FileMode
	infoErr error
}

func (d stubDirEntry) Name() string      { return d.name }
func (d stubDirEntry) IsDir() bool       { return d.mode.IsDir() }
func (d stubDirEntry) Type() fs.FileMode { return d.mode }
func (d stubDirEntry) Info() (fs.FileInfo, error) {
	return stubFileInfo{name: d.name, mode: d.mode}, d.infoErr
}

type stubFileInfo struct {
	name string
	mode fs.FileMode
}

func (i stubFileInfo) Name() string       { return i.name }
func (i stubFileInfo) Size() int64        { return 0 }
func (i stubFileInfo) Mode() fs.FileMode  { return i.mode }
func (i stubFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (i stubFileInfo) IsDir() bool        { return i.mode.IsDir() }
func (i stubFileInfo) Sys() any           { return nil }

func stubTemplateOSFuncs(t *testing.T) {
	t.Helper()
	oldMkdirAll := osMkdirAll
	oldWriteFile := osWriteFile
	t.Cleanup(func() {
		osMkdirAll = oldMkdirAll
		osWriteFile = oldWriteFile
	})
}

func TestRenderTemplate(t *testing.T) {
	vars := TemplateVars{
		ProjectName:      "testproj",
		ProjectNameLower: "testproj",
		ModulePath:       "github.com/user/testproj",
		Author:           "testuser",
		Year:             2025,
	}

	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			"simple variable",
			"module {{.ModulePath}}",
			"module github.com/user/testproj",
			false,
		},
		{
			"multiple variables",
			"# {{.ProjectName}}\nBy {{.Author}} ({{.Year}})",
			"# testproj\nBy testuser (2025)",
			false,
		},
		{
			"no template syntax",
			"plain text content",
			"plain text content",
			false,
		},
		{
			"empty content",
			"",
			"",
			false,
		},
		{
			"invalid template syntax",
			"{{.ProjectName",
			"",
			true,
		},
		{
			"unknown variable",
			"{{.Unknown}}",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplate([]byte(tt.content), vars)
			if tt.wantErr {
				if err == nil {
					t.Fatal("RenderTemplate() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("RenderTemplate() unexpected error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("RenderTemplate() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestCopyEmbedDir(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/hello.txt.tmpl":                    {Data: []byte("Hello, {{.ProjectName}}!")},
		"lang/plain.txt":                         {Data: []byte("no templates here")},
		"lang/sub/nested.txt.tmpl":               {Data: []byte("nested {{.Author}}")},
		"lang/config.yaml.tmpl":                  {Data: []byte("name: {{.ProjectName}}")},
		"lang/cmd/{{.ProjectNameLower}}/main.go": {Data: []byte("package main\n")},
	}
	vars := TemplateVars{
		ProjectName:      "Demo",
		ProjectNameLower: "demo",
		ModulePath:       "github.com/user/demo",
		Author:           "alice",
		Year:             2025,
	}

	destDir := t.TempDir()
	dest := filepath.Join(destDir, "output")

	var buf bytes.Buffer
	if err := CopyEmbedDir(&buf, fsys, "lang", dest, vars); err != nil {
		t.Fatalf("CopyEmbedDir() error = %v", err)
	}

	// Verify rendered file
	got, err := os.ReadFile(filepath.Join(dest, "hello.txt"))
	if err != nil {
		t.Fatalf("read hello.txt: %v", err)
	}
	if string(got) != "Hello, Demo!" {
		t.Errorf("hello.txt = %q, want %q", string(got), "Hello, Demo!")
	}

	// Verify plain file
	got, err = os.ReadFile(filepath.Join(dest, "plain.txt"))
	if err != nil {
		t.Fatalf("read plain.txt: %v", err)
	}
	if string(got) != "no templates here" {
		t.Errorf("plain.txt = %q, want %q", string(got), "no templates here")
	}

	// Verify nested file
	got, err = os.ReadFile(filepath.Join(dest, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("read sub/nested.txt: %v", err)
	}
	if string(got) != "nested alice" {
		t.Errorf("sub/nested.txt = %q, want %q", string(got), "nested alice")
	}

	// Verify .tmpl suffix stripped
	got, err = os.ReadFile(filepath.Join(dest, "config.yaml"))
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}
	if string(got) != "name: Demo" {
		t.Errorf("config.yaml = %q, want %q", string(got), "name: Demo")
	}
	if _, err := os.Stat(filepath.Join(dest, "cmd", "demo", "main.go")); err != nil {
		t.Fatalf("stat cmd/demo/main.go: %v", err)
	}

	// Verify .tmpl file does NOT exist
	if _, err := os.Stat(filepath.Join(dest, "config.yaml.tmpl")); err == nil {
		t.Error("config.yaml.tmpl should not exist after .tmpl suffix stripping")
	}
}

func TestCopyEmbedDir_TargetDirConflict(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/file.txt": {Data: []byte("content")},
	}
	vars := TemplateVars{ProjectName: "test"}

	destDir := t.TempDir()
	dest := filepath.Join(destDir, "existing")
	if err := os.Mkdir(dest, 0755); err != nil {
		t.Fatal(err)
	}

	// CopyEmbedDir itself doesn't check for conflicts (Creator.Create does),
	// but it should still work when the directory exists
	var buf bytes.Buffer
	if err := CopyEmbedDir(&buf, fsys, "lang", dest, vars); err != nil {
		t.Fatalf("CopyEmbedDir() error = %v", err)
	}
}

func TestCopyEmbedDir_InvalidTemplateFails(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/bad.txt.tmpl": {Data: []byte("{{.ProjectName")},
	}
	vars := TemplateVars{ProjectName: "demo"}

	dest := filepath.Join(t.TempDir(), "output")
	var buf bytes.Buffer
	err := CopyEmbedDir(&buf, fsys, "lang", dest, vars)
	if err == nil {
		t.Fatal("CopyEmbedDir() expected error for invalid .tmpl file, got nil")
	}
}

func TestCopyEmbedDir_NonTemplateFileBypassesRendering(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/raw.txt": {Data: []byte("{{.ProjectName")},
	}
	vars := TemplateVars{ProjectName: "demo"}

	dest := filepath.Join(t.TempDir(), "output")
	var buf bytes.Buffer
	if err := CopyEmbedDir(&buf, fsys, "lang", dest, vars); err != nil {
		t.Fatalf("CopyEmbedDir() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "raw.txt"))
	if err != nil {
		t.Fatalf("read raw.txt: %v", err)
	}
	if string(got) != "{{.ProjectName" {
		t.Errorf("raw.txt = %q, want %q", string(got), "{{.ProjectName")
	}
}

func TestCopyEmbedDir_PreservesExecutableBit(t *testing.T) {
	stubTemplateOSFuncs(t)
	fsys := fstest.MapFS{
		"lang/script.sh": {
			Data: []byte("#!/bin/sh\necho hello\n"),
			Mode: 0755,
		},
	}
	vars := TemplateVars{ProjectName: "demo"}

	dest := filepath.Join(t.TempDir(), "output")
	var buf bytes.Buffer
	var gotPerm os.FileMode
	osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		gotPerm = perm
		return os.WriteFile(name, data, perm)
	}
	if err := CopyEmbedDir(&buf, fsys, "lang", dest, vars); err != nil {
		t.Fatalf("CopyEmbedDir() error = %v", err)
	}
	if gotPerm != 0755 {
		t.Fatalf("osWriteFile perm = %o, want %o", gotPerm, 0755)
	}
}

func TestPreviewEmbedDir(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/plain.txt":           {Data: []byte("plain")},
		"lang/sub/nested.txt.tmpl": {Data: []byte("hello {{.ProjectName}}")},
	}

	var buf bytes.Buffer
	if err := PreviewEmbedDir(&buf, fsys, "lang", "demo", TemplateVars{ProjectName: "Demo", ProjectNameLower: "demo"}); err != nil {
		t.Fatalf("PreviewEmbedDir() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "create "+filepath.Join("demo", "plain.txt")) || !strings.Contains(got, "create "+filepath.Join("demo", "sub")+string(filepath.Separator)) || !strings.Contains(got, "create "+filepath.Join("demo", "sub", "nested.txt")) {
		t.Fatalf("PreviewEmbedDir() output = %q", got)
	}
}

func TestWriteDryRunPlan(t *testing.T) {
	plan := DryRunPlan{
		Template:    "go",
		Description: "Go starter",
		TargetDir:   "demo",
		ResolvedInputs: []DryRunResolvedInput{
			{Name: "module_path", TemplateVar: "ModulePath", Value: "example.com/demo"},
			{Name: "go_version", TemplateVar: "GoVersion", Value: "1.21"},
			{Name: "author", TemplateVar: "Author", Value: "alice"},
		},
		Actions: []DryRunAction{
			{Kind: DryRunActionCopyFile, Source: "go/README.md", Target: filepath.Join("demo", "README.md")},
			{Kind: DryRunActionCreateDir, Target: filepath.Join("demo", "cmd")},
			{Kind: DryRunActionRenderFile, Source: "go/go.mod.tmpl", Target: filepath.Join("demo", "go.mod")},
		},
	}

	var buf bytes.Buffer
	err := writeDryRunPlan(&buf, plan, Options{Lang: "go", ProjectName: "demo", ModulePath: "example.com/demo", GitMode: GitModeNone})
	if err != nil {
		t.Fatalf("writeDryRunPlan() error = %v", err)
	}

	want := strings.Join([]string{
		"template: go",
		"description: Go starter",
		"target_dir: demo",
		"resolved inputs:",
		"  project_name: demo",
		"  module_path: example.com/demo",
		"  go_version: 1.21",
		"  author: alice",
		"  git_mode: none",
		"explicit overrides:",
		"  module_path: example.com/demo",
		"  git_mode: none",
		"actions:",
		"  copy go/README.md -> " + filepath.Join("demo", "README.md"),
		fmt.Sprintf("  create %s%s", filepath.Join("demo", "cmd"), string(filepath.Separator)),
		"  render go/go.mod.tmpl -> " + filepath.Join("demo", "go.mod"),
	}, "\n") + "\n"

	if buf.String() != want {
		t.Fatalf("writeDryRunPlan() output = %q, want %q", buf.String(), want)
	}
}

func TestPreviewEmbedDir_ReadDirError(t *testing.T) {
	err := PreviewEmbedDir(&bytes.Buffer{}, failingReadDirFS{err: os.ErrPermission}, "lang", "demo", TemplateVars{})
	if err == nil {
		t.Fatal("PreviewEmbedDir() expected error, got nil")
	}
}

func TestPreviewEmbedDir_NestedReadDirError(t *testing.T) {
	fsys := openErrorFS{
		MapFS: fstest.MapFS{
			"lang/sub/nested.txt": {Data: []byte("hello")},
		},
		err: errors.New("nested read failed"),
	}
	err := PreviewEmbedDir(&bytes.Buffer{}, fsys, "lang", "demo", TemplateVars{})
	if err == nil {
		t.Fatal("PreviewEmbedDir() expected nested error, got nil")
	}
}

func TestPreviewEmbedDir_ReadFileError(t *testing.T) {
	fsys := openErrorFS{
		MapFS: fstest.MapFS{"lang/bad.txt": {Data: []byte("content")}},
		err:   errors.New("read file failed"),
	}

	err := PreviewEmbedDir(&bytes.Buffer{}, fsys, "lang", "demo", TemplateVars{})
	if err == nil {
		t.Fatal("PreviewEmbedDir() expected read file error, got nil")
	}
}

func TestPreviewEmbedDir_InvalidTemplateFails(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/bad.txt.tmpl": {Data: []byte("{{.ProjectName")},
	}

	err := PreviewEmbedDir(&bytes.Buffer{}, fsys, "lang", "demo", TemplateVars{ProjectName: "demo"})
	if err == nil {
		t.Fatal("PreviewEmbedDir() expected error for invalid .tmpl file, got nil")
	}
}

func TestPreviewEmbedDir_NonTemplateFileBypassesRendering(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/raw.txt": {Data: []byte("{{.ProjectName")},
	}

	if err := PreviewEmbedDir(&bytes.Buffer{}, fsys, "lang", "demo", TemplateVars{ProjectName: "demo"}); err != nil {
		t.Fatalf("PreviewEmbedDir() error = %v", err)
	}
}

func TestCopyEmbedDir_Errors(t *testing.T) {
	vars := TemplateVars{ProjectName: "demo"}

	t.Run("mkdir all failure", func(t *testing.T) {
		fsys := fstest.MapFS{"lang/file.txt": {Data: []byte("content")}}
		dest := filepath.Join(t.TempDir(), "existing")
		if err := os.WriteFile(dest, []byte("x"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		if err := CopyEmbedDir(&bytes.Buffer{}, fsys, "lang", dest, vars); err == nil {
			t.Fatal("CopyEmbedDir() expected mkdir error, got nil")
		}
	})

	t.Run("read dir failure", func(t *testing.T) {
		if err := CopyEmbedDir(&bytes.Buffer{}, failingReadDirFS{err: os.ErrPermission}, "lang", filepath.Join(t.TempDir(), "out"), vars); err == nil {
			t.Fatal("CopyEmbedDir() expected read dir error, got nil")
		}
	})

	t.Run("read file failure", func(t *testing.T) {
		fsys := openErrorFS{
			MapFS: fstest.MapFS{"lang/bad.txt": {Data: []byte("content")}},
			err:   errors.New("read file failed"),
		}
		if err := CopyEmbedDir(&bytes.Buffer{}, fsys, "lang", filepath.Join(t.TempDir(), "out"), vars); err == nil {
			t.Fatal("CopyEmbedDir() expected read file error, got nil")
		}
	})

	t.Run("write file failure", func(t *testing.T) {
		stubTemplateOSFuncs(t)
		fsys := fstest.MapFS{"lang/file.txt": {Data: []byte("content")}}
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			return errors.New("write failed")
		}

		if err := CopyEmbedDir(&bytes.Buffer{}, fsys, "lang", filepath.Join(t.TempDir(), "out"), vars); err == nil {
			t.Fatal("CopyEmbedDir() expected write error, got nil")
		}
	})

	t.Run("entry info error falls back to default mode", func(t *testing.T) {
		entry := templateEntry{entry: stubDirEntry{name: "file.txt", infoErr: errors.New("info failed")}}
		if got := templateEntryMode(entry); got != 0644 {
			t.Fatalf("templateEntryMode() = %o, want %o", got, 0644)
		}
	})

	t.Run("nested directory recursion error", func(t *testing.T) {
		fsys := openErrorFS{
			MapFS: fstest.MapFS{
				"lang/sub/file.txt": {Data: []byte("content")},
			},
			err: errors.New("nested copy failed"),
		}
		if err := CopyEmbedDir(&bytes.Buffer{}, fsys, "lang", filepath.Join(t.TempDir(), "out"), vars); err == nil {
			t.Fatal("CopyEmbedDir() expected nested recursion error, got nil")
		}
	})
}

func TestRenderTemplatePathSegment(t *testing.T) {
	tests := []struct {
		name    string
		segment string
		vars    TemplateVars
		want    string
		wantErr string
	}{
		{
			name:    "valid rendered segment",
			segment: "{{.ProjectNameLower}}.txt",
			vars:    TemplateVars{ProjectNameLower: "demo"},
			want:    "demo.txt",
		},
		{
			name:    "invalid template syntax is returned",
			segment: "{{.ProjectName",
			vars:    TemplateVars{ProjectName: "demo"},
			wantErr: "unclosed action",
		},
		{
			name:    "empty rendered segment is rejected",
			segment: "{{if .ProjectName}}{{.ProjectName}}{{end}}",
			vars:    TemplateVars{},
			wantErr: "invalid rendered path segment",
		},
		{
			name:    "dot segment is rejected",
			segment: ".",
			vars:    TemplateVars{},
			wantErr: "invalid rendered path segment",
		},
		{
			name:    "dot dot segment is rejected",
			segment: "..",
			vars:    TemplateVars{},
			wantErr: "invalid rendered path segment",
		},
		{
			name:    "slash in rendered segment is rejected",
			segment: "{{.ModulePath}}",
			vars:    TemplateVars{ModulePath: "acme/demo"},
			wantErr: "must not contain path separators",
		},
		{
			name:    "backslash in rendered segment is rejected",
			segment: "{{.ProjectName}}",
			vars:    TemplateVars{ProjectName: `demo\\nested`},
			wantErr: "must not contain path separators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTemplatePathSegment(tt.segment, tt.vars)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("renderTemplatePathSegment() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("renderTemplatePathSegment() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("renderTemplatePathSegment() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("renderTemplatePathSegment() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPreviewEmbedDir_InvalidRenderedPathFails(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/{{.ModulePath}}.txt": {Data: []byte("content")},
	}

	err := PreviewEmbedDir(&bytes.Buffer{}, fsys, "lang", "demo", TemplateVars{ModulePath: "acme/demo"})
	if err == nil {
		t.Fatal("PreviewEmbedDir() expected path rendering error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to render template path") {
		t.Fatalf("PreviewEmbedDir() error = %v, want wrapped path rendering error", err)
	}
}
