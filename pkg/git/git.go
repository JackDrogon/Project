package git

import (
	"fmt"
	"os/exec"
)

// Run executes a git command in the given directory.
func Run(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %w\n%s", args[0], err, string(output))
	}
	return nil
}
