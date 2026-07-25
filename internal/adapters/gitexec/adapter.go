package gitexec

import (
	"context"
	"fmt"
	"os/exec"
)

// commandFunc builds the child process. It is a field rather than a package
// variable so tests substituting a helper process do not mutate global state,
// which is what kept this package from running its tests in parallel.
type commandFunc func(ctx context.Context, name string, arg ...string) *exec.Cmd

type Adapter struct {
	command commandFunc
}

func New() *Adapter {
	return &Adapter{command: exec.CommandContext}
}

// newWithCommand builds an adapter around a substitute command factory.
func newWithCommand(command commandFunc) *Adapter {
	if command == nil {
		command = exec.CommandContext
	}

	return &Adapter{command: command}
}

// Run executes git inside dir. Cancelling ctx kills the child process, so an
// interrupted scaffold does not leave a git subprocess running behind it.
func (a *Adapter) Run(ctx context.Context, dir string, args ...string) error {
	cmd := a.command(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("git %s canceled: %w", args[0], ctxErr)
		}
		return fmt.Errorf("git %s failed: %w\n%s", args[0], err, string(output))
	}
	return nil
}
