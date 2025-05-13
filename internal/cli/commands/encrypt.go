package commands

import (
	"fmt"
	"simple-sops/internal/config"
	"simple-sops/internal/encrypt"
	"simple-sops/internal/keymgmt"
	"strings"

	"github.com/spf13/cobra"
)

// EncryptCmd returns the encrypt command
func EncryptCmd() *cobra.Command {
	var (
		keyFile     string
		keyFiles    string
		opItems     []string
		opVaults    []string
		opFieldName string
	)

	cmd := &cobra.Command{
		Use:   "encrypt [file...]",
		Short: "Encrypt one or more files with Age",
		Long:  `Encrypt one or more files using SOPS with Age encryption.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load application config
			appConfig, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// If both a key file is specified AND AlwaysUseOnePassword is true,
			// use both keys for encryption
			if keyFile != "" && appConfig.AlwaysUseOnePassword && appConfig.OnePasswordEnabled {
				// Get key from 1Password
				opItem := keymgmt.DefaultOnePasswordItem

				// Create a slice of key files containing the specified key file
				keyFilesSlice := []string{keyFile}

				// Create 1Password items slice
				opItemsSlice := []keymgmt.OnePasswordItem{opItem}

				// Use multiple keys for encryption (both 1Password and file)
				return encrypt.EncryptFilesWithMultipleKeys(
					args,
					keyFilesSlice,
					nil,  // pubKeys can be nil, they'll be extracted from the key files
					true, // Always use 1Password
					opItemsSlice)
			}

			// Process multiple keys if specified
			var multipleKeyFiles []string
			if keyFiles != "" {
				multipleKeyFiles = strings.Split(keyFiles, ",")
			} else if keyFile != "" {
				multipleKeyFiles = []string{keyFile}
			} else {
				multipleKeyFiles = []string{appConfig.KeyFile}
			}

			// Process 1Password items if specified
			var opItemsList []keymgmt.OnePasswordItem
			if len(opItems) > 0 {
				// Build the list of 1Password items
				for i, item := range opItems {
					vault := "Personal" // Default vault
					if i < len(opVaults) {
						vault = opVaults[i]
					}

					field := "text" // Default field name
					if opFieldName != "" {
						field = opFieldName
					}

					opItemsList = append(opItemsList, keymgmt.OnePasswordItem{
						ItemName:   item,
						VaultName:  vault,
						FieldLabel: field,
					})
				}

				// Encrypt using multiple 1Password items
				if err := encrypt.EncryptFilesWithMultipleKeys(
					args,
					nil,
					nil,
					appConfig.AlwaysUseOnePassword,
					opItemsList); err != nil {
					return err
				}
			} else if len(multipleKeyFiles) > 1 {
				// Encrypt with multiple key files
				if err := encrypt.EncryptFilesWithMultipleKeys(
					args,
					multipleKeyFiles,
					nil,
					appConfig.AlwaysUseOnePassword,
					nil); err != nil {
					return err
				}
			} else {
				// Standard single key encryption
				if err := encrypt.EncryptFiles(args, multipleKeyFiles[0], appConfig.AlwaysUseOnePassword); err != nil {
					return err
				}
			}

			return nil
		},
	}

	// Add flags for key specification
	cmd.Flags().StringVarP(&keyFile, "key-file", "k", "", "Age key file to use (defaults to config setting)")
	cmd.Flags().StringVar(&keyFiles, "key-files", "", "Comma-separated list of Age key files to use")
	cmd.Flags().StringSliceVar(&opItems, "op-items", nil, "1Password items to fetch keys from")
	cmd.Flags().StringSliceVar(&opVaults, "op-vaults", nil, "1Password vaults for the items (defaults to 'Personal' if not specified)")
	cmd.Flags().StringVar(&opFieldName, "op-field", "", "Field name in 1Password items (defaults to 'text')")

	return cmd
}
