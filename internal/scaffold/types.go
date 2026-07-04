package scaffold

// GitMode specifies the Git initialization behavior for a new project.
// It controls whether Git is initialized, and if so, whether an initial commit is created.
type GitMode string

const (
	// GitModeNone skips Git initialization entirely.
	GitModeNone GitMode = "none"
	// GitModeInitOnly initializes a Git repository without creating an initial commit.
	GitModeInitOnly GitMode = "init-only"
	// GitModeInitCommit initializes a Git repository and creates an initial commit.
	GitModeInitCommit GitMode = "init+commit"
)

// CreateRequest represents a project scaffolding request with all necessary parameters.
// It encapsulates the template selection, destination path, Git initialization mode,
// and template variable bindings required to generate a new project.
type CreateRequest struct {
	Lang                  string
	ProjectName           string
	TargetDir             string
	ModulePath            string
	TemplateInputValues   map[string]string
	Force                 bool
	AllowExistingEmptyDir bool
	Signoff               bool
	DryRun                bool
	NoGit                 bool
	GitMode               GitMode
}

// DestinationDir returns the effective destination directory for the project.
// It returns TargetDir if explicitly set, otherwise falls back to ProjectName.
func (req CreateRequest) DestinationDir() string {
	if req.TargetDir != "" {
		return req.TargetDir
	}

	return req.ProjectName
}

// TemplateVars holds the variables available for template rendering.
// These variables are substituted into .tmpl files during project scaffolding.
type TemplateVars struct {
	ProjectName      string
	ProjectNameLower string
	ModulePath       string
	GoVersion        string
	Author           string
	Year             int
}
