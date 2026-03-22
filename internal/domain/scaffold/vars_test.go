package scaffold

import "testing"

func TestNewTemplateVars(t *testing.T) {
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

func TestResolveTemplateVarsRejectsUnknownAndInvalidInputs(t *testing.T) {
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
