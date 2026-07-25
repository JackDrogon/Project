package catalog

import (
	"io/fs"
	"strings"
)

type failingFS struct{ err error }

func (f failingFS) Open(string) (fs.File, error)          { return nil, f.err }
func (f failingFS) ReadDir(string) ([]fs.DirEntry, error) { return nil, f.err }

func filepathToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
