package protocoltoml

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	domain "github.com/JackDrogon/project/internal/scaffold"
	toml "github.com/pelletier/go-toml/v2"
)

const ConfigVersion = 1

var reservedConfigInputKeys = map[string]struct{}{
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

type Config struct {
	Version    int                `toml:"-"`
	New        *ConfigNewSection  `toml:"new"`
	Init       *ConfigInitSection `toml:"init"`
	List       *ConfigListSection `toml:"list"`
	Inspect    *ConfigInspect     `toml:"inspect"`
	VersionCmd *ConfigVersionCmd  `toml:"version"`
	Completion *ConfigCompletion  `toml:"completion"`
}

var (
	configVersionPattern = regexp.MustCompile(`^\s*version\s*=\s*([0-9]+)\s*(?:#.*)?$`)
	configSectionPattern = regexp.MustCompile(`^\s*\[{1,2}[^\]]+\]{1,2}\s*(?:#.*)?$`)
)

type ConfigNewSection struct {
	Lang        *string           `toml:"lang"`
	ProjectName *string           `toml:"project_name"`
	Module      *string           `toml:"module"`
	GitMode     *string           `toml:"git_mode"`
	Signoff     *bool             `toml:"signoff"`
	Inputs      map[string]string `toml:"inputs"`
}

type ConfigInitSection struct {
	Lang      *string           `toml:"lang"`
	TargetDir *string           `toml:"target_dir"`
	Module    *string           `toml:"module"`
	GitMode   *string           `toml:"git_mode"`
	Signoff   *bool             `toml:"signoff"`
	Inputs    map[string]string `toml:"inputs"`
}

type ConfigListSection struct {
	Format        *string  `toml:"format"`
	Compact       *bool    `toml:"compact"`
	Detail        *bool    `toml:"detail"`
	Table         *bool    `toml:"table"`
	Sort          *string  `toml:"sort"`
	MinGovernance *string  `toml:"min_governance"`
	RequiredAsset []string `toml:"required_assets"`
}

type ConfigInspect struct {
	Lang    *string `toml:"lang"`
	Format  *string `toml:"format"`
	Compact *bool   `toml:"compact"`
	Mode    *string `toml:"mode"`
}

type ConfigVersionCmd struct {
	Verbose *bool `toml:"verbose"`
}

type ConfigCompletion struct {
	Shell *string `toml:"shell"`
}

func DecodeConfig(content []byte, path string) (Config, error) {
	if err := rejectLegacyJSON(content, path); err != nil {
		return Config{}, err
	}

	version, strippedContent, err := extractConfigVersion(content, path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	decoder := toml.NewDecoder(bytes.NewReader(strippedContent))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config file %s: %w", path, err)
	}
	cfg.Version = version

	if err := validateConfig(cfg, path); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func extractConfigVersion(content []byte, path string) (int, []byte, error) {
	lines := strings.Split(string(content), "\n")
	version := 0
	seen := false
	inTopLevel := true
	kept := make([]string, 0, len(lines))

	for _, line := range lines {
		if configSectionPattern.MatchString(line) {
			inTopLevel = false
			kept = append(kept, line)
			continue
		}

		if !inTopLevel {
			kept = append(kept, line)
			continue
		}

		matches := configVersionPattern.FindStringSubmatch(line)
		if len(matches) == 0 {
			kept = append(kept, line)
			continue
		}
		if seen {
			return 0, nil, fmt.Errorf("config file %s defines version more than once", path)
		}
		parsed, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, nil, fmt.Errorf("config file %s has invalid version value %q", path, matches[1])
		}
		version = parsed
		seen = true
	}

	if !seen {
		return 0, nil, fmt.Errorf("config file %s must declare version", path)
	}

	return version, []byte(strings.Join(kept, "\n")), nil
}

func validateConfig(cfg Config, path string) error {
	if cfg.Version != ConfigVersion {
		return fmt.Errorf("config file %s has unsupported version %d", path, cfg.Version)
	}

	if cfg.New != nil {
		if err := validateConfigGitMode(path, "new.git_mode", cfg.New.GitMode); err != nil {
			return err
		}
		if err := validateConfigInputs(path, "new.inputs", cfg.New.Inputs); err != nil {
			return err
		}
	}

	if cfg.Init != nil {
		if err := validateConfigGitMode(path, "init.git_mode", cfg.Init.GitMode); err != nil {
			return err
		}
		if err := validateConfigInputs(path, "init.inputs", cfg.Init.Inputs); err != nil {
			return err
		}
	}

	if cfg.List != nil {
		if err := validateConfigOutputFormat(path, "list.format", cfg.List.Format); err != nil {
			return err
		}
		if err := validateConfigListSort(path, cfg.List.Sort); err != nil {
			return err
		}
		if err := validateConfigListGovernance(path, cfg.List.MinGovernance); err != nil {
			return err
		}
	}

	if cfg.Inspect != nil {
		if err := validateConfigOutputFormat(path, "inspect.format", cfg.Inspect.Format); err != nil {
			return err
		}
		if err := validateConfigInspectMode(path, cfg.Inspect.Mode); err != nil {
			return err
		}
	}

	if cfg.Completion != nil {
		if err := validateConfigCompletionShell(path, cfg.Completion.Shell); err != nil {
			return err
		}
	}

	return nil
}

func validateConfigOutputFormat(path, field string, value *string) error {
	if value == nil {
		return nil
	}

	switch strings.TrimSpace(*value) {
	case "text", "toml":
		return nil
	default:
		return fmt.Errorf("config file %s %s %q is invalid: must be one of text, toml", path, field, *value)
	}
}

func validateConfigGitMode(path, field string, value *string) error {
	if value == nil {
		return nil
	}

	switch domain.GitMode(strings.TrimSpace(*value)) {
	case domain.GitModeNone, domain.GitModeInitOnly, domain.GitModeInitCommit:
		return nil
	default:
		return fmt.Errorf(
			"config file %s %s %q is invalid: must be one of %s, %s, %s",
			path,
			field,
			*value,
			domain.GitModeNone,
			domain.GitModeInitOnly,
			domain.GitModeInitCommit,
		)
	}
}

func validateConfigListSort(path string, value *string) error {
	if value == nil {
		return nil
	}

	switch strings.TrimSpace(*value) {
	case "name", "governance", "repo-files":
		return nil
	default:
		return fmt.Errorf("config file %s list.sort %q is invalid: must be one of name, governance, repo-files", path, *value)
	}
}

func validateConfigListGovernance(path string, value *string) error {
	if value == nil {
		return nil
	}

	switch strings.TrimSpace(*value) {
	case "", "minimal", "basic", "standard", "rich":
		return nil
	default:
		return fmt.Errorf("config file %s list.min_governance %q is invalid: must be one of minimal, basic, standard, rich", path, *value)
	}
}

func validateConfigInspectMode(path string, value *string) error {
	if value == nil {
		return nil
	}

	switch strings.TrimSpace(*value) {
	case "all", "render", "copy":
		return nil
	default:
		return fmt.Errorf("config file %s inspect.mode %q is invalid: must be one of all, render, copy", path, *value)
	}
}

func validateConfigCompletionShell(path string, value *string) error {
	if value == nil {
		return nil
	}

	switch strings.TrimSpace(*value) {
	case "bash", "zsh", "fish", "powershell":
		return nil
	default:
		return fmt.Errorf("config file %s completion.shell %q is invalid: must be one of bash, zsh, fish, powershell", path, *value)
	}
}

func validateConfigInputs(path, field string, values map[string]string) error {
	for key := range values {
		if _, reserved := reservedConfigInputKeys[key]; reserved {
			return fmt.Errorf("config file %s %s contains reserved key %q", path, field, key)
		}
	}
	return nil
}
