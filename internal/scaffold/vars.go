package scaffold

import (
	"fmt"
	"strconv"
	"strings"
)

// Template variable names, as they appear in a manifest's template_var field.
// These are the contract between a template's manifest and TemplateVars, so
// they belong to the domain rather than to whichever caller spells them out.
const (
	TemplateVarProjectName = "ProjectName"
	TemplateVarModulePath  = "ModulePath"
	TemplateVarGoVersion   = "GoVersion"
	TemplateVarAuthor      = "Author"
	TemplateVarYear        = "Year"

	// templateVarProjectNameLower is derived from ProjectName rather than
	// supplied by a manifest, so it is not part of the input vocabulary.
	templateVarProjectNameLower = "ProjectNameLower"
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

// ManifestInput is one input a template declares. Required means the bound
// template variable must hold a value by the time rendering starts, no matter
// whether it came from --set, a replay, the config file, or a computed default.
type ManifestInput struct {
	Name        string
	TemplateVar string
	Required    bool
}

func ResolveTemplateVars(inputs []ManifestInput, req CreateRequest, base TemplateVars) (TemplateVars, error) {
	vars := base

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

		if err := applyTemplateInputValue(&vars, input, value); err != nil {
			return TemplateVars{}, err
		}
	}

	if err := validateRequiredInputs(inputs, vars); err != nil {
		return TemplateVars{}, err
	}

	return vars, nil
}

// validateRequiredInputs walks the declared inputs in manifest order so the
// reported failure never depends on map iteration order.
func validateRequiredInputs(inputs []ManifestInput, vars TemplateVars) error {
	for _, input := range inputs {
		if !input.Required {
			continue
		}

		if _, ok := TemplateVarValue(vars, input.TemplateVar); !ok {
			return fmt.Errorf("template input %q has unsupported template_var %q", input.Name, input.TemplateVar)
		}
		if !templateVarIsSet(vars, input.TemplateVar) {
			return fmt.Errorf("template input %q is required by the template but resolved to an empty value", input.Name)
		}
	}

	return nil
}

// TemplateVarValue renders one template variable as the string a template
// would see. The second result reports whether the name is a known variable.
func TemplateVarValue(vars TemplateVars, templateVar string) (string, bool) {
	switch templateVar {
	case TemplateVarProjectName:
		return vars.ProjectName, true
	case templateVarProjectNameLower:
		return vars.ProjectNameLower, true
	case TemplateVarModulePath:
		return vars.ModulePath, true
	case TemplateVarGoVersion:
		return vars.GoVersion, true
	case TemplateVarAuthor:
		return vars.Author, true
	case TemplateVarYear:
		return strconv.Itoa(vars.Year), true
	default:
		return "", false
	}
}

// templateVarIsSet reports whether the variable carries a usable value. Year is
// numeric, so its zero value means "unset" rather than the literal "0".
func templateVarIsSet(vars TemplateVars, templateVar string) bool {
	if templateVar == TemplateVarYear {
		return vars.Year != 0
	}

	value, ok := TemplateVarValue(vars, templateVar)
	return ok && strings.TrimSpace(value) != ""
}

func applyTemplateInputValue(vars *TemplateVars, input ManifestInput, value string) error {
	switch input.TemplateVar {
	case TemplateVarModulePath:
		return fmt.Errorf("template input %q must be provided via module path options", input.Name)
	case TemplateVarGoVersion:
		vars.GoVersion = value
	case TemplateVarAuthor:
		vars.Author = value
	case TemplateVarYear:
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
