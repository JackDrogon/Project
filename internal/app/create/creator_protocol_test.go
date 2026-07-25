package create

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

// brokenFS fails every read with a non-ErrNotExist error, standing in for a
// corrupt or unreadable template tree.
type brokenFS struct{ err error }

func (f brokenFS) Open(string) (fs.File, error) {
	return nil, &fs.PathError{Op: "open", Path: "go", Err: f.err}
}

func newProtocolTestCreator(t *testing.T, manifest string) *Creator {
	t.Helper()

	return NewCreatorWithDeps(fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte(manifest)},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\n")},
	}, &bytes.Buffer{}, func(context.Context, string, ...string) error { return nil }, nil)
}

func TestCheckLang_SeparatesMissingTemplateFromReadFailure(t *testing.T) {
	t.Parallel()

	opts := Options{Lang: "go", ProjectName: "demo", GitMode: domain.GitModeNone}

	t.Run("missing template is a user error", func(t *testing.T) {
		creator := NewCreator(fstest.MapFS{}, io.Discard)

		err := creator.checkLang(opts)
		if err == nil || err.Error() != "unsupported language: go" {
			t.Fatalf("checkLang() error = %v, want unsupported language", err)
		}
	})

	t.Run("invalid template name is a user error", func(t *testing.T) {
		creator := NewCreator(fstest.MapFS{}, io.Discard)

		err := creator.checkLang(Options{Lang: "../escape", ProjectName: "demo"})
		if err == nil || err.Error() != "unsupported language: ../escape" {
			t.Fatalf("checkLang() error = %v, want unsupported language", err)
		}
	})

	t.Run("read failure surfaces the cause", func(t *testing.T) {
		sentinel := errors.New("disk on fire")
		creator := NewCreator(brokenFS{err: sentinel}, io.Discard)

		err := creator.checkLang(opts)
		if err == nil {
			t.Fatal("checkLang() error = nil, want the read failure")
		}
		if !errors.Is(err, sentinel) {
			t.Fatalf("checkLang() error = %v, want wrapped %v", err, sentinel)
		}
		if strings.Contains(err.Error(), "unsupported language") {
			t.Fatalf("checkLang() error = %v, want an I/O failure not a user error", err)
		}
	})
}

func TestValidateCreateOptions_EnforcesManifestRequiredInputs(t *testing.T) {
	t.Parallel()

	const manifest = "version = 2\nname = \"go\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = true\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n"

	base := Options{
		Lang:        "go",
		ProjectName: "demo",
		ModulePath:  "example.com/demo",
		GitMode:     domain.GitModeNone,
	}

	t.Run("blank required input is rejected", func(t *testing.T) {
		opts := base
		opts.TemplateInputValues = map[string]string{"go_version": ""}

		err := newProtocolTestCreator(t, manifest).validateCreateOptions(t.Context(), opts)
		if err == nil {
			t.Fatal("validateCreateOptions() error = nil, want required-input failure")
		}
		if !strings.Contains(err.Error(), `"go_version"`) || !strings.Contains(err.Error(), "required") {
			t.Fatalf("validateCreateOptions() error = %v, want required go_version failure", err)
		}
	})

	t.Run("blank optional input is accepted", func(t *testing.T) {
		opts := base
		opts.TemplateInputValues = map[string]string{"author": ""}

		if err := newProtocolTestCreator(t, manifest).validateCreateOptions(t.Context(), opts); err != nil {
			t.Fatalf("validateCreateOptions() error = %v, want optional input to allow an empty value", err)
		}
	})

	t.Run("defaults satisfy required inputs", func(t *testing.T) {
		if err := newProtocolTestCreator(t, manifest).validateCreateOptions(t.Context(), base); err != nil {
			t.Fatalf("validateCreateOptions() error = %v, want detected defaults to satisfy the manifest", err)
		}
	})
}
