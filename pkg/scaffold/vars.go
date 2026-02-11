package scaffold

import (
	"os/user"
	"time"
)

// TemplateVars holds the variables available for template rendering.
type TemplateVars struct {
	ProjectName string
	ModulePath  string
	Author      string
	Year        int
}

// NewTemplateVars creates a TemplateVars with sensible defaults.
func NewTemplateVars(projectName, modulePath string) TemplateVars {
	if modulePath == "" {
		modulePath = projectName
	}

	author := "author"
	if u, err := user.Current(); err == nil && u.Username != "" {
		author = u.Username
	}

	return TemplateVars{
		ProjectName: projectName,
		ModulePath:  modulePath,
		Author:      author,
		Year:        time.Now().Year(),
	}
}
