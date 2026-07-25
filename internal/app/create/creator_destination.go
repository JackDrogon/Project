package create

import (
	"errors"
	"fmt"
	"os"
)

var (
	osStat    = os.Stat
	osReadDir = os.ReadDir
)

// checkDestDir reports whether the destination can receive a scaffold.
// Scaffolding never removes existing files: a non-empty destination is always
// rejected so deleting anything stays an explicit user action.
func (c *Creator) checkDestDir(opts Options) error {
	targetDir := opts.DestinationDir()

	info, err := osStat(targetDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to inspect destination %q: %w", targetDir, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("destination %q already exists and is not a directory", targetDir)
	}

	empty, err := isEmptyDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to inspect destination %q: %w", targetDir, err)
	}
	if !empty {
		return fmt.Errorf("directory %q already exists and is not empty; remove it manually before scaffolding", targetDir)
	}

	// An empty directory is safe to populate, but `new` still demands an explicit
	// opt-in so a mistyped project name cannot fill an unrelated directory.
	if opts.Force || opts.AllowExistingEmptyDir {
		return nil
	}

	return fmt.Errorf("directory %q already exists; pass --force to scaffold into it", targetDir)
}

func isEmptyDir(dir string) (bool, error) {
	entries, err := osReadDir(dir)
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}
