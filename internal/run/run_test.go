// internal/run/run_test.go
package run

import (
	"testing"
)

func TestParseRunCommand(t *testing.T) {
	// Test with minimum arguments
	encryptedFile, outputFile, command, commandArgs, err := ParseRunCommand([]string{"test.enc.yaml", "cat"})
	if err != nil {
		t.Fatalf("ParseRunCommand failed with minimum args: %v", err)
	}
	if encryptedFile != "test.enc.yaml" {
		t.Errorf("Expected encrypted file 'test.enc.yaml', got '%s'", encryptedFile)
	}
	if outputFile != "" {
		t.Errorf("Expected empty output file, got '%s'", outputFile)
	}
	if command != "cat" {
		t.Errorf("Expected command 'cat', got '%s'", command)
	}
	if len(commandArgs) != 0 {
		t.Errorf("Expected 0 command args, got %d", len(commandArgs))
	}

	// Test with output file
	encryptedFile, outputFile, command, commandArgs, err = ParseRunCommand([]string{"test.enc.yaml", "output.yaml", "kubectl", "apply", "-f"})
	if err != nil {
		t.Fatalf("ParseRunCommand failed with output file: %v", err)
	}
	if encryptedFile != "test.enc.yaml" {
		t.Errorf("Expected encrypted file 'test.enc.yaml', got '%s'", encryptedFile)
	}
	if outputFile != "output.yaml" {
		t.Errorf("Expected output file 'output.yaml', got '%s'", outputFile)
	}
	if command != "kubectl" {
		t.Errorf("Expected command 'kubectl', got '%s'", command)
	}
	if len(commandArgs) != 2 || commandArgs[0] != "apply" || commandArgs[1] != "-f" {
		t.Errorf("Command args mismatch: %v", commandArgs)
	}

	// Test with command containing spaces
	encryptedFile, outputFile, command, commandArgs, err = ParseRunCommand([]string{"test.enc.yaml", "cat test.env"})
	if err != nil {
		t.Fatalf("ParseRunCommand failed with command containing spaces: %v", err)
	}
	if command != "cat" {
		t.Errorf("Expected command 'cat', got '%s'", command)
	}
	if len(commandArgs) != 1 || commandArgs[0] != "test.env" {
		t.Errorf("Command args mismatch: %v", commandArgs)
	}

	// Test with insufficient arguments
	_, _, _, _, err = ParseRunCommand([]string{"test.enc.yaml"})
	if err == nil {
		t.Error("Expected error for insufficient arguments, got nil")
	}
}

func TestIsCommand(t *testing.T) {
	// Test known commands
	if !isCommand("cat") {
		t.Error("'cat' should be recognized as a command")
	}

	// Test flags
	if !isCommand("-f") {
		t.Error("'-f' should be recognized as a command flag")
	}

	// Test quoted strings
	if !isCommand("\"command with spaces\"") {
		t.Error("Quoted string should be recognized as a command")
	}

	// Test file paths
	if isCommand("/path/to/file") {
		t.Error("Absolute path shouldn't be recognized as a command")
	}
	if isCommand("./relative/path") {
		t.Error("Relative path shouldn't be recognized as a command")
	}
}
