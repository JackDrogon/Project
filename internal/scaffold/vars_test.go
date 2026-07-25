package scaffold

import (
	"strings"
	"testing"
)

func TestNewTemplateVars(t *testing.T) {
	t.Parallel()

	got := NewTemplateVars("Demo", "example.com/demo", "1.25", "alice", 2030)
	if got.ProjectName != "Demo" {
		t.Fatalf("ProjectName = %q, want %q", got.ProjectName, "Demo")
	}
	if got.ProjectNameLower != "demo" {
		t.Fatalf("ProjectNameLower = %q, want %q", got.ProjectNameLower, "demo")
	}
	if got.ModulePath != "example.com/demo" {
		t.Fatalf("ModulePath = %q, want %q", got.ModulePath, "example.com/demo")
	}
	if got.GoVersion != "1.25" || got.Author != "alice" || got.Year != 2030 {
		t.Fatalf("TemplateVars = %#v, want explicit values preserved", got)
	}
}

func TestResolveTemplateVars(t *testing.T) {
	t.Parallel()

	base := NewTemplateVars("Demo", "example.com/demo", "1.22", "author", 2025)
	inputs := []ManifestInput{{Name: "go_version", TemplateVar: "GoVersion"}, {Name: "author", TemplateVar: "Author"}, {Name: "year", TemplateVar: "Year"}}
	req := CreateRequest{TemplateInputValues: map[string]string{"go_version": "1.25", "author": "alice", "year": "2030"}}

	got, err := ResolveTemplateVars(inputs, req, base)
	if err != nil {
		t.Fatalf("ResolveTemplateVars() error = %v", err)
	}
	if got.GoVersion != "1.25" || got.Author != "alice" || got.Year != 2030 {
		t.Fatalf("ResolveTemplateVars() = %#v, want overrides applied", got)
	}
}

func TestResolveTemplateVarsEnforcesRequiredInputs(t *testing.T) {
	t.Parallel()

	base := NewTemplateVars("Demo", "example.com/demo", "1.22", "author", 2025)
	inputs := []ManifestInput{
		{Name: "module_path", TemplateVar: "ModulePath", Required: true},
		{Name: "go_version", TemplateVar: "GoVersion", Required: true},
		{Name: "year", TemplateVar: "Year", Required: true},
		{Name: "author", TemplateVar: "Author"},
	}

	t.Run("defaults satisfy required inputs", func(t *testing.T) {
		if _, err := ResolveTemplateVars(inputs, CreateRequest{}, base); err != nil {
			t.Fatalf("ResolveTemplateVars() error = %v, want defaults to satisfy required inputs", err)
		}
	})

	t.Run("blank override fails", func(t *testing.T) {
		req := CreateRequest{TemplateInputValues: map[string]string{"go_version": "   "}}

		_, err := ResolveTemplateVars(inputs, req, base)
		if err == nil {
			t.Fatal("ResolveTemplateVars() error = nil, want required-input failure")
		}
		if !strings.Contains(err.Error(), `"go_version"`) || !strings.Contains(err.Error(), "required") {
			t.Fatalf("ResolveTemplateVars() error = %v, want required go_version failure", err)
		}
	})

	t.Run("zero year counts as unset", func(t *testing.T) {
		req := CreateRequest{TemplateInputValues: map[string]string{"year": "0"}}

		_, err := ResolveTemplateVars(inputs, req, base)
		if err == nil {
			t.Fatal("ResolveTemplateVars() error = nil, want required-input failure")
		}
		if !strings.Contains(err.Error(), `"year"`) {
			t.Fatalf("ResolveTemplateVars() error = %v, want required year failure", err)
		}
	})

	t.Run("empty base fails even without overrides", func(t *testing.T) {
		_, err := ResolveTemplateVars(inputs, CreateRequest{}, NewTemplateVars("Demo", "example.com/demo", "", "author", 2025))
		if err == nil {
			t.Fatal("ResolveTemplateVars() error = nil, want required-input failure")
		}
	})

	t.Run("optional inputs may stay empty", func(t *testing.T) {
		req := CreateRequest{TemplateInputValues: map[string]string{"author": ""}}

		got, err := ResolveTemplateVars(inputs, req, base)
		if err != nil {
			t.Fatalf("ResolveTemplateVars() error = %v, want optional input to accept an empty value", err)
		}
		if got.Author != "" {
			t.Fatalf("Author = %q, want the empty override applied", got.Author)
		}
	})
}

func TestTemplateVarValue(t *testing.T) {
	t.Parallel()

	vars := NewTemplateVars("Demo", "example.com/demo", "1.25", "alice", 2030)

	for _, tt := range []struct{ name, want string }{
		{name: "ProjectName", want: "Demo"},
		{name: "ProjectNameLower", want: "demo"},
		{name: "ModulePath", want: "example.com/demo"},
		{name: "GoVersion", want: "1.25"},
		{name: "Author", want: "alice"},
		{name: "Year", want: "2030"},
	} {
		got, ok := TemplateVarValue(vars, tt.name)
		if !ok || got != tt.want {
			t.Fatalf("TemplateVarValue(%q) = (%q, %t), want (%q, true)", tt.name, got, ok, tt.want)
		}
	}

	if got, ok := TemplateVarValue(vars, "Nope"); ok || got != "" {
		t.Fatalf(`TemplateVarValue("Nope") = (%q, %t), want ("", false)`, got, ok)
	}
}

func TestResolveTemplateVarsRejectsUnknownAndInvalidInputs(t *testing.T) {
	t.Parallel()

	base := NewTemplateVars("Demo", "example.com/demo", "1.22", "author", 2025)
	inputs := []ManifestInput{{Name: "year", TemplateVar: "Year"}}

	t.Run("unknown input", func(t *testing.T) {
		_, err := ResolveTemplateVars(inputs, CreateRequest{TemplateInputValues: map[string]string{"author": "alice"}}, base)
		if err == nil {
			t.Fatal("ResolveTemplateVars() expected error, got nil")
		}
	})

	t.Run("invalid year", func(t *testing.T) {
		_, err := ResolveTemplateVars(inputs, CreateRequest{TemplateInputValues: map[string]string{"year": "twenty"}}, base)
		if err == nil {
			t.Fatal("ResolveTemplateVars() expected error, got nil")
		}
	})
}
