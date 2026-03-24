package scaffold

import appcreate "github.com/JackDrogon/project/internal/app/create"

func newTestDependencies(creator *appcreate.Creator) Dependencies {
	return Dependencies{Creator: creator, NewService: appcreate.NewService}
}
