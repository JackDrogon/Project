package create

import (
	"io"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appconfig "github.com/JackDrogon/project/internal/app/config"
)

type Command string

const (
	CommandNew  Command = "new"
	CommandInit Command = "init"
)

type ValueOrigin string

const (
	ValueOriginDefault ValueOrigin = "default"
	ValueOriginFlag    ValueOrigin = "flag"
	ValueOriginArg     ValueOrigin = "arg"
	ValueOriginSet     ValueOrigin = "set"
	ValueOriginReplay  ValueOrigin = "replay"
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
	ExplainConfig   bool
	Stderr          io.Writer
	ActiveConfig    appconfig.ActiveConfig
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
	Command             Command
	ActiveConfig        appconfig.ActiveConfig
	Replay              protocoltoml.Replay
	HasReplay           bool
	ExplicitSetValues   map[string]string
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
	Origins ResolutionOrigins
}

type ResolutionOrigins struct {
	Lang           ValueOrigin
	ProjectName    ValueOrigin
	TargetDir      ValueOrigin
	Module         ValueOrigin
	GitMode        ValueOrigin
	Signoff        ValueOrigin
	TemplateInputs map[string]ValueOrigin
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
