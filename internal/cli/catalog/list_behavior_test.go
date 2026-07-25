package catalog

import (
	"bytes"
	"strings"
	"testing"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
)

func TestListCmd_SortBehavior(t *testing.T) {
	t.Parallel()

	t.Run("governance order", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--detail", "--sort", "governance"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := buf.String()
		rustIndex := strings.Index(got, "rust\tdesc=")
		cppIndex := strings.Index(got, "cpp\tdesc=")
		goIndex := strings.Index(got, "go\tdesc=")
		if rustIndex == -1 || cppIndex == -1 || goIndex == -1 {
			t.Fatalf("output = %q, want rust/cpp/go detail rows", got)
		}
		if rustIndex >= cppIndex || cppIndex >= goIndex {
			t.Fatalf("governance order = %q, want rust before cpp before go", got)
		}
	})

	t.Run("repo-files order", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--detail", "--sort", "repo-files"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := buf.String()
		rustIndex := strings.Index(got, "rust\tdesc=")
		cppIndex := strings.Index(got, "cpp\tdesc=")
		goIndex := strings.Index(got, "go\tdesc=")
		if rustIndex == -1 || cppIndex == -1 || goIndex == -1 {
			t.Fatalf("output = %q, want rust/cpp/go detail rows", got)
		}
		if rustIndex >= cppIndex || cppIndex >= goIndex {
			t.Fatalf("repo-files order = %q, want rust before cpp before go", got)
		}
	})

	t.Run("sort without detail errors", func(t *testing.T) {
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetArgs([]string{"--sort", "governance"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "requires --detail") {
			t.Fatalf("Execute() error = %v, want --detail requirement", err)
		}
	})

	t.Run("invalid sort errors", func(t *testing.T) {
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetArgs([]string{"--detail", "--sort", "bogus"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --sort") {
			t.Fatalf("Execute() error = %v, want invalid sort error", err)
		}
	})

	t.Run("min governance filter", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--detail", "--min-governance", "rich"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := buf.String()
		if !strings.Contains(got, "rust\tdesc=") || strings.Contains(got, "cpp\tdesc=") || strings.Contains(got, "go\tdesc=") {
			t.Fatalf("min-governance output = %q, want only rust detail row", got)
		}
	})

	t.Run("repo asset filter", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--detail", "--has-repo-asset", "security"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := buf.String()
		if !strings.Contains(got, "rust\tdesc=") || strings.Contains(got, "cpp\tdesc=") || strings.Contains(got, "go\tdesc=") {
			t.Fatalf("repo-asset output = %q, want only rust detail row", got)
		}
	})

	t.Run("filters without detail errors", func(t *testing.T) {
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetArgs([]string{"--min-governance", "standard"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "filters require --detail") {
			t.Fatalf("Execute() error = %v, want detail filter requirement", err)
		}
	})

	t.Run("invalid governance filter errors", func(t *testing.T) {
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetArgs([]string{"--detail", "--min-governance", "elite"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --min-governance") {
			t.Fatalf("Execute() error = %v, want invalid governance filter error", err)
		}
	})

	t.Run("invalid repo asset filter errors", func(t *testing.T) {
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetArgs([]string{"--detail", "--has-repo-asset", "license"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid --has-repo-asset") {
			t.Fatalf("Execute() error = %v, want invalid repo asset error", err)
		}
	})

	t.Run("table without detail errors", func(t *testing.T) {
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetArgs([]string{"--table"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "--table requires --detail") {
			t.Fatalf("Execute() error = %v, want table detail requirement", err)
		}
	})

	t.Run("table toml rejected", func(t *testing.T) {
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetArgs([]string{"--detail", "--table", "--toml"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "table output is only supported for text format") {
			t.Fatalf("Execute() error = %v, want table/toml error", err)
		}
	})

	t.Run("table compact rejected", func(t *testing.T) {
		cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
		cmd.SetArgs([]string{"--detail", "--table", "--compact"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "--table cannot be combined with --compact") {
			t.Fatalf("Execute() error = %v, want table/compact error", err)
		}
	})
}

func TestListCmd_PropagatesCatalogErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "langs", args: nil},
		{name: "summaries", args: []string{"--detail"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewListCommand(newTestDependencies(func() *appcatalog.Service { return newFailingCatalogService(t) }))
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() expected error, got nil")
			}
			if !strings.Contains(err.Error(), "failed to read templates") {
				t.Fatalf("Execute() error = %v, want template read error", err)
			}
		})
	}
}

func TestListCmdRejectsPositionalArgs(t *testing.T) {
	t.Parallel()

	cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"go"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want unknown-argument rejection")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("Execute() error = %v, want unknown command error", err)
	}
}
