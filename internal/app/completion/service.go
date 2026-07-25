// Package completion resolves completion command settings from the active
// CLI configuration.
package completion

import appconfig "github.com/JackDrogon/project/internal/app/config"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// ResolveShell picks the target shell: an explicit argument wins, then the
// [completion] config section; empty means no shell was provided.
func (s *Service) ResolveShell(arg string, hasArg bool, active appconfig.ActiveConfig) string {
	if hasArg {
		return arg
	}

	if active.Config == nil || active.Config.Completion == nil || active.Config.Completion.Shell == nil {
		return ""
	}

	return *active.Config.Completion.Shell
}
