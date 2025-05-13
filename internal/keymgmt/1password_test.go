package keymgmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Mock the 1Password CLI response
const mockOpResponse = `{
  "fields": [
    {
      "label": "text",
      "value": "# created: 2023-01-01T00:00:00Z\n# public key: age123\nAGE-SECRET-KEY-123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    }
  ]
}`

// Save original execCommand and lookPathFunc variables
var (
	originalExecCommand = execCommand
	originalLookPath    = lookPathFunc
)

// Mock for execCommand for 1Password tests
func mockOpCommand(command string, args ...string) *exec.Cmd {
	// If mocking 'op' command
	if command == "op" {
		// Create a fake command that returns our mock response
		cs := []string{"-test.run=TestOpHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "OP_TEST_RESPONSE=" + mockOpResponse}
		return cmd
	}

	// Also mock "which" or "where" commands to check for 'op' binary
	if command == "which" || command == "where" {
		if len(args) > 0 && args[0] == "op" {
			// Mock a successful "which op" response
			cs := []string{"-test.run=TestOpHelperProcess", "--", command}
			cs = append(cs, args...)
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "OP_PATH_RESPONSE=/usr/local/bin/op"}
			return cmd
		}
	}

	// Fall back to original for other commands
	return originalExecCommand(command, args...)
}

// Mock for exec.LookPath to avoid actually looking for 'op' in PATH
func mockLookPath(file string) (string, error) {
	if file == "op" {
		// Pretend "op" is in the PATH
		return "/usr/local/bin/op", nil
	}
	// For other commands, use the actual exec.LookPath
	return originalLookPath(file)
}

// TestOpHelperProcess mocks the 'op' command
func TestOpHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Check if this is a "which op" command
	if os.Getenv("OP_PATH_RESPONSE") != "" {
		os.Stdout.Write([]byte(os.Getenv("OP_PATH_RESPONSE")))
		os.Exit(0)
	}

	// Get the mock response from environment
	response := os.Getenv("OP_TEST_RESPONSE")
	if response != "" {
		os.Stdout.Write([]byte(response))
	}

	os.Exit(0)
}

func setupOpTest(t *testing.T) func() {
	// Replace execCommand with mock
	execCommand = mockOpCommand

	// Replace lookPathFunc with mock
	lookPathFunc = mockLookPath

	// Return cleanup function
	return func() {
		execCommand = originalExecCommand
		lookPathFunc = originalLookPath
	}
}

func TestGetKeyFromOnePassword(t *testing.T) {
	cleanup := setupOpTest(t)
	defer cleanup()

	// Test getting key from 1Password
	keyPath, err := GetKeyFromOnePassword(OnePasswordItem{
		ItemName:   "test-item",
		VaultName:  "test-vault",
		FieldLabel: "text",
	})
	if err != nil {
		t.Fatalf("GetKeyFromOnePassword failed: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Temp key file was not created: %v", err)
	}

	// Clean up the temporary file
	defer os.RemoveAll(filepath.Dir(keyPath))

	// Verify the content was written with the expected public key
	content, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to read temp key file: %v", err)
	}

	if !strings.Contains(string(content), "public key: age123") {
		t.Errorf("Key content does not contain expected public key")
	}
}

func TestGetKeysFromOnePassword(t *testing.T) {
	cleanup := setupOpTest(t)
	defer cleanup()

	// Test getting multiple keys from 1Password
	items := []OnePasswordItem{
		{
			ItemName:   "test-item1",
			VaultName:  "test-vault",
			FieldLabel: "text",
		},
		{
			ItemName:   "test-item2",
			VaultName:  "test-vault",
			FieldLabel: "text",
		},
	}

	keyPath, isTemp, err := GetKeysFromOnePassword(items)
	if err != nil {
		t.Fatalf("GetKeysFromOnePassword failed: %v", err)
	}

	if !isTemp {
		t.Errorf("Expected isTemp to be true")
	}

	// Verify the file was created
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Temp key file was not created: %v", err)
	}

	// Clean up the temporary file
	defer os.RemoveAll(filepath.Dir(keyPath))
}

func TestEnsureAgeKey(t *testing.T) {
	cleanup := setupOpTest(t)
	defer cleanup()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "age-key-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test key file
	keyPath := filepath.Join(tempDir, "key.txt")
	keyContent := `# created: 2023-01-01T00:00:00Z
# public key: age123
AGE-SECRET-KEY-123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ`

	err = os.WriteFile(keyPath, []byte(keyContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// Test with existing key file
	resultPath, isTemp, err := EnsureAgeKey(keyPath, false, false)
	if err != nil {
		t.Fatalf("EnsureAgeKey failed with existing file: %v", err)
	}
	if isTemp {
		t.Errorf("Expected isTemp to be false for existing key file")
	}
	if resultPath != keyPath {
		t.Errorf("Expected result path to be %s, got %s", keyPath, resultPath)
	}

	// Test with non-existent file but 1Password enabled
	resultPath, isTemp, err = EnsureAgeKey("nonexistent.txt", true, false)
	if err != nil {
		t.Fatalf("EnsureAgeKey failed with 1Password: %v", err)
	}
	if !isTemp {
		t.Errorf("Expected isTemp to be true when getting key from 1Password")
	}

	// Clean up the temporary file
	if isTemp {
		defer os.RemoveAll(filepath.Dir(resultPath))
	}

	// Test with always use 1Password flag
	resultPath, isTemp, err = EnsureAgeKey(keyPath, true, true)
	if err != nil {
		t.Fatalf("EnsureAgeKey failed with alwaysUseOnePassword: %v", err)
	}
	if !isTemp {
		t.Errorf("Expected isTemp to be true when always using 1Password")
	}

	// Clean up the temporary file
	if isTemp {
		defer os.RemoveAll(filepath.Dir(resultPath))
	}

	// Test with multiple 1Password items
	items := []OnePasswordItem{
		{
			ItemName:   "test-item1",
			VaultName:  "test-vault",
			FieldLabel: "text",
		},
	}

	resultPath, isTemp, err = EnsureAgeKey("", true, false, items...)
	if err != nil {
		t.Fatalf("EnsureAgeKey failed with multiple items: %v", err)
	}
	if !isTemp {
		t.Errorf("Expected isTemp to be true with multiple items")
	}

	// Clean up the temporary file
	if isTemp {
		defer os.RemoveAll(filepath.Dir(resultPath))
	}
}

// Test failure case with mocking
func TestOnePasswordCliNotFound(t *testing.T) {
	// Save the original lookPathFunc
	original := lookPathFunc

	// Replace with a mock that simulates CLI not found
	lookPathFunc = func(file string) (string, error) {
		if file == "op" {
			return "", os.ErrNotExist
		}
		return original(file)
	}

	// Restore original after test
	defer func() {
		lookPathFunc = original
	}()

	// Test failure to find 1Password CLI
	_, err := GetKeyFromOnePassword(OnePasswordItem{
		ItemName:   "test-item",
		VaultName:  "test-vault",
		FieldLabel: "text",
	})

	// This should fail since we've mocked lookPathFunc to simulate CLI not found
	if err == nil {
		t.Errorf("Expected GetKeyFromOnePassword to fail with CLI not found")
	}
}
