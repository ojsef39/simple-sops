// internal/keymgmt/age_test.go
package keymgmt

import (
	"os"
	"path/filepath"
	"testing"
)

// mockKeyContent is a sample Age key file content for testing
const mockKeyContent = `# created: 2023-01-01T00:00:00Z
# public key: age123
AGE-SECRET-KEY-123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ
`

const mockKeyContent2 = `# created: 2023-01-02T00:00:00Z
# public key: age456
AGE-SECRET-KEY-ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789
`

func TestExtractPublicKey(t *testing.T) {
	// Test valid key extraction
	pubKey, err := extractPublicKey(mockKeyContent)
	if err != nil {
		t.Fatalf("Failed to extract public key: %v", err)
	}
	if pubKey != " age123" {
		t.Errorf("Expected public key 'age123', got '%s'", pubKey)
	}

	// Test with missing public key
	_, err = extractPublicKey("invalid content")
	if err == nil {
		t.Error("Expected error for invalid content, got nil")
	}
}

func TestGetPublicKeyFromFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "age-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary key file
	keyPath := filepath.Join(tempDir, "key.txt")
	err = os.WriteFile(keyPath, []byte(mockKeyContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// Test getting public key
	pubKey, err := GetPublicKeyFromFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to get public key: %v", err)
	}
	if pubKey != "age123" {
		t.Errorf("Expected public key 'age123', got '%s'", pubKey)
	}

	// Test with non-existent file
	_, err = GetPublicKeyFromFile(filepath.Join(tempDir, "nonexistent.txt"))
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestCreateTempAgeKeyFile(t *testing.T) {
	keyPath, err := CreateTempAgeKeyFile(mockKeyContent)
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	defer CleanupTempAgeKeyFile(keyPath)

	// Verify the file was created
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Temp key file was not created: %v", err)
	}

	// Verify the content was written
	content, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to read temp key file: %v", err)
	}
	if string(content) != mockKeyContent {
		t.Errorf("Key content mismatch")
	}
}

func TestGetAllPublicKeysFromFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "age-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a combined key file with multiple keys
	combinedContent := mockKeyContent + "\n" + mockKeyContent2
	keyPath := filepath.Join(tempDir, "combined-keys.txt")
	err = os.WriteFile(keyPath, []byte(combinedContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write combined key file: %v", err)
	}

	// Test getting all public keys
	pubKeys, err := GetAllPublicKeysFromFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to get public keys: %v", err)
	}
	if len(pubKeys) != 2 {
		t.Errorf("Expected 2 public keys, got %d", len(pubKeys))
	}
	if pubKeys[0] != "age123" || pubKeys[1] != "age456" {
		t.Errorf("Public keys mismatch: %v", pubKeys)
	}
}

// Mock implementation of exec.Command for testing
type MockCmd struct {
	expectedCmd  string
	expectedArgs []string
	output       []byte
	err          error
}

// Tests for 1Password integration with mocks will be implemented here
