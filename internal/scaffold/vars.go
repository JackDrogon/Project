package scaffold

import (
	"fmt"
	"strconv"
	"strings"
)

func NewTemplateVars(projectName, modulePath, goVersion, author string, year int) TemplateVars {
	if modulePath == "" {
		modulePath = projectName
	}

	return TemplateVars{
		ProjectName:      projectName,
		ProjectNameLower: strings.ToLower(projectName),
		ModulePath:       modulePath,
		GoVersion:        goVersion,
		Author:           author,
		Year:             year,
	}
}

type ManifestInput struct {
	Name        string
	TemplateVar string
}

func ResolveTemplateVars(inputs []ManifestInput, req CreateRequest, base TemplateVars) (TemplateVars, error) {
	vars := base
	if len(req.TemplateInputValues) == 0 {
		return vars, nil
	}

	declaredInputs := make(map[string]ManifestInput, len(inputs))
	for _, input := range inputs {
		declaredInputs[input.Name] = input
	}

	for name, value := range req.TemplateInputValues {
		if name == "module_path" {
			return TemplateVars{}, fmt.Errorf("template input %q must be provided via module path options", name)
		}

		input, ok := declaredInputs[name]
		if !ok {
			return TemplateVars{}, fmt.Errorf("template input %q is not declared by template", name)
		}

		if err := ApplyTemplateInputValue(&vars, input, value); err != nil {
			return TemplateVars{}, err
		}
	}

	return vars, nil
}

func ApplyTemplateInputValue(vars *TemplateVars, input ManifestInput, value string) error {
	switch input.TemplateVar {
	case "ModulePath":
		return fmt.Errorf("template input %q must be provided via module path options", input.Name)
	case "GoVersion":
		vars.GoVersion = value
	case "Author":
		vars.Author = value
	case "Year":
		year, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("template input %q must be a valid year: %w", input.Name, err)
		}
		vars.Year = year
	default:
		return fmt.Errorf("template input %q has unsupported template_var %q", input.Name, input.TemplateVar)
	}

	return nil
}
