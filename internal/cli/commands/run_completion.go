package commands

import (
	"fmt"
	"os"
	"simple-sops/internal/config"
	"simple-sops/internal/run"

	"github.com/spf13/cobra"
)

// RunCmd returns the run command
func RunCmd() *cobra.Command {
	var keyFile string

	cmd := &cobra.Command{
		Use:   "run [encrypted-file] [output-file (optional)] [command...]",
		Short: "Run a command with a decrypted file",
		Long:  `Decrypt a file, run a command with the decrypted content, and clean up afterward.`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load application config
			appConfig, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// If key file not specified in flags, use the one from config
			if keyFile == "" {
				keyFile = appConfig.KeyFile
			}

			// Parse run command arguments
			encryptedFile, outputFile, command, commandArgs, err := run.ParseRunCommand(args)
			if err != nil {
				return err
			}

			// Run the command with the decrypted file - pass the new parameter
			if err := run.RunWithEncryptedFile(encryptedFile, outputFile, command, commandArgs, keyFile, appConfig.AlwaysUseOnePassword); err != nil {
				return err
			}

			return nil
		},
		Example: `  simple-sops run config.enc.yaml "kubectl apply -f config.enc.yaml"
  simple-sops run secret.enc.yaml plain.yaml "cat plain.yaml"
  simple-sops run ~/.env.enc cat`,
	}

	cmd.Flags().StringVarP(&keyFile, "key-file", "k", "", "Age key file to use (defaults to config setting)")

	return cmd
}

// CompletionCmd returns the completion command for generating shell completions
func CompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for simple-sops.
To load completions:

Bash:
  $ source <(simple-sops completion bash)

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ simple-sops completion zsh > "${fpath[1]}/_simple-sops"

Fish:
  $ simple-sops completion fish > ~/.config/fish/completions/simple-sops.fish

PowerShell:
  PS> simple-sops completion powershell | Out-String | Invoke-Expression
`,
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			switch args[0] {
			case "bash":
				err = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				err = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				err = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				err = cmd.Root().GenPowerShellCompletion(os.Stdout)
			}

			if err != nil {
				return fmt.Errorf("failed to generate completion script: %w", err)
			}

			return nil
		},
	}

	return cmd
}
