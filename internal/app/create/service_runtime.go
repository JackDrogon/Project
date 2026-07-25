package create

import (
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appconfig "github.com/JackDrogon/project/internal/app/config"
)

var reservedSetKeys = map[string]struct{}{
	"lang":         {},
	"project_name": {},
	"target_dir":   {},
	"module":       {},
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

	runtime := Runtime{
		Command:             expected,
		ActiveConfig:        flags.ActiveConfig,
		ExplicitSetValues:   maps.Clone(templateInputValues),
		TemplateInputValues: templateInputValues,
	}

	if flags.WriteReplayPath != "" && flags.DryRun {
		return Runtime{}, errors.New("--write-replay cannot be combined with --dry-run")
	}

	if flags.ReplayPath == "" {
		if inputDefaults := activeConfigInputs(expected, flags.ActiveConfig); len(inputDefaults) > 0 {
			runtime.TemplateInputValues = mergeInputMaps(inputDefaults, templateInputValues)
		}
		return runtime, nil
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

	runtime.Replay = replay
	runtime.HasReplay = true
	runtime.TemplateInputValues = mergedInputs

	return runtime, nil
}

func activeConfigInputs(command Command, active appconfig.ActiveConfig) map[string]string {
	config := active.Config
	if config == nil {
		return nil
	}

	switch command {
	case CommandNew:
		if config.New == nil || len(config.New.Inputs) == 0 {
			return nil
		}
		return maps.Clone(config.New.Inputs)
	case CommandInit:
		if config.Init == nil || len(config.Init.Inputs) == 0 {
			return nil
		}
		return maps.Clone(config.Init.Inputs)
	default:
		return nil
	}
}

func mergeInputMaps(base, overrides map[string]string) map[string]string {
	if len(base) == 0 && len(overrides) == 0 {
		return nil
	}

	merged := make(map[string]string, len(base)+len(overrides))
	maps.Copy(merged, base)
	maps.Copy(merged, overrides)

	return merged
}
