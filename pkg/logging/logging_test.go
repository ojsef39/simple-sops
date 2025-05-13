// pkg/logging/logging_test.go
package logging

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func captureOutput(f func()) string {
	// Save original stdout
	oldStdout := os.Stdout

	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function that produces output
	f()

	// Close the write end of the pipe
	w.Close()

	// Restore original stdout
	os.Stdout = oldStdout

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

func captureError(f func()) string {
	// Save original stderr
	oldStderr := os.Stderr

	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Call the function that produces output
	f()

	// Close the write end of the pipe
	w.Close()

	// Restore original stderr
	os.Stderr = oldStderr

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

func TestDebug(t *testing.T) {
	// Test debug mode disabled
	SetDebugMode(false)
	output := captureOutput(func() {
		Debug("test message")
	})
	if output != "" {
		t.Errorf("Expected no output with debug mode disabled, got: %s", output)
	}

	// Test debug mode enabled
	SetDebugMode(true)
	output = captureOutput(func() {
		Debug("test message")
	})
	if output == "" {
		t.Error("Expected output with debug mode enabled, got nothing")
	}

	// Reset for other tests
	SetDebugMode(false)
}

func TestInfo(t *testing.T) {
	// Test quiet mode disabled
	SetQuietMode(false)
	output := captureOutput(func() {
		Info("test message")
	})
	if output == "" {
		t.Error("Expected output with quiet mode disabled, got nothing")
	}

	// Test quiet mode enabled
	SetQuietMode(true)
	output = captureOutput(func() {
		Info("test message")
	})
	if output != "" {
		t.Errorf("Expected no output with quiet mode enabled, got: %s", output)
	}

	// Reset for other tests
	SetQuietMode(false)
}

func TestError(t *testing.T) {
	// Error should always output regardless of quiet mode
	SetQuietMode(false)
	output := captureError(func() {
		Error("test error")
	})
	if output == "" {
		t.Error("Expected error output, got nothing")
	}

	SetQuietMode(true)
	output = captureError(func() {
		Error("test error")
	})
	if output == "" {
		t.Error("Expected error output even with quiet mode, got nothing")
	}
}
