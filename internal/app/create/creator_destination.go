package create

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	osStat      = os.Stat
	osReadDir   = os.ReadDir
	osRemoveAll = os.RemoveAll
	osGetwd     = os.Getwd
	filepathAbs = filepath.Abs
)

func (c *Creator) checkDestDir(opts Options) error {
	return c.inspectDestDir(opts, false)
}

func (c *Creator) preflightDestDir(opts Options) error {
	return c.inspectDestDir(opts, true)
}

func (c *Creator) inspectDestDir(opts Options, previewOnly bool) error {
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

	if opts.Force {
		return c.handleForcedDestination(targetDir, previewOnly)
	}

	return validateReusableDestination(targetDir, opts.AllowExistingEmptyDir)
}

func (c *Creator) handleForcedDestination(targetDir string, previewOnly bool) error {
	currentDir, err := isCurrentDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to inspect destination %q: %w", targetDir, err)
	}
	if currentDir {
		return fmt.Errorf("refusing to remove current directory %q with --force", targetDir)
	}
	if previewOnly {
		return nil
	}

	_, _ = fmt.Fprintf(c.w, "Warning: directory %q already exists, removing due to --force\n", targetDir)
	if err := osRemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to remove existing directory %q: %w", targetDir, err)
	}

	return nil
}

func validateReusableDestination(targetDir string, allowExistingEmptyDir bool) error {
	empty, err := isEmptyDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to inspect destination %q: %w", targetDir, err)
	}

	if empty && allowExistingEmptyDir {
		return nil
	}

	if empty {
		return fmt.Errorf("directory %q already exists; use --force to overwrite", targetDir)
	}

	return fmt.Errorf("directory %q already exists and is not empty", targetDir)
}

func isEmptyDir(dir string) (bool, error) {
	entries, err := osReadDir(dir)
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}

func isCurrentDir(targetDir string) (bool, error) {
	absTarget, err := filepathAbs(targetDir)
	if err != nil {
		return false, err
	}

	cwd, err := osGetwd()
	if err != nil {
		return false, err
	}

	return absTarget == cwd, nil
}
