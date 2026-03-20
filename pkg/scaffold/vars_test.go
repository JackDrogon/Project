package scaffold

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"testing"
)

func stubTemplateVarFuncs(t *testing.T) {
	t.Helper()

	oldCurrentUser := currentUser
	oldExecGoCommand := execGoCommand
	oldRuntimeVersion := runtimeVersion
	t.Cleanup(func() {
		currentUser = oldCurrentUser
		execGoCommand = oldExecGoCommand
		runtimeVersion = oldRuntimeVersion
	})
}

func TestNewTemplateVars(t *testing.T) {
	t.Run("with module path", func(t *testing.T) {
		stubTemplateVarFuncs(t)
		execGoCommand = func(name string, args ...string) *exec.Cmd {
			return helperCommand(t, "printf", "go1.26.1")
		}
		runtimeVersion = func() string { return "go1.25.3" }

		vars := NewTemplateVars("myproj", "github.com/user/myproj")
		if vars.ProjectName != "myproj" {
			t.Errorf("ProjectName = %q, want %q", vars.ProjectName, "myproj")
		}
		if vars.ModulePath != "github.com/user/myproj" {
			t.Errorf("ModulePath = %q, want %q", vars.ModulePath, "github.com/user/myproj")
		}
		if vars.ProjectNameLower != "myproj" {
			t.Errorf("ProjectNameLower = %q, want %q", vars.ProjectNameLower, "myproj")
		}
		if vars.GoVersion != "1.26" {
			t.Errorf("GoVersion = %q, want %q", vars.GoVersion, "1.26")
		}
		if vars.Year == 0 {
			t.Error("Year should not be 0")
		}
	})

	t.Run("without module path defaults to project name", func(t *testing.T) {
		stubTemplateVarFuncs(t)
		execGoCommand = func(name string, args ...string) *exec.Cmd {
			return helperCommand(t, "printf", "go1.24.0")
		}

		vars := NewTemplateVars("myproj", "")
		if vars.ModulePath != "myproj" {
			t.Errorf("ModulePath = %q, want %q", vars.ModulePath, "myproj")
		}
		if vars.ProjectNameLower != "myproj" {
			t.Errorf("ProjectNameLower = %q, want %q", vars.ProjectNameLower, "myproj")
		}
		if vars.GoVersion != "1.24" {
			t.Errorf("GoVersion = %q, want %q", vars.GoVersion, "1.24")
		}
	})

	t.Run("falls back to runtime version when go env fails", func(t *testing.T) {
		stubTemplateVarFuncs(t)
		execGoCommand = func(name string, args ...string) *exec.Cmd {
			return helperCommand(t, "exit1")
		}
		runtimeVersion = func() string { return "go1.23.7" }

		vars := NewTemplateVars("myproj", "")
		if vars.GoVersion != "1.23" {
			t.Errorf("GoVersion = %q, want %q", vars.GoVersion, "1.23")
		}
	})

	t.Run("falls back to default author when current user lookup fails", func(t *testing.T) {
		stubTemplateVarFuncs(t)
		currentUser = func() (*user.User, error) {
			return nil, errors.New("lookup failed")
		}
		execGoCommand = func(name string, args ...string) *exec.Cmd {
			return helperCommand(t, "printf", "go1.24.0")
		}

		vars := NewTemplateVars("myproj", "")
		if vars.Author != "author" {
			t.Fatalf("Author = %q, want %q", vars.Author, "author")
		}
	})

	t.Run("falls back to default author when username is empty", func(t *testing.T) {
		stubTemplateVarFuncs(t)
		currentUser = func() (*user.User, error) {
			return &user.User{}, nil
		}
		execGoCommand = func(name string, args ...string) *exec.Cmd {
			return helperCommand(t, "printf", "go1.24.0")
		}

		vars := NewTemplateVars("myproj", "")
		if vars.Author != "author" {
			t.Fatalf("Author = %q, want %q", vars.Author, "author")
		}
	})
}

func TestDetectGoVersion_FallsBackToTrimmedRuntimeString(t *testing.T) {
	stubTemplateVarFuncs(t)
	execGoCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand(t, "exit1")
	}
	runtimeVersion = func() string { return "  devel custom-build  " }

	if got := detectGoVersion(); got != "devel custom-build" {
		t.Fatalf("detectGoVersion() = %q, want %q", got, "devel custom-build")
	}
}

func TestParseGoLanguageVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "go env output", input: "go1.26.1\n", want: "1.26"},
		{name: "release candidate", input: "go1.27rc1", want: "1.27"},
		{name: "runtime version", input: "go1.25.3", want: "1.25"},
		{name: "devel version", input: "devel go1.27-abcdef", want: "1.27"},
		{name: "invalid", input: "not-a-go-version", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGoLanguageVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("parseGoLanguageVersion() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseGoLanguageVersion() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("parseGoLanguageVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func helperCommand(t *testing.T, mode string, payload ...string) *exec.Cmd {
	t.Helper()

	args := []string{"-test.run=TestHelperProcess", "--", mode}
	args = append(args, payload...)
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if len(os.Args) < 4 || os.Args[2] != "--" {
		return
	}

	switch os.Args[3] {
	case "printf":
		_, _ = os.Stdout.WriteString(os.Args[4])
	case "exit1":
		os.Exit(1)
	}
	os.Exit(0)
}
