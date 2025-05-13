package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"simple-sops/internal/config"
	"simple-sops/pkg/logging"
)

// ConfigCmd returns the config command
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show current SOPS configurations",
		Long:  `Display the current SOPS configuration settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the SOPS config path
			configPath, err := config.GetSopsConfigPath()
			if err != nil {
				return fmt.Errorf("failed to determine SOPS config path: %w", err)
			}

			// Load the SOPS config
			sopsConfig, err := config.LoadSopsConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load SOPS config: %w", err)
			}

			// Check if config file exists and has rules
			if len(sopsConfig.CreationRules) == 0 {
				logging.Info("No SOPS configuration found at %s.", configPath)
				return nil
			}

			// Display config
			logging.Info("Current SOPS configuration (%s):", configPath)
			logging.Info("--------------------------")

			// Display rules
			logging.Info("Rules:")
			for _, rule := range sopsConfig.CreationRules {
				logging.Info("")
				logging.Info("File pattern: %s", rule.PathRegex)
				logging.Info("  Age key: %s", rule.Age)

				if rule.EncryptedRegex != "" {
					logging.Info("  Encrypts: %s", rule.EncryptedRegex)
				}
			}

			logging.Info("")
			logging.Info("This configuration will be used when encrypting files with SOPS.")

			return nil
		},
	}

	return cmd
}

// CleanConfigCmd returns the clean-config command
func CleanConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean-config",
		Short: "Clean orphaned rules from SOPS config",
		Long:  `Remove rules for files that no longer exist from the SOPS configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the SOPS config path
			configPath, err := config.GetSopsConfigPath()
			if err != nil {
				return fmt.Errorf("failed to determine SOPS config path: %w", err)
			}

			// Load the SOPS config
			sopsConfig, err := config.LoadSopsConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load SOPS config: %w", err)
			}

			// Check if config file exists and has rules
			if len(sopsConfig.CreationRules) == 0 {
				logging.Info("No SOPS configuration found at %s. Nothing to clean up.", configPath)
				return nil
			}

			// Clean orphaned rules
			orphanedCount, err := config.CleanOrphanedRules(sopsConfig)
			if err != nil {
				return fmt.Errorf("failed to clean orphaned rules: %w", err)
			}

			if orphanedCount == 0 {
				logging.Info("No orphaned rules found in %s.", configPath)
				return nil
			}

			// Ask for confirmation
			if !logging.Confirm(fmt.Sprintf("Found %d orphaned rules in %s. Do you want to remove them?", orphanedCount, configPath)) {
				logging.Info("Operation cancelled.")
				return nil
			}

			// Save the updated config
			if err := config.SaveSopsConfig(configPath, sopsConfig); err != nil {
				return fmt.Errorf("failed to save SOPS config: %w", err)
			}

			logging.Success("Removed %d orphaned rules from %s.", orphanedCount, configPath)

			// Check if the config is now empty
			if len(sopsConfig.CreationRules) == 0 {
				if logging.Confirm(fmt.Sprintf("No rules remain in %s. Do you want to remove it?", configPath)) {
					if err := os.Remove(configPath); err != nil {
						return fmt.Errorf("failed to remove empty config file: %w", err)
					}
					logging.Success("%s removed since it no longer contains any rules.", configPath)
				}
			} else {
				logging.Info("Remaining rules in %s: %d", configPath, len(sopsConfig.CreationRules))
			}

			return nil
		},
	}

	return cmd
}

// RemoveCmd returns the rm command
func RemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm [file...]",
		Short: "Remove files and their SOPS configurations",
		Long:  `Remove files and their corresponding rules from the SOPS configuration.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the SOPS config path
			configPath, err := config.GetSopsConfigPath()
			if err != nil {
				return fmt.Errorf("failed to determine SOPS config path: %w", err)
			}

			// Load the SOPS config
			sopsConfig, err := config.LoadSopsConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load SOPS config: %w", err)
			}

			for _, filePath := range args {
				fileName := filepath.Base(filePath)

				// Check if the file exists
				fileExists := true
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					logging.Info("Warning: File %s not found.", filePath)
					fileExists = false

					if !logging.Confirm("Do you want to still check and clean up SOPS configuration for this file?") {
						logging.Info("Skipping %s...", filePath)
						continue
					}
				} else if fileExists {
					// Prompt for confirmation
					if !logging.Confirm(fmt.Sprintf("This will remove the file %s and its SOPS configuration. Are you sure?", filePath)) {
						logging.Info("Skipping %s...", filePath)
						continue
					}

					// Remove the file
					if err := os.Remove(filePath); err != nil {
						logging.Error("Failed to remove file %s: %v", filePath, err)
					} else {
						logging.Success("File %s removed.", filePath)
					}
				}

				// Check if there's a rule for this file
				ruleExists := false
				for _, rule := range sopsConfig.CreationRules {
					if rule.PathRegex == fileName {
						ruleExists = true
						break
					}
				}

				if !ruleExists {
					logging.Info("No configuration found for %s in %s.", fileName, configPath)
					continue
				}

				// Remove the rule
				if err := config.RemoveCreationRule(sopsConfig, fileName); err != nil {
					logging.Error("Failed to remove rule for %s: %v", fileName, err)
					continue
				}

				// Save the updated config
				if err := config.SaveSopsConfig(configPath, sopsConfig); err != nil {
					logging.Error("Failed to save SOPS config: %v", err)
					continue
				}

				logging.Success("SOPS configuration for %s removed successfully.", fileName)
			}

			// Check if the config is now empty
			if len(sopsConfig.CreationRules) == 0 {
				if logging.Confirm(fmt.Sprintf("No rules remain in %s. Do you want to remove it?", configPath)) {
					if err := os.Remove(configPath); err != nil {
						return fmt.Errorf("failed to remove empty config file: %w", err)
					}
					logging.Success("%s removed since it no longer contains any rules.", configPath)
				}
			} else {
				logging.Info("Remaining rules in %s: %d", configPath, len(sopsConfig.CreationRules))
			}

			return nil
		},
	}

	return cmd
}
