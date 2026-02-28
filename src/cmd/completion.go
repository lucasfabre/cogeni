package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

### Bash:

` + "```bash" + `
  $ source <(cogeni completion bash)

  # Note: if you are calling the binary via a relative path (e.g., ./cogeni),
  # bash completion might not work unless you also register it for that path:
  # complete -F __start_cogeni ./cogeni

  # To load completions for each session, add to your .bashrc:
  # cogeni completion bash > /etc/bash_completion.d/cogeni
` + "```" + `

### Zsh:

` + "```zsh" + `
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, add to your .zshrc:
  $ cogeni completion zsh > "${fpath[1]}/_cogeni"

  # Note: if you are calling the binary via a relative path (e.g., ./cogeni),
  # you must also register it for that path:
  # compdef _cogeni ./cogeni
` + "```" + `

### Fish:

` + "```fish" + `
  $ cogeni completion fish | source

  # To load completions for each session, add to your config.fish:
  # cogeni completion fish > ~/.config/fish/completions/cogeni.fish
` + "```" + `

### PowerShell:

` + "```powershell" + `
  PS> cogeni completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> cogeni completion powershell > cogeni.ps1
  # and source this file from your PowerShell profile.
` + "```" + `
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			// We generate the standard completion but then tweak it to support
			// both 'cogeni' and the actual path used to invoke it (like './cogeni').
			// This makes it work out of the box with 'source <(./cogeni completion bash)'.
			cmd.Root().GenBashCompletionV2(os.Stdout, true)
			fmt.Printf("\n# Register for the actual path used to invoke cogeni\n")
			fmt.Printf("complete -o default -F __start_cogeni %q 2>/dev/null || true\n", os.Args[0])
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
			fmt.Printf("\n# Register for the actual path used to invoke cogeni\n")
			fmt.Printf("compdef _cogeni %q 2>/dev/null || true\n", os.Args[0])
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
