package main

import (
	"io"
	"io/fs"

	"github.com/JackDrogon/project/internal/adapters/templatesrc"
	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func newCommandCreator(out io.Writer) *appcreate.Creator {
	source := templatesrc.New()
	resolveMode := func(sourcePath string, isDir bool) (fs.FileMode, bool) {
		return templatesrc.ModeForPath(sourcePath)
	}
	return appcreate.NewCreatorWithDeps(source.FS(), out, nil, resolveMode)
}

func newCommandCreatorFromFS(sourceFS fs.FS, out io.Writer) *appcreate.Creator {
	return appcreate.NewCreator(sourceFS, out)
}
