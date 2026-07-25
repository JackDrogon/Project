package scaffold

import (
	"io"

	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func newTestDependencies(creator *appcreate.Creator) Dependencies {
	return Dependencies{
		NewCreator: func(io.Writer) *appcreate.Creator { return creator },
		NewService: appcreate.NewService,
	}
}
