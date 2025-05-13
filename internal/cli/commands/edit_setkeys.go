package commands

import (
	"fmt"
	"simple-sops/internal/config"
	"simple-sops/internal/encrypt"
	"simple-sops/pkg/logging"

	"github.com/spf13/cobra"
)

// EditCmd returns the edit command
func EditCmd() *cobra.Command {
	var keyFile string

	cmd := &cobra.Command{
		Use:   "edit [file]",
		Short: "Edit an encrypted file",
		Long:  `Edit an encrypted file using SOPS.`,
		Args:  cobra.ExactArgs(1),
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

			// Edit the file
			if err := encrypt.EditFile(args[0], keyFile, appConfig.AlwaysUseOnePassword); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&keyFile, "key-file", "k", "", "Age key file to use (defaults to config setting)")

	return cmd
}

// SetKeysCmd returns the set-keys command
func SetKeysCmd() *cobra.Command {
	var keyFile string

	cmd := &cobra.Command{
		Use:   "set-keys [file]",
		Short: "Choose which keys to encrypt in a file",
		Long:  `Set the encryption rules for a specific file in the SOPS configuration.`,
		Args:  cobra.ExactArgs(1),
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

			// Get predefined patterns
			patterns := encrypt.PredefinedEncryptionPatterns()

			// Create list of choices
			var choices []string
			for name := range patterns {
				choices = append(choices, name)
			}
			choices = append(choices, "Custom pattern")

			// Prompt user for encryption pattern
			choice, err := logging.PromptChoice("What do you want to encrypt in this file?", choices)
			if err != nil {
				return fmt.Errorf("invalid choice: %w", err)
			}

			var encryptedRegex string

			if choice <= len(patterns) {
				// Use predefined pattern
				encryptedRegex = patterns[choices[choice-1]]
			} else {
				// Custom pattern
				logging.Info("Enter your regex pattern to match keys you want to encrypt:")
				logging.Info("Example: ^(password|api_key|secret)")
				encryptedRegex = logging.PromptInput("Pattern")
			}

			// Set encryption keys for the file
			if err := encrypt.SetEncryptionKeys(args[0], keyFile, encryptedRegex, appConfig.AlwaysUseOnePassword); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&keyFile, "key-file", "k", "", "Age key file to use (defaults to config setting)")

	return cmd
}
