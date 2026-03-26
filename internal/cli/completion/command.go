package completion

import (
	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate shell completion script",
		Long:                  longDescription,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := resolveCompletionShell(cmd, args)
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

func resolveCompletionShell(cmd *cobra.Command, args []string) string {
	if len(args) > 0 {
		return args[0]
	}

	active, ok := appconfig.ActiveConfigFromContext(cmd.Context())
	if !ok || active.Config == nil || active.Config.Completion == nil || active.Config.Completion.Shell == nil {
		return ""
	}

	return *active.Config.Completion.Shell
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
