package create

import "github.com/JackDrogon/project/internal/adapters/protocoltoml"

type Command string

const (
	CommandNew  Command = "new"
	CommandInit Command = "init"
)

type Flags struct {
	Lang            string
	Module          string
	Signoff         bool
	DryRun          bool
	NoGit           bool
	GitMode         string
	ReplayPath      string
	WriteReplayPath string
	SetValues       []string
}

type Changed struct {
	Lang    bool
	Module  bool
	Signoff bool
	NoGit   bool
	Git     bool
	Force   bool
}

type Runtime struct {
	Replay              protocoltoml.Replay
	HasReplay           bool
	TemplateInputValues map[string]string
}

type NewRequest struct {
	Flags   Flags
	Changed Changed
	Force   bool
	Arg     string
	HasArg  bool
}

type InitRequest struct {
	Flags   Flags
	Changed Changed
	Arg     string
	HasArg  bool
}

type ScaffoldSpec struct {
	Command Command
	Flags   Flags
	Options Options
}

type resolvedScaffoldSettings struct {
	Lang                string
	ModulePath          string
	Signoff             bool
	NoGit               bool
	GitMode             string
	TemplateInputValues map[string]string
}

type targetResolution struct {
	ProjectName           string
	TargetDir             string
	ModulePath            string
	Force                 bool
	AllowExistingEmptyDir bool
}

type Service struct {
	settingsResolver   ScaffoldSettingsResolver
	newTargetResolver  NewTargetResolver
	initTargetResolver InitTargetResolver
}
