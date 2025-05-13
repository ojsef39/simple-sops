package keymgmt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"simple-sops/pkg/logging"
	"strings"
)

// OnePasswordItem represents a key stored in 1Password
type OnePasswordItem struct {
	// The name of the item in 1Password
	ItemName string
	// The vault where the item is stored
	VaultName string
	// The field label containing the key
	FieldLabel string
}

// DefaultOnePasswordItem is the default item for 1Password
var DefaultOnePasswordItem = OnePasswordItem{
	ItemName:   "SOPS_AGE_KEY_FILE",
	VaultName:  "Personal",
	FieldLabel: "text",
}

// For backward compatibility with existing code
var DefaultOnePasswordConfig = DefaultOnePasswordItem

// 1Password JSON structures
type opItemResponse struct {
	Fields []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"fields"`
}

// GetKeyFromOnePassword retrieves an Age key from a single 1Password item and saves it to a temporary file
func GetKeyFromOnePassword(item OnePasswordItem) (string, error) {
	logging.Debug("Fetching SOPS key from 1Password item %s in vault %s...", item.ItemName, item.VaultName)

	// Check if 1Password CLI is available
	if err := checkOnePasswordCLI(); err != nil {
		return "", err
	}

	// Get the key content from 1Password
	keyContent, err := getKeyContentFromOnePassword(item)
	if err != nil {
		return "", err
	}

	// Create a temporary file for the key
	return CreateTempAgeKeyFile(keyContent)
}

// GetKeysFromOnePassword retrieves multiple Age keys from 1Password items and combines them into a single temporary file
func GetKeysFromOnePassword(items []OnePasswordItem) (string, bool, error) {
	logging.Debug("Fetching multiple SOPS keys from 1Password...")

	// Check if 1Password CLI is available
	if err := checkOnePasswordCLI(); err != nil {
		return "", false, err
	}

	// Create a temporary directory for the keys
	tempDir, err := os.MkdirTemp("", "simple-sops-*")
	if err != nil {
		return "", false, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Create a combined key file
	tempKeyFile := filepath.Join(tempDir, "age-keys.txt")
	keyFile, err := os.Create(tempKeyFile)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", false, fmt.Errorf("failed to create temporary key file: %w", err)
	}
	defer keyFile.Close()

	// Fetch each key and append to the combined file
	for _, item := range items {
		logging.Debug("Fetching key from item: %s in vault: %s", item.ItemName, item.VaultName)

		// Get key content from 1Password
		keyContent, err := getKeyContentFromOnePassword(item)
		if err != nil {
			logging.Debug("Failed to get key from 1Password item %s: %v", item.ItemName, err)
			continue
		}

		// Write the key to the combined file with a newline if needed
		if !strings.HasSuffix(keyContent, "\n") {
			keyContent += "\n"
		}
		if _, err := keyFile.WriteString(keyContent); err != nil {
			os.RemoveAll(tempDir)
			return "", false, fmt.Errorf("failed to write key to temporary file: %w", err)
		}

		logging.Debug("Successfully added key from item: %s", item.ItemName)
	}

	return tempKeyFile, true, nil
}

// getKeyContentFromOnePassword retrieves the key content from a 1Password item
func getKeyContentFromOnePassword(item OnePasswordItem) (string, error) {
	// Get the key from 1Password
	cmd := execCommand("op", "item", "get", item.ItemName, "--vault", item.VaultName, "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get key from 1Password: %w", err)
	}

	// Parse the JSON response
	var response opItemResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return "", fmt.Errorf("failed to parse 1Password response: %w", err)
	}

	// Find the field with the key
	var keyContent string
	for _, field := range response.Fields {
		if field.Label == item.FieldLabel {
			keyContent = field.Value
			break
		}
	}

	if keyContent == "" {
		return "", fmt.Errorf("no field with label '%s' found in 1Password item", item.FieldLabel)
	}

	return keyContent, nil
}

// checkOnePasswordCLI checks if the 1Password CLI is available
func checkOnePasswordCLI() error {
	_, err := lookPathFunc("op")
	if err != nil {
		return fmt.Errorf("1Password CLI (op) not found in PATH. Please install it and try again")
	}

	return nil
}

// EnsureAgeKey makes sure an Age key is available, either from a file or from 1Password
// Now supports multiple 1Password items through the opItems parameter
func EnsureAgeKey(keyFile string, useOnePassword bool, alwaysUseOnePassword bool, opItems ...OnePasswordItem) (string, bool, error) {
	// If AlwaysUseOnePassword is true, we always try to get the key from 1Password first
	if alwaysUseOnePassword && useOnePassword {
		// Check if we have multiple items specified
		if len(opItems) > 0 {
			logging.Debug("Auto-fetching multiple Age keys from 1Password")
			tempKeyFile, isTemp, err := GetKeysFromOnePassword(opItems)
			if err == nil {
				logging.Debug("Successfully retrieved multiple Age keys from 1Password")
				return tempKeyFile, isTemp, nil
			}
			logging.Debug("Failed to get keys from 1Password: %v", err)
		} else {
			// Use default item
			logging.Debug("Auto-fetching Age key from 1Password")
			tempKeyFile, err := GetKeyFromOnePassword(DefaultOnePasswordItem)
			if err == nil {
				logging.Debug("Successfully retrieved Age key from 1Password")
				return tempKeyFile, true, nil
			}
			logging.Debug("Failed to get key from 1Password: %v", err)
		}
	}

	// Check if key file is specified and exists
	if keyFile != "" {
		expandedPath, err := expandPath(keyFile)
		if err != nil {
			return "", false, err
		}

		if _, err := os.Stat(expandedPath); err == nil {
			logging.Debug("Using specified Age key file: %s", expandedPath)
			return expandedPath, false, nil
		}
	}

	// If allowed to use 1Password, try to get the key from there
	if useOnePassword {
		// Check if we have multiple items specified
		if len(opItems) > 0 {
			logging.Debug("Trying to get multiple Age keys from 1Password")
			tempKeyFile, isTemp, err := GetKeysFromOnePassword(opItems)
			if err == nil {
				logging.Debug("Successfully retrieved multiple Age keys from 1Password")
				return tempKeyFile, isTemp, nil
			}
			logging.Debug("Failed to get keys from 1Password: %v", err)
		} else {
			// Use default item
			logging.Debug("Trying to get Age key from 1Password")
			tempKeyFile, err := GetKeyFromOnePassword(DefaultOnePasswordItem)
			if err == nil {
				logging.Debug("Successfully retrieved Age key from 1Password")
				return tempKeyFile, true, nil
			}
			logging.Debug("Failed to get key from 1Password: %v", err)
		}
	}

	// If we got here, we couldn't find a key
	return "", false, fmt.Errorf("no Age key available. Use gen-key to create one or specify an existing key file")
}
