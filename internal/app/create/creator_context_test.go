package create

import (
	"bytes"
	"context"
	"testing"
	"testing/fstest"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

func TestCreatorCreate_HandsGitTheCallerContext(t *testing.T) {
	t.Parallel()

	type gitCall struct {
		ctx  context.Context
		args []string
	}

	var calls []gitCall
	creator := NewCreatorWithDeps(fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\n")},
		"go/main.go.tmpl":                    {Data: []byte("package main\n")},
	}, &bytes.Buffer{}, func(ctx context.Context, _ string, args ...string) error {
		calls = append(calls, gitCall{ctx: ctx, args: args})
		return nil
	}, nil)

	ctx, cancel := context.WithCancel(t.Context())
	opts := Options{
		Lang:        "go",
		ProjectName: "demo",
		TargetDir:   t.TempDir() + "/demo",
		GitMode:     domain.GitModeInitCommit,
	}

	if err := creator.Create(ctx, opts); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(calls) == 0 {
		t.Fatal("git was never invoked")
	}

	cancel()
	for _, call := range calls {
		if call.ctx.Err() == nil {
			t.Fatalf("git %v received a context that does not observe the caller's cancel", call.args)
		}
	}
}

// A cancelled `go env` must degrade to the built-in version rather than
// leaving GoVersion empty or failing the scaffold.
func TestDetectGoVersion_FallsBackWhenContextIsCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	got := detectGoVersion(ctx)
	if got == "" {
		t.Fatal("detectGoVersion() = \"\", want the runtime fallback version")
	}

	want, err := parseGoLanguageVersion(runtimeVersion())
	if err != nil {
		t.Fatalf("parseGoLanguageVersion() error = %v", err)
	}
	if got != want {
		t.Fatalf("detectGoVersion() = %q, want runtime fallback %q", got, want)
	}
}
