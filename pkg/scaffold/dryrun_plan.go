package scaffold

import "fmt"

type DryRunPlan struct {
	Template       string                `json:"template"`
	Description    string                `json:"description"`
	TargetDir      string                `json:"target_dir"`
	ResolvedInputs []DryRunResolvedInput `json:"resolved_inputs"`
	Actions        []DryRunAction        `json:"actions"`
}

type DryRunResolvedInput struct {
	Name        string `json:"name"`
	TemplateVar string `json:"template_var"`
	Value       string `json:"value"`
}

type DryRunActionKind string

const (
	DryRunActionCreateDir  DryRunActionKind = "create_dir"
	DryRunActionRenderFile DryRunActionKind = "render_file"
	DryRunActionCopyFile   DryRunActionKind = "copy_file"
)

type DryRunAction struct {
	Kind   DryRunActionKind `json:"kind"`
	Source string           `json:"source,omitempty"`
	Target string           `json:"target"`
}

func (c *Creator) BuildDryRunPlan(opts Options) (DryRunPlan, error) {
	if err := c.validateCreateOptions(opts); err != nil {
		return DryRunPlan{}, err
	}
	if err := c.preflightDestDir(opts); err != nil {
		return DryRunPlan{}, err
	}

	manifest, vars, err := c.templateManifestAndVars(opts)
	if err != nil {
		return DryRunPlan{}, err
	}

	plan := DryRunPlan{
		Template:       opts.Lang,
		Description:    manifest.Description,
		TargetDir:      opts.destinationDir(),
		ResolvedInputs: resolveDryRunInputs(manifest.Inputs, vars),
	}

	err = walkTemplateEntries(c.fsys, opts.Lang, plan.TargetDir, vars, func(entry templateEntry) error {
		action := DryRunAction{Target: entry.destPath}
		if entry.isDir {
			action.Kind = DryRunActionCreateDir
			plan.Actions = append(plan.Actions, action)
			return nil
		}

		loaded, err := readTemplateEntry(c.fsys, entry)
		if err != nil {
			return err
		}

		action.Source = entry.srcPath
		if entry.isTemplate {
			action.Kind = DryRunActionRenderFile
			if _, err := renderTemplateEntry(loaded, vars); err != nil {
				return err
			}
		} else {
			action.Kind = DryRunActionCopyFile
		}

		plan.Actions = append(plan.Actions, action)
		return nil
	})
	if err != nil {
		return DryRunPlan{}, err
	}

	return plan, nil
}

func resolveDryRunInputs(inputs []TemplateManifestInput, vars TemplateVars) []DryRunResolvedInput {
	resolved := make([]DryRunResolvedInput, 0, len(inputs))
	for _, input := range inputs {
		resolved = append(resolved, DryRunResolvedInput{
			Name:        input.Name,
			TemplateVar: input.TemplateVar,
			Value:       resolveDryRunInputValue(input.TemplateVar, vars),
		})
	}

	return resolved
}

func resolveDryRunInputValue(templateVar string, vars TemplateVars) string {
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
