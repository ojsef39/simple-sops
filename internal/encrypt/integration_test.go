package encrypt

import (
	"os"
	"path/filepath"
	"simple-sops/internal/keymgmt"
	"testing"
)

// Test keys for integration testing
const testKey1 = `# created: 2023-01-01T00:00:00Z
# public key: age123
AGE-SECRET-KEY-123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ
`

const testKey2 = `# created: 2023-01-02T00:00:00Z
# public key: age456
AGE-SECRET-KEY-ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789
`

func setupIntegrationTest(t *testing.T) (string, []string, string, string, func()) {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test key files
	keyPath1 := filepath.Join(tempDir, "key1.txt")
	err = os.WriteFile(keyPath1, []byte(testKey1), 0600)
	if err != nil {
		t.Fatalf("Failed to write key file 1: %v", err)
	}

	keyPath2 := filepath.Join(tempDir, "key2.txt")
	err = os.WriteFile(keyPath2, []byte(testKey2), 0600)
	if err != nil {
		t.Fatalf("Failed to write key file 2: %v", err)
	}

	// Create a combined key file
	combinedKeyPath := filepath.Join(tempDir, "combined-keys.txt")
	combinedContent := testKey1 + "\n" + testKey2
	err = os.WriteFile(combinedKeyPath, []byte(combinedContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write combined key file: %v", err)
	}

	// Create a test file to encrypt
	testFilePath := filepath.Join(tempDir, "test.env")
	err = os.WriteFile(testFilePath, []byte("TEST=value"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return combinedKeyPath, []string{keyPath1, keyPath2}, testFilePath, tempDir, cleanup
}

func TestMultipleKeysExtraction(t *testing.T) {
	combinedKeyPath, _, _, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Test extracting public keys from combined file
	pubKeys, err := keymgmt.GetAllPublicKeysFromFile(combinedKeyPath)
	if err != nil {
		t.Fatalf("Failed to get public keys: %v", err)
	}

	if len(pubKeys) != 2 {
		t.Errorf("Expected 2 public keys, got %d", len(pubKeys))
	}

	if pubKeys[0] != "age123" || pubKeys[1] != "age456" {
		t.Errorf("Public keys mismatch. Expected [age123, age456], got %v", pubKeys)
	}
}

// Additional integration tests can be added here
