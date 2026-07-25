package scaffold

import (
	"io"

	appcreate "github.com/JackDrogon/project/internal/app/create"
)

type Dependencies struct {
	// NewCreator builds the creator bound to the writer the command owns, so
	// scaffold progress output follows the command's out stream.
	NewCreator func(out io.Writer) *appcreate.Creator
	NewService func() *appcreate.Service
}

func (d Dependencies) newCreator(out io.Writer) *appcreate.Creator {
	if d.NewCreator == nil {
		panic("scaffold dependencies require NewCreator")
	}

	return d.NewCreator(out)
}

func (d Dependencies) newService() *appcreate.Service {
	if d.NewService == nil {
		panic("scaffold dependencies require NewService")
	}

	return d.NewService()
}
