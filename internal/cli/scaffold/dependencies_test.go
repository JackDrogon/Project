package scaffold

import (
	"bytes"
	"io"
	"testing"
	"testing/fstest"

	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func TestDependenciesRequireNewCreator(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("recover() = nil, want panic")
		}
	}()

	Dependencies{}.newCreator(io.Discard)
}

func TestDependenciesRequireNewService(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("recover() = nil, want panic")
		}
	}()

	Dependencies{NewCreator: func(io.Writer) *appcreate.Creator { return &appcreate.Creator{} }}.newService()
}

// The command owns its output stream: the creator must be built against it
// rather than against a writer captured at wiring time.
func TestNewCommand_BuildsCreatorFromCommandOut(t *testing.T) {
	var got io.Writer
	var out bytes.Buffer
	deps := Dependencies{
		NewCreator: func(w io.Writer) *appcreate.Creator {
			got = w
			return appcreate.NewCreatorWithDeps(fstest.MapFS{"go/go.mod.tmpl": {Data: []byte("module x\n")}}, w, func(string, ...string) error { return nil }, nil)
		},
		NewService: appcreate.NewService,
	}

	cmd := NewNewCommand(deps)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--dry-run", "demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got != &out {
		t.Fatalf("NewCreator received %#v, want the command's out writer", got)
	}
	if out.Len() == 0 {
		t.Fatal("command out is empty, want scaffold output routed through it")
	}
}
