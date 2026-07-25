package completion

import (
	"bytes"
	"strings"
	"testing"

	appcompletion "github.com/JackDrogon/project/internal/app/completion"
	"github.com/spf13/cobra"
)

func newTestCommand() *cobra.Command {
	return NewCommand(Dependencies{NewService: appcompletion.NewService})
}

func TestCommand_GeneratesAllShells(t *testing.T) {
	tests := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range tests {
		t.Run(shell, func(t *testing.T) {
			var buf bytes.Buffer
			root := &cobra.Command{Use: "project"}
			root.SetOut(&buf)
			root.SetErr(&buf)
			root.AddCommand(newTestCommand())
			root.SetArgs([]string{"completion", shell})

			if err := root.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if !strings.Contains(buf.String(), "project") {
				t.Fatalf("output = %q, want completion content", buf.String())
			}
		})
	}
}

func TestCommand_RejectsInvalidShell(t *testing.T) {
	var buf bytes.Buffer
	root := &cobra.Command{Use: "project"}
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.AddCommand(newTestCommand())
	root.SetArgs([]string{"completion", "invalid"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid argument \"invalid\"") {
		t.Fatalf("Execute() error = %v, want invalid argument error", err)
	}
}

func TestDependenciesRequireNewService(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("recover() = nil, want panic")
		}
	}()

	_ = NewCommand(Dependencies{})
}
