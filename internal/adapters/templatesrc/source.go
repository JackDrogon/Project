package templatesrc

import (
	"io/fs"
)

type Source struct {
	fs fs.FS
}

func New() *Source {
	return &Source{fs: FS}
}

func (s *Source) FS() fs.FS {
	return s.fs
}

func ModeForPath(sourcePath string) (fs.FileMode, bool) {
	return LookupMode(sourcePath)
}
