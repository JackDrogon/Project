package create

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

func buildScaffoldExplainReport(ctx context.Context, creator *Creator, spec ScaffoldSpec) (string, error) {
	manifest, vars, err := creator.templateManifestAndVars(ctx, spec.Options)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("config source report:\n")
	_, _ = fmt.Fprintf(&b, "  command: %s\n", spec.Command)
	_, _ = fmt.Fprintf(&b, "  active_config_source: %s\n", explainActiveConfigSource(spec))
	_, _ = fmt.Fprintf(&b, "  active_config_path: %s\n", explainActiveConfigPath(spec))
	b.WriteString("  resolved values:\n")
	writeExplainField(&b, "lang", spec.Options.Lang, spec.Origins.Lang)
	writeExplainField(&b, "project_name", spec.Options.ProjectName, spec.Origins.ProjectName)
	writeExplainField(&b, "target_dir", spec.Options.DestinationDir(), spec.Origins.TargetDir)
	writeExplainField(&b, "module", spec.Options.ModulePath, spec.Origins.Module)
	writeExplainField(&b, "git_mode", string(spec.Options.GitMode), spec.Origins.GitMode)
	writeExplainField(&b, "signoff", strconv.FormatBool(spec.Options.Signoff), spec.Origins.Signoff)
	b.WriteString("  template inputs:\n")

	inputs := resolveDryRunInputs(manifest.DomainInputs(), vars)
	if len(inputs) == 0 {
		b.WriteString("    (none)\n")
		return b.String(), nil
	}
	for _, input := range inputs {
		writeExplainField(&b, input.Name, input.Value, explainInputOrigin(spec.Origins, input))
	}

	return b.String(), nil
}

func explainActiveConfigSource(spec ScaffoldSpec) string {
	source := strings.TrimSpace(string(spec.Flags.ActiveConfig.Source))
	if source == "" {
		return "none"
	}
	return source
}

func explainActiveConfigPath(spec ScaffoldSpec) string {
	path := strings.TrimSpace(spec.Flags.ActiveConfig.Path)
	if path == "" {
		return "(none)"
	}
	return path
}

func writeExplainField(b *strings.Builder, name, value string, origin ValueOrigin) {
	_, _ = fmt.Fprintf(b, "    %s: %s (source: %s)\n", name, value, normalizeOrigin(origin))
}

func explainInputOrigin(origins ResolutionOrigins, input domain.DryRunResolvedInput) ValueOrigin {
	if input.TemplateVar == domain.TemplateVarModulePath {
		return normalizeOrigin(origins.Module)
	}
	if origin, ok := origins.TemplateInputs[input.Name]; ok {
		return normalizeOrigin(origin)
	}
	return ValueOriginDefault
}

func normalizeOrigin(origin ValueOrigin) ValueOrigin {
	if origin == "" {
		return ValueOriginDefault
	}
	return origin
}
