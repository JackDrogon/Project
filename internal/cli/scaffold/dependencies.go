package scaffold

import appcreate "github.com/JackDrogon/project/internal/app/create"

type Dependencies struct {
	Creator    *appcreate.Creator
	NewService func() *appcreate.Service
}

func (d Dependencies) creator() *appcreate.Creator {
	if d.Creator == nil {
		panic("scaffold dependencies require Creator")
	}

	return d.Creator
}

func (d Dependencies) newService() *appcreate.Service {
	if d.NewService == nil {
		panic("scaffold dependencies require NewService")
	}

	return d.NewService()
}
