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
	MkdirAll      func(path string, perm os.FileMode) error
	Stat          func(name string) (os.FileInfo, error)
	WriteFile     func(name string, data []byte, perm os.FileMode) error
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
		MkdirAll:      os.MkdirAll,
		Stat:          os.Stat,
		WriteFile:     os.WriteFile,
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
	if d.MkdirAll == nil {
		d.MkdirAll = defaults.MkdirAll
	}
	if d.Stat == nil {
		d.Stat = defaults.Stat
	}
	if d.WriteFile == nil {
		d.WriteFile = defaults.WriteFile
	}
	if d.ReadFile == nil {
		d.ReadFile = defaults.ReadFile
	}
	if d.Decode == nil {
		d.Decode = defaults.Decode
	}
	return d
}

func (s *Service) ResolvePath(opts LoadOptions) (string, error) {
	if opts.ExplicitPath != "" {
		return opts.ExplicitPath, nil
	}

	return s.defaultConfigPath()
}

func (s *Service) InitConfig(opts LoadOptions) (string, error) {
	path, err := s.ResolvePath(opts)
	if err != nil {
		return "", err
	}

	if err := s.deps.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create config directory for %q: %w", path, err)
	}

	if _, err := s.deps.Stat(path); err == nil {
		return "", fmt.Errorf("config file %q already exists", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat config file %q: %w", path, err)
	}

	content := fmt.Appendf(nil, "version = %d\n", protocoltoml.ConfigVersion)
	if err := s.deps.WriteFile(path, content, 0o644); err != nil {
		return "", fmt.Errorf("write config file %q: %w", path, err)
	}

	return path, nil
}

func (s *Service) LoadActiveConfig(opts LoadOptions) (ActiveConfig, error) {
	if opts.ExplicitPath != "" {
		return s.loadFromPath(SourceExplicit, opts.ExplicitPath)
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
