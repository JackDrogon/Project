package scaffold

type GitMode string

const (
	GitModeNone       GitMode = "none"
	GitModeInitOnly   GitMode = "init-only"
	GitModeInitCommit GitMode = "init+commit"
)

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

func (req CreateRequest) DestinationDir() string {
	if req.TargetDir != "" {
		return req.TargetDir
	}

	return req.ProjectName
}

type CreateResult struct {
	TargetDir string
	GitMode   GitMode
}

type TemplateVars struct {
	ProjectName      string
	ProjectNameLower string
	ModulePath       string
	GoVersion        string
	Author           string
	Year             int
}
