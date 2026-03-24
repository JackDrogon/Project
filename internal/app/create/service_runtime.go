package create

import (
	"fmt"
	"strings"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
)

var reservedSetKeys = map[string]struct{}{
	"lang":         {},
	"project_name": {},
	"target_dir":   {},
	"module_path":  {},
	"git_mode":     {},
	"signoff":      {},
	"force":        {},
	"dry_run":      {},
}

func (s *Service) ParseSetValues(flags Flags) (map[string]string, error) {
	values := make(map[string]string, len(flags.SetValues))
	for _, raw := range flags.SetValues {
		key, value, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --set value %q: must be key=value", raw)
		}
		if key == "" {
			return nil, fmt.Errorf("invalid --set value %q: key must not be empty", raw)
		}
		if _, reserved := reservedSetKeys[key]; reserved {
			return nil, fmt.Errorf("invalid --set key %q: reserved for command options", key)
		}
		if _, exists := values[key]; exists {
			return nil, fmt.Errorf("invalid --set key %q: specified more than once", key)
		}
		values[key] = value
	}

	return values, nil
}

func (s *Service) RuntimeState(flags Flags, expected Command) (Runtime, error) {
	templateInputValues, err := s.ParseSetValues(flags)
	if err != nil {
		return Runtime{}, err
	}

	if flags.WriteReplayPath != "" && flags.DryRun {
		return Runtime{}, fmt.Errorf("--write-replay cannot be combined with --dry-run")
	}

	if flags.ReplayPath == "" {
		return Runtime{TemplateInputValues: templateInputValues}, nil
	}

	replay, err := protocoltoml.ReadReplay(flags.ReplayPath)
	if err != nil {
		return Runtime{}, err
	}
	if replay.Mode != string(expected) {
		return Runtime{}, fmt.Errorf(
			"invalid --replay %q: replay command %q does not match %q",
			flags.ReplayPath,
			replay.Mode,
			expected,
		)
	}

	mergedInputs := mergeReplayInputs(replay, templateInputValues)

	return Runtime{Replay: replay, HasReplay: true, TemplateInputValues: mergedInputs}, nil
}
