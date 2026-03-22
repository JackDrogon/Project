package main

import (
	"io"
	"io/fs"

	"github.com/JackDrogon/project/internal/adapters/templatesrc"
	appcreate "github.com/JackDrogon/project/internal/app/create"
)

type commandCreator = appcreate.Creator

func newCommandCreator(out io.Writer) *commandCreator {
	source := templatesrc.New()
	resolveMode := func(sourcePath string, isDir bool) (fs.FileMode, bool) {
		return templatesrc.ModeForPath(sourcePath)
	}
	return appcreate.NewCreatorWithDeps(source.FS(), out, nil, resolveMode)
}

func newCommandCreatorFromFS(sourceFS fs.FS, out io.Writer) *commandCreator {
	return appcreate.NewCreator(sourceFS, out)
}
