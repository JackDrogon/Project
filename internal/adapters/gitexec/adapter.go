package gitexec

import (
	"fmt"
	"os/exec"
)

var execCommand = exec.Command

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Run(dir string, args ...string) error {
	cmd := execCommand("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %w\n%s", args[0], err, string(output))
	}
	return nil
}
