package config

import (
	"fmt"
	"strings"

	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/spf13/cobra"
)

type Dependencies struct {
	NewService func() *appconfig.Service
}

func (d Dependencies) newService() *appconfig.Service {
	if d.NewService == nil {
		panic("config dependencies require NewService")
	}

	return d.NewService()
}

func NewCommand(deps Dependencies) *cobra.Command {
	service := deps.newService()

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show active config summary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			active, _ := appconfig.ActiveConfigFromContext(cmd.Context())
			loadCtx, _ := appconfig.LoadContextFromContext(cmd.Context())
			resolvedPath, err := service.ResolvePath(loadCtx)
			if err != nil {
				return err
			}
			loadErr := appconfig.LoadErrorFromContext(cmd.Context())
			_, err = fmt.Fprint(cmd.OutOrStdout(), renderConfigSummary(active, resolvedPath, loadErr))
			return err
		},
	}

	cmd.AddCommand(newPathCommand(service), newInitCommand(service), newValidateCommand(service))
	return cmd
}

func newPathCommand(service *appconfig.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show resolved config path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loadCtx, _ := appconfig.LoadContextFromContext(cmd.Context())
			path, err := service.ResolvePath(loadCtx)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
			return err
		},
	}
}

func newInitCommand(service *appconfig.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a seed config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loadCtx, _ := appconfig.LoadContextFromContext(cmd.Context())
			path, err := service.InitConfig(loadCtx)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created config file: %s\n", path)
			return err
		},
	}
}

func newValidateCommand(service *appconfig.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the resolved config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loadCtx, _ := appconfig.LoadContextFromContext(cmd.Context())
			path, err := service.ResolvePath(loadCtx)
			if err != nil {
				return err
			}

			loadErr := appconfig.LoadErrorFromContext(cmd.Context())
			if loadErr != nil {
				return fmt.Errorf("config validation failed for %q: %w", path, loadErr)
			}

			active, _ := appconfig.ActiveConfigFromContext(cmd.Context())
			if active.Config == nil {
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
	return fmt.Sprintf("%d", active.Config.Version)
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
