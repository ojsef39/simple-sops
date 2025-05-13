package encrypt

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// testKey is a mock key for testing
const testKey = `# created: 2023-01-01T00:00:00Z
# public key: age123456789abcdef
AGE-SECRET-KEY-TESTKEYFORTESTING000000000000000000000000
`

// Mock for exec.Command
type mockExecCommand struct {
	cmd  string
	args []string
}

var (
	lastExecCommand mockExecCommand
	mockExecOutput  []byte
	mockExecError   error
	// Store the original execCommand function
	originalExecCommand = execCommand
)

// Mock exec.Command
func mockCommand(command string, args ...string) *exec.Cmd {
	lastExecCommand = mockExecCommand{cmd: command, args: args}

	// Create a fake command that returns our mock data
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	return cmd
}

// TestHelperProcess isn't a real test - it's used by the mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if mockExecError != nil {
		os.Exit(1)
	}
	if len(mockExecOutput) > 0 {
		os.Stdout.Write(mockExecOutput)
	}
	os.Exit(0)
}

func setupTestEnvironment(t *testing.T) (string, string, string, func()) {
	// Replace execCommand with mock
	execCommand = mockCommand

	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "encrypt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a test key file
	keyPath := filepath.Join(tempDir, "key.txt")
	err = os.WriteFile(keyPath, []byte(testKey), 0600)
	if err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// Create a test file to encrypt
	testFilePath := filepath.Join(tempDir, "test.env")
	err = os.WriteFile(testFilePath, []byte("TEST=value"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a SOPS config file
	configPath := filepath.Join(tempDir, ".sops.yaml")

	// Return cleanup function
	cleanup := func() {
		// Restore original execCommand
		execCommand = originalExecCommand
		os.RemoveAll(tempDir)

		// Reset mock state
		lastExecCommand = mockExecCommand{}
		mockExecOutput = nil
		mockExecError = nil
	}

	return keyPath, testFilePath, configPath, cleanup
}

func TestEncryptFile(t *testing.T) {
	keyPath, testFilePath, configPath, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Set up mock response
	mockExecOutput = []byte("Encryption successful")
	mockExecError = nil

	// Test encryption
	err := EncryptFile(testFilePath, keyPath, configPath)
	if err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	// Verify the command was called correctly
	if lastExecCommand.cmd != "sops" {
		t.Errorf("Expected 'sops' command, got '%s'", lastExecCommand.cmd)
	}

	hasEncryptArg := false
	hasInPlaceArg := false
	hasAgeArg := false

	for _, arg := range lastExecCommand.args {
		if arg == "--encrypt" {
			hasEncryptArg = true
		}
		if arg == "--in-place" {
			hasInPlaceArg = true
		}
		if arg == "--age" {
			hasAgeArg = true
		}
	}

	if !hasEncryptArg || !hasInPlaceArg || !hasAgeArg {
		t.Errorf("Missing required arguments to sops command: %v", lastExecCommand.args)
	}
}

// Additional tests for other encrypt package functions will be implemented here
