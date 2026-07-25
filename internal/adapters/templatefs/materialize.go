package templatefs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"golang.org/x/sync/errgroup"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

const (
	// defaultDirMode is the standard permission for created directories (rwxr-xr-x).
	defaultDirMode fs.FileMode = 0o755
	// defaultFileMode is the standard permission for regular files (rw-r--r--).
	defaultFileMode fs.FileMode = 0o644
	// tempFileMode is the permission a file is created with, before
	// applyMaterializedMode sets its final mode. It is deliberately not 0o666:
	// creating world-writable and narrowing afterwards leaves a window where a
	// file in a shared directory is writable by anyone, and umask is not a
	// guarantee - a caller running with umask 0 got exactly that file.
	tempFileMode fs.FileMode = 0o644
	// tempDirMode is the permission for parent directories created on demand.
	// The owner-write bit is all that is needed to populate them; the final
	// pass in Materialize tightens the directories it walked.
	tempDirMode fs.FileMode = 0o755
	// ownerWriteMask ensures directory owner has write permission (rwx------).
	ownerWriteMask fs.FileMode = 0o700
	// maxConcurrentWrites limits parallel file write operations.
	maxConcurrentWrites = 8
	// goosWindows matches runtime.GOOS on Windows, which has no POSIX modes.
	goosWindows = "windows"
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
	if err := os.MkdirAll(destDir, defaultDirMode); err != nil {
		return err
	}

	type pendingDirMode struct {
		path string
		mode fs.FileMode
	}

	var dirs []Entry
	var files []Entry
	var pendingDirModes []pendingDirMode

	err := WalkEntries(fsys, srcDir, destDir, vars, func(entry Entry) error {
		if entry.IsDir {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, entry := range dirs {
		_, _ = fmt.Fprintf(w, "  create %s/\n", entry.Destination)
		mode := resolvedMode(entry, true, resolveMode)
		if err := os.MkdirAll(entry.Destination, ensureWritableDirMode(mode)); err != nil {
			return err
		}
		pendingDirModes = append(pendingDirModes, pendingDirMode{path: entry.Destination, mode: mode})
	}

	if err := materializeFilesConcurrently(w, fsys, files, vars, resolveMode); err != nil {
		return err
	}

	// Deepest first: tightening a parent's mode before its children are written
	// would make the children unwritable.
	for _, dir := range slices.Backward(pendingDirModes) {
		if err := applyMaterializedMode(dir.path, dir.mode); err != nil {
			return err
		}
	}

	return nil
}

// materializeFilesConcurrently writes the rendered templates in parallel but
// reports them in walk order.
//
// Reporting used to happen inside each goroutine under a mutex, which made the
// "create <path>" listing come out in completion order - so two identical runs
// produced different output, defeating diffable logs and stable golden tests.
// Writes stay concurrent; only the reporting is serialized, and files that were
// written before an error are still listed so a failed run says what it left
// behind (scaffolding never deletes, so the caller has to clean up by hand).
func materializeFilesConcurrently(w io.Writer, fsys fs.FS, files []Entry, vars domain.TemplateVars, resolveMode ModeResolver) error {
	if len(files) == 0 {
		return nil
	}

	// Each goroutine writes its own index, so no synchronization is needed.
	written := make([]bool, len(files))

	var group errgroup.Group
	group.SetLimit(maxConcurrentWrites)

	for i, file := range files {
		group.Go(func() error {
			if err := materializeFile(fsys, file, vars, resolveMode); err != nil {
				return err
			}
			written[i] = true
			return nil
		})
	}

	err := group.Wait()

	for i, file := range files {
		if written[i] {
			_, _ = fmt.Fprintf(w, "  create %s\n", file.Destination)
		}
	}

	return err
}

func materializeFile(fsys fs.FS, entry Entry, vars domain.TemplateVars, resolveMode ModeResolver) error {
	loaded, err := ReadEntry(fsys, entry)
	if err != nil {
		return err
	}

	rendered, err := RenderEntry(loaded, vars)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(entry.Destination), tempDirMode); err != nil {
		return err
	}

	if err := os.WriteFile(entry.Destination, rendered, tempFileMode); err != nil {
		return err
	}

	return applyMaterializedMode(entry.Destination, resolvedMode(entry, false, resolveMode))
}

func resolvedMode(entry Entry, isDir bool, resolveMode ModeResolver) fs.FileMode {
	if resolveMode != nil {
		if mode, ok := resolveMode(entry.SourcePath, isDir); ok {
			return mode.Perm()
		}
	}
	if isDir {
		return defaultDirMode
	}
	return defaultFileMode
}

func applyMaterializedMode(path string, mode fs.FileMode) error {
	if mode.Perm() == 0 {
		return nil
	}
	if err := os.Chmod(path, mode.Perm()); err != nil {
		// Windows has no POSIX permission bits, so chmod failing there is
		// expected rather than a scaffolding failure.
		if runtime.GOOS == goosWindows {
			return nil
		}
		return err
	}

	return nil
}

func ensureWritableDirMode(mode fs.FileMode) fs.FileMode {
	if mode.Perm() == 0 {
		return defaultDirMode
	}

	return mode.Perm() | ownerWriteMask
}
