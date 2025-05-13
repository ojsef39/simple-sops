package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"simple-sops/internal/config"
	"simple-sops/internal/keymgmt"
	"simple-sops/pkg/logging"
)

// GetKeyCmd returns the get-key command
func GetKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-key",
		Short: "Load SOPS Age key from 1Password",
		Long:  `Retrieve the SOPS Age key from 1Password and store it in a temporary file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the key from 1Password
			tempKeyFile, err := keymgmt.GetKeyFromOnePassword(keymgmt.DefaultOnePasswordConfig)
			if err != nil {
				return fmt.Errorf("failed to get key from 1Password: %w", err)
			}

			// Set the environment variable
			os.Setenv("SOPS_AGE_KEY_FILE", tempKeyFile)

			logging.Success("SOPS Age key loaded from 1Password")
			logging.Info("SOPS_AGE_KEY_FILE set to %s", tempKeyFile)
			logging.Info("The key will be removed when the shell exits or when clear-key is called.")

			// Register cleanup on exit
			keymgmt.RegisterCleanupOnExit(tempKeyFile)

			return nil
		},
	}

	return cmd
}

// ClearKeyCmd returns the clear-key command
func ClearKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-key",
		Short: "Remove SOPS key when finished",
		Long:  `Clear the SOPS Age key from environment and remove temporary files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if SOPS_AGE_KEY_FILE is set
			keyFile := os.Getenv("SOPS_AGE_KEY_FILE")
			if keyFile == "" {
				logging.Info("No SOPS key was set.")
				return nil
			}

			// Check if it's a temporary file
			if tempDir := filepath.Dir(keyFile); strings.HasPrefix(filepath.Base(tempDir), "simple-sops-") {
				// Remove the temporary directory
				if err := keymgmt.CleanupTempAgeKeyFile(keyFile); err != nil {
					return fmt.Errorf("failed to remove temporary key file: %w", err)
				}
			}

			// Unset the environment variable
			os.Unsetenv("SOPS_AGE_KEY_FILE")

			logging.Success("SOPS key removed.")

			return nil
		},
	}

	return cmd
}

// GenerateKeyCmd returns the gen-key command
func GenerateKeyCmd() *cobra.Command {
	var (
		keyFile string
		force   bool
	)

	cmd := &cobra.Command{
		Use:   "gen-key",
		Short: "Generate a new Age key pair",
		Long:  `Generate a new Age key pair for use with SOPS.`,
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

			// Check if key file already exists
			expandedPath, err := keymgmt.ExpandPath(keyFile)
			if err != nil {
				return fmt.Errorf("failed to expand path: %w", err)
			}

			if _, err := os.Stat(expandedPath); err == nil && !force {
				return fmt.Errorf("key file already exists at %s. Use --force to overwrite", expandedPath)
			}

			// Generate the key
			if err := keymgmt.GenerateAgeKey(keyFile); err != nil {
				return fmt.Errorf("failed to generate Age key: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&keyFile, "key-file", "k", "", "Path to save the generated key (defaults to config setting)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing key file if it exists")

	return cmd
}
