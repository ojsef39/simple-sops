package commands

import (
	"fmt"
	"simple-sops/internal/config"
	"simple-sops/internal/encrypt"

	"github.com/spf13/cobra"
)

// DecryptCmd returns the decrypt command
func DecryptCmd() *cobra.Command {
	var (
		keyFile   string
		useStdout bool
	)

	cmd := &cobra.Command{
		Use:   "decrypt [file...]",
		Short: "Decrypt one or more files",
		Long:  `Decrypt one or more files encrypted with SOPS.`,
		Args:  cobra.MinimumNArgs(1),
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

			// Decrypt the files
			if err := encrypt.DecryptFiles(args, keyFile, useStdout, appConfig.AlwaysUseOnePassword); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&keyFile, "key-file", "k", "", "Age key file to use (defaults to config setting)")
	cmd.Flags().BoolVar(&useStdout, "stdout", false, "Output to stdout instead of files")

	return cmd
}
