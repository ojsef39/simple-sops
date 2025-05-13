package run

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"simple-sops/internal/encrypt"
	"simple-sops/internal/keymgmt"
	"simple-sops/pkg/logging"
	"strings"
	"syscall"
)

// RunWithEncryptedFile executes a command with a temporarily decrypted file
func RunWithEncryptedFile(encryptedFilePath string, outputPath string, command string, args []string, keyFile string, alwaysUseOnePassword bool) error {
	// Check if encrypted file exists
	if _, err := os.Stat(encryptedFilePath); os.IsNotExist(err) {
		return fmt.Errorf("encrypted file not found: %s", encryptedFilePath)
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

	// Determine the output path
	tempFileNeeded := outputPath == ""
	var tempDir string

	if tempFileNeeded {
		// Create temporary directory for decrypted file
		tempDir, err = os.MkdirTemp("", "simple-sops-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		// Generate a temporary file path
		outputPath = filepath.Join(tempDir, filepath.Base(encryptedFilePath)+".plain")
	} else {
		// For user-specified output path, ensure we clean it up afterwards
		defer func() {
			if err := os.Remove(outputPath); err != nil {
				logging.Debug("Failed to remove output file %s: %v", outputPath, err)
			} else {
				logging.Debug("Removed output file %s", outputPath)
			}
		}()
	}

	// Decrypt the file to the output path
	if err := encrypt.DecryptToFile(encryptedFilePath, outputPath, keyPath); err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	// Replace any references to the original file in command arguments
	// with the path to the decrypted file
	originalFileName := filepath.Base(encryptedFilePath)
	for i, arg := range args {
		if arg == originalFileName || arg == encryptedFilePath {
			args[i] = outputPath
		}
	}

	// Check if the command itself is the filename
	if command == originalFileName || command == encryptedFilePath {
		command = outputPath
	}

	// Prepare to execute the command
	logging.Info("Running command: %s %s", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)

	// Add output path to environment variables
	cmd.Env = append(os.Environ(), fmt.Sprintf("DECRYPTED_FILE=%s", outputPath))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set up signal handling to ensure cleanup
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Handle process termination in a separate goroutine
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	// Wait for either command completion or signal
	select {
	case err := <-cmdDone:
		// Command completed
		if err != nil {
			return fmt.Errorf("command execution failed: %w", err)
		}
	case sig := <-signalChan:
		// Signal received, kill process
		logging.Info("Received signal %v, terminating command", sig)
		if err := cmd.Process.Kill(); err != nil {
			logging.Error("Failed to kill process: %v", err)
		}
		return fmt.Errorf("command terminated by signal: %v", sig)
	}

	logging.Success("Command completed successfully")

	return nil
}

// ParseRunCommand parses the run command arguments
func ParseRunCommand(args []string) (encryptedFile string, outputFile string, command string, commandArgs []string, err error) {
	if len(args) < 2 {
		return "", "", "", nil, fmt.Errorf("insufficient arguments. Usage: simple-sops run [encrypted-file] [output-file (optional)] [command...]")
	}

	encryptedFile = args[0]

	// Check if the second argument is an output file or the command
	if len(args) > 2 && !isCommand(args[1]) {
		outputFile = args[1]

		// For commands like "cat test.env", we need to handle them properly
		if strings.Contains(args[2], " ") {
			// Split the command by spaces
			parts := strings.Fields(args[2])
			command = parts[0]
			commandArgs = append(parts[1:], args[3:]...)
		} else {
			command = args[2]
			commandArgs = args[3:]
		}
	} else {
		// For commands like "cat test.env", we need to handle them properly
		if strings.Contains(args[1], " ") {
			// Split the command by spaces
			parts := strings.Fields(args[1])
			command = parts[0]
			commandArgs = append(parts[1:], args[2:]...)
		} else {
			command = args[1]
			commandArgs = args[2:]
		}
	}

	return encryptedFile, outputFile, command, commandArgs, nil
}

// isCommand checks if the argument is likely a command
func isCommand(arg string) bool {
	// If the argument starts with a quote, it's likely a command
	if (len(arg) > 0 && (arg[0] == '"' || arg[0] == '\'')) ||
		// If the argument starts with a dash, it's likely a flag for a command
		(len(arg) > 0 && arg[0] == '-') {
		return true
	}

	// Check if the string contains a path separator, which might indicate it's a file
	if filepath.IsAbs(arg) || arg == "." || arg == ".." ||
		filepath.Base(arg) != arg {
		return false
	}

	// If we can find the command in PATH, it's likely a command
	_, err := exec.LookPath(arg)
	return err == nil
}
