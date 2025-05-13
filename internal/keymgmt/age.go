package keymgmt

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"simple-sops/pkg/logging"
	"strings"
	"syscall"
)

const (
	// DefaultKeyFile is the default path for the Age key file
	DefaultKeyFile = "~/.config/simple-sops/key.txt"
)

// KeyConfig represents the configuration for Age keys
type KeyConfig struct {
	// Path to the key file
	KeyFile string
	// KeyName is used when multiple keys are supported
	KeyName string
}

// GenerateAgeKey generates a new Age key pair and saves it to a file
func GenerateAgeKey(keyFile string) error {
	logging.Debug("Generating new Age key pair")

	// Expand homedir if needed
	expandedPath, err := expandPath(keyFile)
	if err != nil {
		return fmt.Errorf("failed to expand path: %w", err)
	}

	// Check if directory exists, create if not
	dir := filepath.Dir(expandedPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Check if file already exists
	if _, err := os.Stat(expandedPath); err == nil {
		return fmt.Errorf("key file already exists at %s", expandedPath)
	}

	// Generate key using age-keygen
	cmd := exec.Command("age-keygen")
	var keyOutput bytes.Buffer
	cmd.Stdout = &keyOutput

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate Age key: %w", err)
	}

	// Save key to file
	if err := os.WriteFile(expandedPath, keyOutput.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to save key to file: %w", err)
	}

	logging.Success("Generated Age key pair and saved to %s", expandedPath)
	logging.Info("Make sure to back up this key file securely!")

	// Extract and display public key
	pubKey, err := extractPublicKey(keyOutput.String())
	if err != nil {
		logging.Error("Could not extract public key from generated key")
	} else {
		logging.Info("Public key: %s", pubKey)
	}

	return nil
}

// extractPublicKey extracts the public key from an Age key file content
func extractPublicKey(keyContent string) (string, error) {
	lines := strings.Split(keyContent, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# public key:") {
			return strings.TrimPrefix(line, "# public key:"), nil
		}
	}
	return "", fmt.Errorf("public key not found in key content")
}

// GetPublicKeyFromFile extracts the public key from an Age key file
func GetPublicKeyFromFile(keyFile string) (string, error) {
	expandedPath, err := expandPath(keyFile)
	if err != nil {
		return "", fmt.Errorf("failed to expand path: %w", err)
	}

	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	pubKey, err := extractPublicKey(string(content))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(pubKey), nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}

// LoadAgeKey loads an Age key from a file and returns the path to the key file
func LoadAgeKey(keyFile string) (string, error) {
	expandedPath, err := expandPath(keyFile)
	if err != nil {
		return "", fmt.Errorf("failed to expand path: %w", err)
	}

	// Check if the key file exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return "", fmt.Errorf("key file not found at %s", expandedPath)
	}

	// Verify the key file contains a valid Age key
	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	if !strings.Contains(string(content), "AGE-SECRET-KEY-") {
		return "", fmt.Errorf("key file does not contain a valid Age key")
	}

	return expandedPath, nil
}

// CreateTempAgeKeyFile creates a temporary file with an Age key and returns the path
func CreateTempAgeKeyFile(keyContent string) (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "simple-sops-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Create key file
	tempKeyFile := filepath.Join(tempDir, "age-key.txt")
	if err := os.WriteFile(tempKeyFile, []byte(keyContent), 0600); err != nil {
		os.RemoveAll(tempDir) // Clean up if we can't write
		return "", fmt.Errorf("failed to write temporary key file: %w", err)
	}

	return tempKeyFile, nil
}

// CleanupTempAgeKeyFile removes a temporary Age key file and its directory
func CleanupTempAgeKeyFile(keyFile string) error {
	// Get the directory containing the key file
	dir := filepath.Dir(keyFile)

	// Only remove if it looks like our temp directory
	if strings.HasPrefix(filepath.Base(dir), "simple-sops-") {
		return os.RemoveAll(dir)
	}

	return fmt.Errorf("not a simple-sops temporary directory")
}

// RegisterCleanupOnExit registers a cleanup function for when the process exits
func RegisterCleanupOnExit(keyFile string) {
	// Go doesn't have a direct equivalent to Fish's --on-process-exit
	// You can use signal handling instead
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		CleanupTempAgeKeyFile(keyFile)
		os.Exit(1)
	}()
}

// ExpandPath is just a renamed version of the internal expandPath function
func ExpandPath(path string) (string, error) {
	return expandPath(path)
}

func GetAllPublicKeysFromFile(keyFile string) ([]string, error) {
	expandedPath, err := expandPath(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}

	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var pubKeys []string

	for _, line := range lines {
		if strings.HasPrefix(line, "# public key:") {
			pubKey := strings.TrimSpace(strings.TrimPrefix(line, "# public key:"))
			if pubKey != "" {
				pubKeys = append(pubKeys, pubKey)
			}
		}
	}

	if len(pubKeys) == 0 {
		return nil, fmt.Errorf("no public keys found in key file")
	}

	return pubKeys, nil
}
