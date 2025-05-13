package encrypt

import (
	"os"
	"path/filepath"
	"simple-sops/pkg/logging"
	"testing"
)

func TestDecryptFile(t *testing.T) {
	keyPath, testFilePath, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Set up mock response
	mockExecOutput = []byte("Decryption successful")
	mockExecError = nil

	// Test decryption to stdout
	err := DecryptFile(testFilePath, keyPath, DecryptModeStdout)
	if err != nil {
		t.Fatalf("DecryptFile failed in stdout mode: %v", err)
	}

	// Verify the command was called correctly
	if lastExecCommand.cmd != "sops" {
		t.Errorf("Expected 'sops' command, got '%s'", lastExecCommand.cmd)
	}

	hasDecryptArg := false
	for _, arg := range lastExecCommand.args {
		if arg == "--decrypt" {
			hasDecryptArg = true
		}
		if arg == "--in-place" {
			t.Errorf("Should not have --in-place arg in stdout mode")
		}
	}

	if !hasDecryptArg {
		t.Errorf("Missing --decrypt argument to sops command: %v", lastExecCommand.args)
	}

	// Reset mock state
	lastExecCommand = mockExecCommand{}

	// Test decryption in-place
	err = DecryptFile(testFilePath, keyPath, DecryptModeInPlace)
	if err != nil {
		t.Fatalf("DecryptFile failed in in-place mode: %v", err)
	}

	// Verify the command was called correctly
	if lastExecCommand.cmd != "sops" {
		t.Errorf("Expected 'sops' command, got '%s'", lastExecCommand.cmd)
	}

	hasDecryptArg = false
	hasInPlaceArg := false
	for _, arg := range lastExecCommand.args {
		if arg == "--decrypt" {
			hasDecryptArg = true
		}
		if arg == "--in-place" {
			hasInPlaceArg = true
		}
	}

	if !hasDecryptArg || !hasInPlaceArg {
		t.Errorf("Missing required arguments to sops command: %v", lastExecCommand.args)
	}
}

func TestDecryptFiles(t *testing.T) {
	keyPath, testFilePath, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Set up mocks for the prompt
	restoreMock := logging.MockPromptChoice(2) // Always choose the second option (in-place)
	defer restoreMock()                        // Restore original function after test

	// Create a second test file
	testDir := filepath.Dir(testFilePath)
	testFilePath2 := filepath.Join(testDir, "test2.env")
	err := os.WriteFile(testFilePath2, []byte("TEST2=value"), 0644)
	if err != nil {
		t.Fatalf("Failed to write second test file: %v", err)
	}

	// Set up mock response
	mockExecOutput = []byte("Decryption successful")
	mockExecError = nil

	// Test decryption of multiple files with stdout option
	filePaths := []string{testFilePath, testFilePath2}
	err = DecryptFiles(filePaths, keyPath, true, false)
	if err != nil {
		t.Fatalf("DecryptFiles failed with stdout option: %v", err)
	}

	// Test decryption of multiple files with in-place option
	// This would normally prompt, but we've mocked the prompt function
	err = DecryptFiles(filePaths, keyPath, false, false)
	if err != nil {
		t.Fatalf("DecryptFiles failed with in-place option: %v", err)
	}

	// Test with empty file list
	err = DecryptFiles([]string{}, keyPath, false, false)
	if err == nil {
		t.Error("DecryptFiles should fail with empty file list")
	}
}
