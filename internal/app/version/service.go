package version

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
