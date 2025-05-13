package encrypt

import (
	"fmt"
	"os"
	"path/filepath"
	"simple-sops/internal/config"
	"simple-sops/internal/keymgmt"
	"simple-sops/pkg/logging"
	"strings"
)

// EncryptFile encrypts a file using SOPS
func EncryptFile(filePath string, keyFile string, configPath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Get public key from key file
	pubKey, err := keymgmt.GetPublicKeyFromFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Load or create SOPS config
	sopsConfig, err := config.LoadSopsConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load SOPS config: %w", err)
	}

	// Add or update rule for this file
	fileName := filepath.Base(filePath)
	if err := config.AddCreationRule(sopsConfig, fileName, pubKey, ""); err != nil {
		return fmt.Errorf("failed to add rule to SOPS config: %w", err)
	}

	// Save the updated config
	if err := config.SaveSopsConfig(configPath, sopsConfig); err != nil {
		return fmt.Errorf("failed to save SOPS config: %w", err)
	}

	// Encrypt the file
	logging.Info("Encrypting %s...", filePath)

	// Set the SOPS_AGE_KEY_FILE environment variable
	cmd := execCommand("sops", "--encrypt", "--age", pubKey, "--in-place", filePath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("SOPS_AGE_KEY_FILE=%s", keyFile))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to encrypt file: %s\n%s", err, string(output))
	}

	logging.Success("File encrypted successfully: %s", filePath)
	return nil
}

// EncryptFilesWithMultipleKeys encrypts files with multiple keys
func EncryptFilesWithMultipleKeys(filePaths []string, keyFiles []string, pubKeys []string,
	alwaysUseOnePassword bool, opItems []keymgmt.OnePasswordItem,
) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("no files specified")
	}

	var keyPath string
	var err error

	// Create a temporary directory for the combined key
	tempDir, err := os.MkdirTemp("", "simple-sops-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up the combined key file
	combinedKeyPath := filepath.Join(tempDir, "combined-keys.txt")
	combinedFile, err := os.Create(combinedKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create combined key file: %w", err)
	}
	defer combinedFile.Close()

	// Track if we've added any keys
	keysAdded := false

	// First, add keys from 1Password if available
	if len(opItems) > 0 {
		logging.Debug("Getting keys from 1Password items...")
		opKeyPath, opIsTemp, err := keymgmt.GetKeysFromOnePassword(opItems)
		if err != nil {
			logging.Error("Failed to get keys from 1Password: %v", err)
		} else {
			// Read 1Password key file and add to combined file
			content, err := os.ReadFile(opKeyPath)
			if err != nil {
				logging.Error("Failed to read 1Password key file: %v", err)
			} else {
				if _, err := combinedFile.Write(content); err != nil {
					logging.Error("Failed to write 1Password key to combined file: %v", err)
				} else {
					// Add newline if needed
					if !strings.HasSuffix(string(content), "\n") {
						combinedFile.WriteString("\n")
					}
					keysAdded = true
					logging.Debug("Added keys from 1Password")
				}
			}

			// Clean up temporary 1Password key file
			if opIsTemp {
				if tempDir := filepath.Dir(opKeyPath); strings.HasPrefix(filepath.Base(tempDir), "simple-sops-") {
					os.RemoveAll(tempDir)
				}
			}
		}
	}

	// Next, add keys from file(s) if available
	if len(keyFiles) > 0 {
		logging.Debug("Adding keys from %d key files", len(keyFiles))
		for _, kf := range keyFiles {
			// Read the key file
			expandedPath, err := keymgmt.ExpandPath(kf)
			if err != nil {
				logging.Error("Failed to expand path %s: %v", kf, err)
				continue
			}

			content, err := os.ReadFile(expandedPath)
			if err != nil {
				logging.Error("Failed to read key file %s: %v", expandedPath, err)
				continue
			}

			// Append to combined file
			if _, err := combinedFile.Write(content); err != nil {
				logging.Error("Failed to write key to combined file: %v", err)
				continue
			}

			// Add newline if needed
			if !strings.HasSuffix(string(content), "\n") {
				combinedFile.WriteString("\n")
			}

			keysAdded = true
			logging.Debug("Added key from file: %s", kf)
		}
	}

	// If no keys added yet and alwaysUseOnePassword is true, try to get default key
	if !keysAdded && alwaysUseOnePassword {
		logging.Debug("Attempting to get default key from 1Password")
		defaultKeyPath, defaultIsTemp, err := keymgmt.EnsureAgeKey("", true, true)
		if err != nil {
			return fmt.Errorf("failed to get any keys: %w", err)
		}

		// Read default key file and add to combined file
		content, err := os.ReadFile(defaultKeyPath)
		if err != nil {
			logging.Error("Failed to read default key file: %v", err)
		} else {
			if _, err := combinedFile.Write(content); err != nil {
				logging.Error("Failed to write default key to combined file: %v", err)
			} else {
				// Add newline if needed
				if !strings.HasSuffix(string(content), "\n") {
					combinedFile.WriteString("\n")
				}
				keysAdded = true
				logging.Debug("Added default key")
			}
		}

		// Clean up default key file if temporary
		if defaultIsTemp {
			if tempDir := filepath.Dir(defaultKeyPath); strings.HasPrefix(filepath.Base(tempDir), "simple-sops-") {
				os.RemoveAll(tempDir)
			}
		}
	}

	// If still no keys, return error
	if !keysAdded {
		return fmt.Errorf("no valid keys found from any source")
	}

	// Close the file to ensure all writes are flushed
	combinedFile.Close()

	// Set keyPath to the combined key file path
	keyPath = combinedKeyPath

	// Get public keys from the combined key file
	var allPubKeys []string
	if len(pubKeys) > 0 {
		// Public keys explicitly provided
		allPubKeys = pubKeys
	} else {
		// Extract ALL public keys from the combined key file
		extractedPubKeys, err := keymgmt.GetAllPublicKeysFromFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to get public keys: %w", err)
		}
		allPubKeys = extractedPubKeys
		logging.Debug("Extracted %d public keys from combined key file", len(allPubKeys))
	}

	// Get the SOPS config path
	configPath, err := config.GetSopsConfigPath()
	if err != nil {
		return fmt.Errorf("failed to determine SOPS config path: %w", err)
	}

	// Process each file
	var encryptErr error
	for _, filePath := range filePaths {
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			logging.Error("File not found: %s", filePath)
			encryptErr = err
			continue
		}

		// Load or create SOPS config
		sopsConfig, err := config.LoadSopsConfig(configPath)
		if err != nil {
			logging.Error("Failed to load SOPS config: %v", err)
			encryptErr = err
			continue
		}

		// Combine multiple public keys with commas
		pubKeyStr := strings.Join(allPubKeys, ",")

		// Add or update rule for this file
		fileName := filepath.Base(filePath)
		if err := config.AddCreationRuleWithMultipleKeys(sopsConfig, fileName, pubKeyStr, ""); err != nil {
			logging.Error("Failed to add rule to SOPS config: %v", err)
			encryptErr = err
			continue
		}

		// Save the updated config
		if err := config.SaveSopsConfig(configPath, sopsConfig); err != nil {
			logging.Error("Failed to save SOPS config: %v", err)
			encryptErr = err
			continue
		}

		// Encrypt the file
		logging.Info("Encrypting %s with multiple keys...", filePath)

		// Use multiple Age recipients (comma-separated)
		cmd := execCommand("sops", "--encrypt", "--age", pubKeyStr, "--in-place", filePath)
		cmd.Env = append(os.Environ(), fmt.Sprintf("SOPS_AGE_KEY_FILE=%s", keyPath))

		output, err := cmd.CombinedOutput()
		if err != nil {
			logging.Error("Failed to encrypt file %s: %s\n%s", filePath, err, string(output))
			encryptErr = err
			continue
		}

		logging.Success("File encrypted successfully: %s", filePath)
	}

	return encryptErr
}

// EncryptFiles encrypts multiple files
func EncryptFiles(filePaths []string, keyFile string, alwaysUseOnePassword bool) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("no files specified")
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

	// Get the SOPS config path
	configPath, err := config.GetSopsConfigPath()
	if err != nil {
		return fmt.Errorf("failed to determine SOPS config path: %w", err)
	}

	// Process each file
	var encryptErr error
	for _, filePath := range filePaths {
		if err := EncryptFile(filePath, keyPath, configPath); err != nil {
			logging.Error("Failed to encrypt %s: %v", filePath, err)
			encryptErr = err
		}
	}

	return encryptErr
}

// SetEncryptionKeys sets the encryption keys for a specific file
func SetEncryptionKeys(filePath string, keyFile string, encryptedRegex string, alwaysUseOnePassword bool) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Check file extension
	ext := filepath.Ext(filePath)
	supportedExts := []string{".yaml", ".yml", ".json", ".ini", ".env"}
	isSupported := false
	for _, supportedExt := range supportedExts {
		if strings.EqualFold(ext, supportedExt) {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return fmt.Errorf("unsupported file type: %s", ext)
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

	// Get public key from key file
	pubKey, err := keymgmt.GetPublicKeyFromFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Get the SOPS config path
	configPath, err := config.GetSopsConfigPath()
	if err != nil {
		return fmt.Errorf("failed to determine SOPS config path: %w", err)
	}

	// Load or create SOPS config
	sopsConfig, err := config.LoadSopsConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load SOPS config: %w", err)
	}

	// Add or update rule for this file
	fileName := filepath.Base(filePath)
	if err := config.AddCreationRule(sopsConfig, fileName, pubKey, encryptedRegex); err != nil {
		return fmt.Errorf("failed to add rule to SOPS config: %w", err)
	}

	// Save the updated config
	if err := config.SaveSopsConfig(configPath, sopsConfig); err != nil {
		return fmt.Errorf("failed to save SOPS config: %w", err)
	}

	logging.Success("SOPS config updated for %s! Pattern: %s", fileName, encryptedRegex)
	logging.Info("")
	logging.Info("You can now encrypt your file with:")
	logging.Info("  simple-sops encrypt %s", filePath)

	return nil
}

// PredefinedEncryptionPatterns returns predefined encryption patterns
func PredefinedEncryptionPatterns() map[string]string {
	return map[string]string{
		"All values":            ".*",
		"Kubernetes":            "^(data|stringData|password|token|secret|key|cert|ca.crt|tls|ingress|backupTarget)",
		"Talos configuration":   "^(secrets|privateKey|token|key|crt|cert|password|secret|kubeconfig|talosconfig)",
		"Common sensitive data": "^(password|token|secret|key|auth|credential|private|apiKey|cert)",
	}
}
