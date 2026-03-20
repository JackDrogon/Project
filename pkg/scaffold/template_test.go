package scaffold

import (
	"bytes"
	"errors"
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

type infoErrorFS struct {
	err error
}

func (f infoErrorFS) Open(name string) (fs.File, error) { return nil, f.err }
func (f infoErrorFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return []fs.DirEntry{stubDirEntry{name: "file.txt", mode: 0, infoErr: f.err}}, nil
}
func (f infoErrorFS) ReadFile(name string) ([]byte, error) { return []byte("content"), nil }

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
	fsys := fstest.MapFS{
		"lang/script.sh": {
			Data: []byte("#!/bin/sh\necho hello\n"),
			Mode: 0755,
		},
	}
	vars := TemplateVars{ProjectName: "demo"}

	dest := filepath.Join(t.TempDir(), "output")
	var buf bytes.Buffer
	if err := CopyEmbedDir(&buf, fsys, "lang", dest, vars); err != nil {
		t.Fatalf("CopyEmbedDir() error = %v", err)
	}

	info, err := os.Stat(filepath.Join(dest, "script.sh"))
	if err != nil {
		t.Fatalf("stat script.sh: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("script.sh mode = %o, want %o", info.Mode().Perm(), 0755)
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
	if !strings.Contains(got, "create demo/plain.txt") || !strings.Contains(got, "create demo/sub/") || !strings.Contains(got, "create demo/sub/nested.txt") {
		t.Fatalf("PreviewEmbedDir() output = %q", got)
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
		dest := filepath.Join(t.TempDir(), "out")
		if err := CopyEmbedDir(&bytes.Buffer{}, infoErrorFS{err: errors.New("info failed")}, "lang", dest, vars); err != nil {
			t.Fatalf("CopyEmbedDir() error = %v", err)
		}
		info, err := os.Stat(filepath.Join(dest, "file.txt"))
		if err != nil {
			t.Fatalf("Stat(file.txt) error = %v", err)
		}
		if info.Mode().Perm() != 0644 {
			t.Fatalf("file.txt mode = %o, want %o", info.Mode().Perm(), 0644)
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
