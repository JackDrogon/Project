package create

import (
	"fmt"
	"path/filepath"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

// resolveNewProjectArgs turns the positional argument of `new` into a project
// name, a target directory, and a module path.
//
// For Go projects with no explicit --module, an argument that is not a valid
// project name but is a valid module path is treated as one: `new -l go
// github.com/acme/tool` scaffolds ./tool with module github.com/acme/tool.
func resolveNewProjectArgs(lang, module, arg string) (projectName, targetDir, modulePath string, err error) {
	if lang == langGo && module == "" {
		if projectErr := domain.ValidateProjectName(arg); projectErr != nil {
			if moduleErr := domain.ValidateModulePath(arg); moduleErr == nil {
				name := domain.ProjectNameFromGoModulePath(arg)
				if err := domain.ValidateProjectName(name); err != nil {
					return "", "", "", err
				}
				return name, name, arg, nil
			}
		}
	}

	return arg, arg, module, nil
}

// projectNameFromTargetDir derives the project name for `init` from the target
// directory, which for the default "." means the name of the current directory.
func projectNameFromTargetDir(targetDir string) (string, error) {
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target directory %q: %w", targetDir, err)
	}
	return filepath.Base(absTarget), nil
}
