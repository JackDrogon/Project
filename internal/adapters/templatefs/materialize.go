package templatefs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

const (
	// defaultDirMode is the standard permission for created directories (rwxr-xr-x).
	defaultDirMode fs.FileMode = 0o755
	// defaultFileMode is the standard permission for regular files (rw-r--r--).
	defaultFileMode fs.FileMode = 0o644
	// tempFileMode is the initial permission for files during creation (rw-rw-rw-).
	tempFileMode fs.FileMode = 0o666
	// tempDirMode is the permission for parent directories during creation (rwxrwxrwx).
	tempDirMode fs.FileMode = 0o777
	// ownerWriteMask ensures directory owner has write permission (rwx------).
	ownerWriteMask fs.FileMode = 0o700
	// maxConcurrentWrites limits parallel file write operations.
	maxConcurrentWrites = 8
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
	if err := osMkdirAll(destDir, defaultDirMode); err != nil {
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
		if err := osMkdirAll(entry.Destination, ensureWritableDirMode(mode)); err != nil {
			return err
		}
		pendingDirModes = append(pendingDirModes, pendingDirMode{path: entry.Destination, mode: mode})
	}

	if err := materializeFilesConcurrently(w, fsys, files, vars, resolveMode); err != nil {
		return err
	}

	for i := len(pendingDirModes) - 1; i >= 0; i-- {
		if err := applyMaterializedMode(pendingDirModes[i].path, pendingDirModes[i].mode); err != nil {
			return err
		}
	}

	return nil
}

func materializeFilesConcurrently(w io.Writer, fsys fs.FS, files []Entry, vars domain.TemplateVars, resolveMode ModeResolver) error {
	if len(files) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, maxConcurrentWrites)
	errCh := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(entry Entry) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if err := materializeFile(fsys, entry, vars, resolveMode); err != nil {
				errCh <- err
				return
			}

			mu.Lock()
			_, _ = fmt.Fprintf(w, "  create %s\n", entry.Destination)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return err
	}

	return nil
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

	if err := osMkdirAll(filepath.Dir(entry.Destination), tempDirMode); err != nil {
		return err
	}

	if err := osWriteFile(entry.Destination, rendered, tempFileMode); err != nil {
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
		return defaultDirMode
	}

	return mode.Perm() | ownerWriteMask
}
