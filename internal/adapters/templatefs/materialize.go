package templatefs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	domain "github.com/JackDrogon/project/internal/domain/scaffold"
)

var (
	osMkdirAll  = os.MkdirAll
	osWriteFile = os.WriteFile
	osChmod     = os.Chmod
)

type ModeResolver func(sourcePath string, isDir bool) (fs.FileMode, bool)

func Preview(w io.Writer, fsys fs.FS, srcDir, destDir string, vars domain.TemplateVars) error {
	return WalkEntries(fsys, srcDir, destDir, vars, func(entry Entry) error {
		if entry.IsDir {
			_, _ = fmt.Fprintf(w, "  create %s/\n", entry.Destination)
			return nil
		}

		_, _ = fmt.Fprintf(w, "  create %s\n", entry.Destination)
		loaded, err := ReadEntry(fsys, entry)
		if err != nil {
			return err
		}
		_, err = RenderEntry(loaded, vars)
		return err
	})
}

func Materialize(w io.Writer, fsys fs.FS, srcDir, destDir string, vars domain.TemplateVars, resolveMode ModeResolver) error {
	if err := osMkdirAll(destDir, 0o755); err != nil {
		return err
	}

	type pendingDirMode struct {
		path string
		mode fs.FileMode
	}
	var pendingDirModes []pendingDirMode

	err := WalkEntries(fsys, srcDir, destDir, vars, func(entry Entry) error {
		_, _ = fmt.Fprintf(w, "  create %s\n", entry.Destination)

		if entry.IsDir {
			mode := resolvedMode(entry, true, resolveMode)
			if err := osMkdirAll(entry.Destination, ensureWritableDirMode(mode)); err != nil {
				return err
			}
			pendingDirModes = append(pendingDirModes, pendingDirMode{path: entry.Destination, mode: mode})
			return nil
		}

		loaded, err := ReadEntry(fsys, entry)
		if err != nil {
			return err
		}

		rendered, err := RenderEntry(loaded, vars)
		if err != nil {
			return err
		}

		if err := osMkdirAll(filepath.Dir(entry.Destination), 0o777); err != nil {
			return err
		}
		if err := osWriteFile(entry.Destination, rendered, 0o666); err != nil {
			return err
		}

		return applyMaterializedMode(entry.Destination, resolvedMode(entry, false, resolveMode))
	})
	if err != nil {
		return err
	}

	for i := len(pendingDirModes) - 1; i >= 0; i-- {
		if err := applyMaterializedMode(pendingDirModes[i].path, pendingDirModes[i].mode); err != nil {
			return err
		}
	}

	return nil
}

func resolvedMode(entry Entry, isDir bool, resolveMode ModeResolver) fs.FileMode {
	if resolveMode != nil {
		if mode, ok := resolveMode(entry.SourcePath, isDir); ok {
			return mode.Perm()
		}
	}
	if isDir {
		return 0o755
	}
	return 0o644
}

func applyMaterializedMode(path string, mode fs.FileMode) error {
	if mode.Perm() == 0 {
		return nil
	}
	if err := osChmod(path, mode.Perm()); err != nil {
		if runtime.GOOS == "windows" {
			return nil
		}
		return err
	}

	return nil
}

func ensureWritableDirMode(mode fs.FileMode) fs.FileMode {
	if mode.Perm() == 0 {
		return 0o755
	}

	return mode.Perm() | 0o700
}
