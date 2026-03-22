package scaffold

type DryRunPlan struct {
	Template       string
	Description    string
	TargetDir      string
	ResolvedInputs []DryRunResolvedInput
	Actions        []DryRunAction
}

type DryRunResolvedInput struct {
	Name        string
	TemplateVar string
	Value       string
}

type DryRunActionKind string

const (
	DryRunActionCreateDir  DryRunActionKind = "create_dir"
	DryRunActionRenderFile DryRunActionKind = "render_file"
	DryRunActionCopyFile   DryRunActionKind = "copy_file"
)

type DryRunAction struct {
	Kind   DryRunActionKind
	Source string
	Target string
}
