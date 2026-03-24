package catalog

import appcatalog "github.com/JackDrogon/project/internal/app/catalog"

type Dependencies struct {
	NewService func() *appcatalog.Service
}

func (d Dependencies) newService() *appcatalog.Service {
	if d.NewService == nil {
		panic("catalog dependencies require NewService")
	}

	return d.NewService()
}
