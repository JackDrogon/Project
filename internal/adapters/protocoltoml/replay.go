package protocoltoml

import (
	"bytes"
	"fmt"
	"os"

	domain "github.com/JackDrogon/project/internal/scaffold"
	toml "github.com/pelletier/go-toml/v2"
)

const (
	ReplayFilename = ".project-replay.toml"
	ReplayVersion  = 2
)

type Replay struct {
	Version  int               `toml:"version"`
	Mode     string            `toml:"mode"`
	Template ReplayTemplate    `toml:"template"`
	Project  ReplayProject     `toml:"project"`
	Git      ReplayGit         `toml:"git"`
	Options  ReplayOptions     `toml:"options"`
	Inputs   map[string]string `toml:"inputs"`
}

type ReplayTemplate struct {
	Lang string `toml:"lang"`
}

type ReplayProject struct {
	Name       string `toml:"name"`
	TargetDir  string `toml:"target_dir"`
	ModulePath string `toml:"module_path"`
}

type ReplayGit struct {
	Mode    domain.GitMode `toml:"mode"`
	Signoff bool           `toml:"signoff"`
}

type ReplayOptions struct {
	Force bool `toml:"force"`
}

func ReadReplay(path string) (Replay, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Replay{}, fmt.Errorf("failed to read replay file %s: %w", path, err)
	}
	return DecodeReplay(content, path)
}

func WriteReplay(path string, replay Replay) error {
	content, err := MarshalReplay(path, replay)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("failed to write replay file %s: %w", path, err)
	}
	return nil
}

func DecodeReplay(content []byte, path string) (Replay, error) {
	if err := rejectLegacyJSON(content, path); err != nil {
		return Replay{}, err
	}

	var replay Replay
	decoder := toml.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&replay); err != nil {
		return Replay{}, fmt.Errorf("failed to decode replay file %s: %w", path, err)
	}
	if err := ValidateReplay(replay, path); err != nil {
		return Replay{}, err
	}

	return replay, nil
}

func MarshalReplay(path string, replay Replay) ([]byte, error) {
	normalized := NormalizeReplayForWrite(replay)
	if err := ValidateReplay(normalized, path); err != nil {
		return nil, err
	}

	content, err := toml.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("failed to encode replay file %s: %w", path, err)
	}

	return append(content, '\n'), nil
}

func NormalizeReplayForWrite(replay Replay) Replay {
	if replay.Version == 0 {
		replay.Version = ReplayVersion
	}
	if replay.Inputs == nil {
		replay.Inputs = map[string]string{}
	}
	return replay
}

func ValidateReplay(replay Replay, path string) error {
	if replay.Version != ReplayVersion {
		return fmt.Errorf("replay file %s has unsupported version %d", path, replay.Version)
	}

	switch replay.Mode {
	case "new", "init":
	default:
		return fmt.Errorf("replay file %s has unsupported mode %q", path, replay.Mode)
	}

	if replay.Template.Lang == "" {
		return fmt.Errorf("replay file %s template.lang must not be empty", path)
	}
	if replay.Project.Name == "" {
		return fmt.Errorf("replay file %s project.name must not be empty", path)
	}
	if replay.Project.TargetDir == "" {
		return fmt.Errorf("replay file %s project.target_dir must not be empty", path)
	}

	switch replay.Git.Mode {
	case domain.GitModeNone, domain.GitModeInitOnly, domain.GitModeInitCommit:
	default:
		return fmt.Errorf(
			"replay file %s git.mode %q is invalid: must be one of %s, %s, %s",
			path,
			replay.Git.Mode,
			domain.GitModeNone,
			domain.GitModeInitOnly,
			domain.GitModeInitCommit,
		)
	}

	if replay.Inputs == nil {
		return fmt.Errorf("replay file %s inputs must be a table", path)
	}

	return nil
}
