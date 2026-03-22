package create

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/JackDrogon/project/internal/adapters/templatefs"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

func (c *Creator) BuildDryRunPlan(opts Options) (domain.DryRunPlan, error) {
	if err := c.validateCreateOptions(opts); err != nil {
		return domain.DryRunPlan{}, err
	}
	if err := c.preflightDestDir(opts); err != nil {
		return domain.DryRunPlan{}, err
	}

	manifest, vars, err := c.templateManifestAndVars(opts)
	if err != nil {
		return domain.DryRunPlan{}, err
	}

	plan := domain.DryRunPlan{
		Template:       opts.Lang,
		Description:    manifest.Description,
		TargetDir:      opts.DestinationDir(),
		ResolvedInputs: resolveDryRunInputs(manifestInputsToDomain(manifest.Inputs), vars),
	}

	err = templatefs.WalkEntries(c.fsys, opts.Lang, plan.TargetDir, vars, func(entry templatefs.Entry) error {
		action := domain.DryRunAction{Target: entry.Destination}
		if entry.IsDir {
			action.Kind = domain.DryRunActionCreateDir
			plan.Actions = append(plan.Actions, action)
			return nil
		}

		loaded, err := templatefs.ReadEntry(c.fsys, entry)
		if err != nil {
			return err
		}

		action.Source = entry.SourcePath
		if entry.IsTemplate {
			action.Kind = domain.DryRunActionRenderFile
			if _, err := templatefs.RenderEntry(loaded, vars); err != nil {
				return err
			}
		} else {
			action.Kind = domain.DryRunActionCopyFile
		}

		plan.Actions = append(plan.Actions, action)
		return nil
	})
	if err != nil {
		return domain.DryRunPlan{}, err
	}

	return plan, nil
}

func resolveDryRunInputs(inputs []domain.ManifestInput, vars domain.TemplateVars) []domain.DryRunResolvedInput {
	resolved := make([]domain.DryRunResolvedInput, 0, len(inputs))
	for _, input := range inputs {
		resolved = append(resolved, domain.DryRunResolvedInput{
			Name:        input.Name,
			TemplateVar: input.TemplateVar,
			Value:       resolveDryRunInputValue(input.TemplateVar, vars),
		})
	}

	return resolved
}

func resolveDryRunInputValue(templateVar string, vars domain.TemplateVars) string {
	switch templateVar {
	case "ModulePath":
		return vars.ModulePath
	case "GoVersion":
		return vars.GoVersion
	case "Author":
		return vars.Author
	case "Year":
		return fmt.Sprintf("%d", vars.Year)
	default:
		return ""
	}
}

func writeDryRunPlan(w io.Writer, plan domain.DryRunPlan, opts Options) error {
	mode, err := domain.ResolveGitMode(opts)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "template: %s\n", plan.Template)
	_, _ = fmt.Fprintf(w, "description: %s\n", plan.Description)
	_, _ = fmt.Fprintf(w, "target_dir: %s\n", plan.TargetDir)
	_, _ = fmt.Fprintln(w, "resolved inputs:")
	_, _ = fmt.Fprintf(w, "  project_name: %s\n", opts.ProjectName)

	modulePath, includeModulePath := dryRunPlanModulePath(plan, opts)
	if includeModulePath {
		_, _ = fmt.Fprintf(w, "  module_path: %s\n", modulePath)
	}

	for _, input := range plan.ResolvedInputs {
		if includeModulePath && input.TemplateVar == "ModulePath" {
			continue
		}
		_, _ = fmt.Fprintf(w, "  %s: %s\n", input.Name, input.Value)
	}
	_, _ = fmt.Fprintf(w, "  git_mode: %s\n", mode)

	_, _ = fmt.Fprintln(w, "explicit overrides:")
	overrides := dryRunPlanOverrides(plan, opts, mode, modulePath, includeModulePath)
	if len(overrides) == 0 {
		_, _ = fmt.Fprintln(w, "  (none)")
	} else {
		for _, override := range overrides {
			_, _ = fmt.Fprintf(w, "  %s: %s\n", override.name, override.value)
		}
	}

	_, _ = fmt.Fprintln(w, "actions:")
	for _, action := range plan.Actions {
		_, _ = fmt.Fprintf(w, "  %s\n", formatDryRunAction(action))
	}

	return nil
}

type dryRunPlanOverride struct {
	name  string
	value string
}

func dryRunPlanModulePath(plan domain.DryRunPlan, opts Options) (string, bool) {
	if value, ok := dryRunPlanResolvedInputValue(plan, "ModulePath"); ok {
		return value, true
	}
	if opts.Lang == "go" {
		return defaultModulePath(opts), true
	}

	return "", false
}

func dryRunPlanResolvedInputValue(plan domain.DryRunPlan, templateVar string) (string, bool) {
	for _, input := range plan.ResolvedInputs {
		if input.TemplateVar == templateVar {
			return input.Value, true
		}
	}

	return "", false
}

func dryRunPlanOverrides(plan domain.DryRunPlan, opts Options, mode domain.GitMode, modulePath string, includeModulePath bool) []dryRunPlanOverride {
	var overrides []dryRunPlanOverride
	if includeModulePath && modulePath != opts.ProjectName {
		overrides = append(overrides, dryRunPlanOverride{name: "module_path", value: modulePath})
	}
	for _, input := range plan.ResolvedInputs {
		if input.TemplateVar == "ModulePath" {
			continue
		}

		if value, ok := opts.TemplateInputValues[input.Name]; ok {
			overrides = append(overrides, dryRunPlanOverride{name: input.Name, value: value})
		}
	}
	if opts.NoGit || opts.GitMode != "" {
		overrides = append(overrides, dryRunPlanOverride{name: "git_mode", value: string(mode)})
	}

	return overrides
}

func formatDryRunAction(action domain.DryRunAction) string {
	switch action.Kind {
	case domain.DryRunActionCreateDir:
		return fmt.Sprintf("create %s%s", action.Target, string(filepath.Separator))
	case domain.DryRunActionRenderFile:
		return fmt.Sprintf("render %s -> %s", action.Source, action.Target)
	case domain.DryRunActionCopyFile:
		return fmt.Sprintf("copy %s -> %s", action.Source, action.Target)
	default:
		return fmt.Sprintf("create %s", action.Target)
	}
}
