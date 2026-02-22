package cmd

import (
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for Shopline CLI.

To load completions:

Bash:
  $ source <(spl completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ spl completion bash > /etc/bash_completion.d/spl
  # macOS:
  $ spl completion bash > $(brew --prefix)/etc/bash_completion.d/spl

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ spl completion zsh > "${fpath[1]}/_spl"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ spl completion fish | source
  # To load completions for each session, execute once:
  $ spl completion fish > ~/.config/fish/completions/spl.fish

PowerShell:
  PS> spl completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> spl completion powershell > spl.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(out)
		case "zsh":
			return cmd.Root().GenZshCompletion(out)
		case "fish":
			return cmd.Root().GenFishCompletion(out, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(out)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
