package main

import (
	"github.com/spf13/cobra"
)

func buildCompletionCommand(commandDependencies) *cobra.Command {
	return &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate shell completion script",
		Long:                  completionLong,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			switch args[0] {
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

func init() {
	registerOrderedCommand(commandKeyCompletion, commandOrderCompletion, buildCompletionCommand)
}

const completionLong = `Generate shell completion scripts for project.

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
