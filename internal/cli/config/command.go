package config

import (
	"fmt"
	"strconv"
	"strings"

	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/spf13/cobra"
)

type Dependencies struct {
	NewService func() *appconfig.Service
	// Config is filled in by the root command before any subcommand runs.
	// A nil Config means "no config file", which is what tests that do not
	// exercise config resolution want.
	Config *appconfig.Resolved
}

func (d Dependencies) newService() *appconfig.Service {
	if d.NewService == nil {
		panic("config dependencies require NewService")
	}

	return d.NewService()
}

func (d Dependencies) resolved() appconfig.Resolved {
	if d.Config == nil {
		return appconfig.Resolved{}
	}

	return *d.Config
}

func NewCommand(deps Dependencies) *cobra.Command {
	service := deps.newService()

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show active config summary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := deps.resolved()
			resolvedPath, err := service.ResolvePath(config.Options)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), renderConfigSummary(config.Active, resolvedPath, config.LoadErr))
			return err
		},
	}

	cmd.AddCommand(newPathCommand(service, deps), newInitCommand(service, deps), newValidateCommand(service, deps))
	return cmd
}

func newPathCommand(service *appconfig.Service, deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show resolved config path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := service.ResolvePath(deps.resolved().Options)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
			return err
		},
	}
}

func newInitCommand(service *appconfig.Service, deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a seed config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := service.InitConfig(deps.resolved().Options)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created config file: %s\n", path)
			return err
		},
	}
}

func newValidateCommand(service *appconfig.Service, deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the resolved config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := deps.resolved()
			path, err := service.ResolvePath(config.Options)
			if err != nil {
				return err
			}

			if config.LoadErr != nil {
				return fmt.Errorf("config validation failed for %q: %w", path, config.LoadErr)
			}

			if config.Active.Config == nil {
				return fmt.Errorf("config file %q does not exist", path)
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Config is valid: %s\n", path)
			return err
		},
	}
}

func renderConfigSummary(active appconfig.ActiveConfig, resolvedPath string, loadErr error) string {
	var b strings.Builder

	b.WriteString("active config summary:\n")
	_, _ = fmt.Fprintf(&b, "  source: %s\n", configSource(active))
	_, _ = fmt.Fprintf(&b, "  path: %s\n", configPath(active, resolvedPath))
	_, _ = fmt.Fprintf(&b, "  loaded: %t\n", active.Config != nil)
	_, _ = fmt.Fprintf(&b, "  version: %s\n", configVersion(active))
	_, _ = fmt.Fprintf(&b, "  sections: %s\n", strings.Join(configSections(active), ", "))
	if loadErr != nil {
		_, _ = fmt.Fprintf(&b, "  load_error: %v\n", loadErr)
	}

	if active.Config == nil {
		b.WriteString("  hint: use --config <path> or create config.toml in the default user config directory\n")
	}

	return b.String()
}

func configSource(active appconfig.ActiveConfig) string {
	if strings.TrimSpace(string(active.Source)) == "" {
		return string(appconfig.SourceNone)
	}
	return string(active.Source)
}

func configPath(active appconfig.ActiveConfig, resolvedPath string) string {
	if strings.TrimSpace(active.Path) != "" {
		return active.Path
	}
	if strings.TrimSpace(resolvedPath) == "" {
		return "(none)"
	}
	return resolvedPath
}

func configVersion(active appconfig.ActiveConfig) string {
	if active.Config == nil {
		return "(none)"
	}
	return strconv.Itoa(active.Config.Version)
}

func configSections(active appconfig.ActiveConfig) []string {
	if active.Config == nil {
		return []string{"(none)"}
	}

	sections := make([]string, 0, 6)
	if active.Config.New != nil {
		sections = append(sections, "new")
	}
	if active.Config.Init != nil {
		sections = append(sections, "init")
	}
	if active.Config.List != nil {
		sections = append(sections, "list")
	}
	if active.Config.Inspect != nil {
		sections = append(sections, "inspect")
	}
	if active.Config.VersionCmd != nil {
		sections = append(sections, "version")
	}
	if active.Config.Completion != nil {
		sections = append(sections, "completion")
	}

	if len(sections) == 0 {
		return []string{"(none)"}
	}

	return sections
}
