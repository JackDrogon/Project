package completion

import (
	appcompletion "github.com/JackDrogon/project/internal/app/completion"
	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/spf13/cobra"
)

type Dependencies struct {
	NewService func() *appcompletion.Service
	// Config is filled in by the root command before any subcommand runs.
	// A nil Config means "no config file", which is what tests that do not
	// exercise config resolution want.
	Config *appconfig.Resolved
}

func (d Dependencies) newService() *appcompletion.Service {
	if d.NewService == nil {
		panic("completion dependencies require NewService")
	}

	return d.NewService()
}

func (d Dependencies) activeConfig() appconfig.ActiveConfig {
	if d.Config == nil {
		return appconfig.ActiveConfig{}
	}

	return d.Config.Active
}

func NewCommand(deps Dependencies) *cobra.Command {
	service := deps.newService()

	return &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate shell completion script",
		Long:                  longDescription,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := ""
			if len(args) > 0 {
				arg = args[0]
			}
			shell := service.ResolveShell(arg, len(args) > 0, deps.activeConfig())
			if shell == "" {
				return cobra.ExactArgs(1)(cmd, args)
			}
			if err := cobra.OnlyValidArgs(cmd, []string{shell}); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletionV2(out, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(out)
			case "fish":
				return cmd.Root().GenFishCompletion(out, true)
			}
			return cmd.Root().GenPowerShellCompletionWithDesc(out)
		},
	}
}

const longDescription = `Generate shell completion scripts for project.

To load completions:

Bash:
  $ source <(project completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ project completion bash > /etc/bash_completion.d/project
  # macOS:
  $ project completion bash > $(brew --prefix)/etc/bash_completion.d/project

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ project completion zsh > "${fpath[1]}/_project"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ project completion fish | source

  # To load completions for each session, execute once:
  $ project completion fish > ~/.config/fish/completions/project.fish

PowerShell:
  PS> project completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> project completion powershell > project.ps1
  # and source this file from your PowerShell profile.
`
