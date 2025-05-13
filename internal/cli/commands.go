package cli

import (
	"simple-sops/internal/cli/commands"

	"github.com/spf13/cobra"
)

// RegisterCommands adds all the CLI commands to the root command
func RegisterCommands(rootCmd *cobra.Command) {
	// Core commands
	rootCmd.AddCommand(commands.EncryptCmd())
	rootCmd.AddCommand(commands.DecryptCmd())
	rootCmd.AddCommand(commands.EditCmd())
	rootCmd.AddCommand(commands.SetKeysCmd())
	rootCmd.AddCommand(commands.ConfigCmd())
	rootCmd.AddCommand(commands.RemoveCmd())
	rootCmd.AddCommand(commands.CleanConfigCmd())
	rootCmd.AddCommand(commands.GetKeyCmd())
	rootCmd.AddCommand(commands.ClearKeyCmd())

	// New commands
	rootCmd.AddCommand(commands.GenerateKeyCmd())
	rootCmd.AddCommand(commands.RunCmd())
	rootCmd.AddCommand(commands.CompletionCmd())
}
