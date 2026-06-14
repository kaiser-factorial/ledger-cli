package commands

import (
	"os"

	"github.com/spf13/cobra"
)

func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `To load completions:

Bash:
  $ source <(ledger completion bash)

  # To load completions for each session, add this line to your ~/.bashrc:
  # source <(ledger completion bash)

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. Run the following once:
  
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, add this line to your ~/.zshrc:
  # source <(ledger completion zsh)

Fish:
  $ ledger completion fish | source

  # To load completions for each session, add the following to your ~/.config/fish/config.fish:
  # ledger completion fish | source
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

	return cmd
}