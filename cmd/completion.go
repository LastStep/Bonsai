package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	// Replace Cobra's auto-generated `completion` command with an
	// explicit one so the help text names the install snippets and
	// the subcommand shows up in `bonsai --help`. The
	// `CompletionOptions.DisableDefaultCmd = true` line in root.go
	// must stay in sync — without it, both this command and the
	// auto-generated one would register, and Cobra rejects duplicate
	// child names at startup.
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for bonsai.

To enable completions for the current shell session, source the script
directly. To make completions persistent, append the script to your
shell's startup file or drop it in the location your shell scans for
completion modules.

Bash:
  $ source <(bonsai completion bash)
  # persist:
  $ bonsai completion bash > /etc/bash_completion.d/bonsai           # Linux
  $ bonsai completion bash > $(brew --prefix)/etc/bash_completion.d/bonsai  # macOS

Zsh:
  $ source <(bonsai completion zsh)
  # persist (with completions enabled, e.g. via 'autoload -U compinit && compinit'):
  $ bonsai completion zsh > "${fpath[1]}/_bonsai"

Fish:
  $ bonsai completion fish | source
  # persist:
  $ bonsai completion fish > ~/.config/fish/completions/bonsai.fish

PowerShell:
  PS> bonsai completion powershell | Out-String | Invoke-Expression
  # persist: append the same line to your $PROFILE.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}
