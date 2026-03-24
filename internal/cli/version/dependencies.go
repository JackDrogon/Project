package version

import appversion "github.com/JackDrogon/project/internal/app/version"

type Dependencies struct {
	NewService func() *appversion.Service
}

func (d Dependencies) newService() *appversion.Service {
	if d.NewService == nil {
		panic("version dependencies require NewService")
	}

	return d.NewService()
}
