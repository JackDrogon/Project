package config

import (
	"fmt"
	"strings"

	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show active config summary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			active, _ := appconfig.ActiveConfigFromContext(cmd.Context())
			_, err := fmt.Fprint(cmd.OutOrStdout(), renderConfigSummary(active))
			return err
		},
	}

	return cmd
}

func renderConfigSummary(active appconfig.ActiveConfig) string {
	var b strings.Builder

	b.WriteString("active config summary:\n")
	_, _ = fmt.Fprintf(&b, "  source: %s\n", configSource(active))
	_, _ = fmt.Fprintf(&b, "  path: %s\n", configPath(active))
	_, _ = fmt.Fprintf(&b, "  loaded: %t\n", active.Config != nil)
	_, _ = fmt.Fprintf(&b, "  version: %s\n", configVersion(active))
	_, _ = fmt.Fprintf(&b, "  sections: %s\n", strings.Join(configSections(active), ", "))

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

func configPath(active appconfig.ActiveConfig) string {
	if strings.TrimSpace(active.Path) == "" {
		return "(none)"
	}
	return active.Path
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
