package encrypt

import (
	"fmt"
	"os"
	"os/exec"
	"simple-sops/internal/keymgmt"
	"simple-sops/pkg/logging"
)

// DecryptionMode represents the mode for decryption
type DecryptionMode int

const (
	// DecryptModeStdout decrypts to stdout
	DecryptModeStdout DecryptionMode = iota
	// DecryptModeInPlace decrypts in-place
	DecryptModeInPlace
)

// DecryptFile decrypts a file using SOPS
func DecryptFile(filePath string, keyFile string, mode DecryptionMode) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Set up the command
	var cmd *exec.Cmd
	if mode == DecryptModeStdout {
		logging.Debug("Decrypting %s to stdout...", filePath)
		cmd = exec.Command("sops", "--decrypt", filePath)
		cmd.Stdout = os.Stdout
	} else {
		logging.Info("Decrypting %s in-place...", filePath)
		cmd = exec.Command("sops", "--decrypt", "--in-place", filePath)
	}

	// Set the SOPS_AGE_KEY_FILE environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("SOPS_AGE_KEY_FILE=%s", keyFile))
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	if mode == DecryptModeInPlace {
		logging.Success("File decrypted successfully: %s", filePath)
	}

	return nil
}

// DecryptFiles decrypts multiple files
func DecryptFiles(filePaths []string, keyFile string, useStdout bool, alwaysUseOnePassword bool) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("no files specified")
	}

	// Determine decryption mode based on useStdout flag
	mode := DecryptModeInPlace
	if useStdout {
		mode = DecryptModeStdout
	} else if len(filePaths) > 1 {
		// If multiple files and not stdout, ask user for mode
		choices := []string{
			"Print to screen/stdout (for piping to commands)",
			"Decrypt in-place (overwrites the encrypted file)",
		}
		choice, err := logging.PromptChoice("How would you like to decrypt file(s)?", choices)
		if err != nil {
			return fmt.Errorf("invalid choice: %w", err)
		}

		if choice == 1 {
			mode = DecryptModeStdout
		}
	}

	// Ensure we have the key available
	keyPath, isTemp, err := keymgmt.EnsureAgeKey(keyFile, true, alwaysUseOnePassword)
	if err != nil {
		return err
	}

	// Clean up the key if it's temporary
	if isTemp {
		defer keymgmt.CleanupTempAgeKeyFile(keyPath)
	}

	// Process each file
	var decryptErr error
	for _, filePath := range filePaths {
		if err := DecryptFile(filePath, keyPath, mode); err != nil {
			logging.Error("Failed to decrypt %s: %v", filePath, err)
			decryptErr = err
		}
	}

	return decryptErr
}

// EditFile opens an encrypted file for editing
func EditFile(filePath string, keyFile string, alwaysUseOnePassword bool) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Ensure we have the key available
	keyPath, isTemp, err := keymgmt.EnsureAgeKey(keyFile, true, alwaysUseOnePassword)
	if err != nil {
		return err
	}

	// Clean up the key if it's temporary
	if isTemp {
		defer keymgmt.CleanupTempAgeKeyFile(keyPath)
	}

	// Edit the file using SOPS
	logging.Info("Opening %s for editing...", filePath)

	cmd := exec.Command("sops", filePath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("SOPS_AGE_KEY_FILE=%s", keyPath))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while editing the file: %w", err)
	}

	logging.Success("File edited and saved successfully.")
	return nil
}

// DecryptToFile decrypts a file to a different file
func DecryptToFile(inputPath string, outputPath string, keyFile string) error {
	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", inputPath)
	}

	// Set up the command
	cmd := exec.Command("sops", "--decrypt", inputPath)

	// Create or truncate the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	cmd.Stdout = outputFile
	cmd.Stderr = os.Stderr

	// Set the SOPS_AGE_KEY_FILE environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("SOPS_AGE_KEY_FILE=%s", keyFile))

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	logging.Success("File decrypted successfully to: %s", outputPath)
	return nil
}
