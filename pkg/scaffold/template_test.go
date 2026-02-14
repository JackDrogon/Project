package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestRenderTemplate(t *testing.T) {
	vars := TemplateVars{
		ProjectName: "testproj",
		ModulePath:  "github.com/user/testproj",
		Author:      "testuser",
		Year:        2025,
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
		"lang/hello.txt.tmpl":      {Data: []byte("Hello, {{.ProjectName}}!")},
		"lang/plain.txt":           {Data: []byte("no templates here")},
		"lang/sub/nested.txt.tmpl": {Data: []byte("nested {{.Author}}")},
		"lang/config.yaml.tmpl":    {Data: []byte("name: {{.ProjectName}}")},
	}
	vars := TemplateVars{
		ProjectName: "demo",
		ModulePath:  "github.com/user/demo",
		Author:      "alice",
		Year:        2025,
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
	if string(got) != "Hello, demo!" {
		t.Errorf("hello.txt = %q, want %q", string(got), "Hello, demo!")
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
	if string(got) != "name: demo" {
		t.Errorf("config.yaml = %q, want %q", string(got), "name: demo")
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
