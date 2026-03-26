package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
)

const defaultConfigFileName = "config.toml"

type Decoder func(content []byte, path string) (protocoltoml.Config, error)

type Dependencies struct {
	UserConfigDir func() (string, error)
	Join          func(elem ...string) string
	ReadFile      func(path string) ([]byte, error)
	Decode        Decoder
}

type Service struct {
	deps Dependencies
}

func NewService() *Service {
	return NewServiceWithDeps(DefaultDependencies())
}

func NewServiceWithDeps(deps Dependencies) *Service {
	return &Service{deps: deps.withDefaults()}
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		UserConfigDir: os.UserConfigDir,
		Join:          filepath.Join,
		ReadFile:      os.ReadFile,
		Decode:        protocoltoml.DecodeConfig,
	}
}

func (d Dependencies) withDefaults() Dependencies {
	defaults := DefaultDependencies()
	if d.UserConfigDir == nil {
		d.UserConfigDir = defaults.UserConfigDir
	}
	if d.Join == nil {
		d.Join = defaults.Join
	}
	if d.ReadFile == nil {
		d.ReadFile = defaults.ReadFile
	}
	if d.Decode == nil {
		d.Decode = defaults.Decode
	}
	return d
}

func (s *Service) LoadActiveConfig(ctx Context) (ActiveConfig, error) {
	if ctx.ExplicitPath != "" {
		return s.loadFromPath(SourceExplicit, ctx.ExplicitPath)
	}

	defaultPath, err := s.defaultConfigPath()
	if err != nil {
		return ActiveConfig{}, err
	}

	config, err := s.loadFromPath(SourceUserConfig, defaultPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ActiveConfig{Source: SourceNone}, nil
		}
		return ActiveConfig{}, err
	}

	return config, nil
}

func (s *Service) defaultConfigPath() (string, error) {
	configDir, err := s.deps.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return s.deps.Join(configDir, "project", defaultConfigFileName), nil
}

func (s *Service) loadFromPath(source Source, path string) (ActiveConfig, error) {
	content, err := s.deps.ReadFile(path)
	if err != nil {
		return ActiveConfig{}, fmt.Errorf("read %s %q: %w", source, path, err)
	}

	decoded, err := s.deps.Decode(content, path)
	if err != nil {
		return ActiveConfig{}, err
	}

	return ActiveConfig{Source: source, Path: path, Config: &decoded}, nil
}
