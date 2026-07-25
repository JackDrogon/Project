package version

import appconfig "github.com/JackDrogon/project/internal/app/config"

type Provider interface {
	Info() string
	Verbose() string
}

type Service struct {
	provider Provider
}

func NewService(provider Provider) *Service {
	return &Service{provider: provider}
}

func (s *Service) Info() string {
	return s.provider.Info()
}

func (s *Service) Verbose() string {
	return s.provider.Verbose()
}

// ResolveVerbose applies the [version] config section default when the
// --verbose flag was not set explicitly.
func ResolveVerbose(verbose, changed bool, active appconfig.ActiveConfig) bool {
	if changed {
		return verbose
	}

	if active.Config == nil || active.Config.VersionCmd == nil || active.Config.VersionCmd.Verbose == nil {
		return verbose
	}

	return *active.Config.VersionCmd.Verbose
}
