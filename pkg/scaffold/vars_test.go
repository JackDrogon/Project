package scaffold

import (
	"testing"
)

func TestNewTemplateVars(t *testing.T) {
	t.Run("with module path", func(t *testing.T) {
		vars := NewTemplateVars("myproj", "github.com/user/myproj")
		if vars.ProjectName != "myproj" {
			t.Errorf("ProjectName = %q, want %q", vars.ProjectName, "myproj")
		}
		if vars.ModulePath != "github.com/user/myproj" {
			t.Errorf("ModulePath = %q, want %q", vars.ModulePath, "github.com/user/myproj")
		}
		if vars.Year == 0 {
			t.Error("Year should not be 0")
		}
	})

	t.Run("without module path defaults to project name", func(t *testing.T) {
		vars := NewTemplateVars("myproj", "")
		if vars.ModulePath != "myproj" {
			t.Errorf("ModulePath = %q, want %q", vars.ModulePath, "myproj")
		}
	})
}
