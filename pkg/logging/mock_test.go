package logging

import (
	"testing"
)

func TestMocking(t *testing.T) {
	// Test mocking PromptChoice
	restoreChoice := MockPromptChoice(42)

	choice, err := PromptChoice("This should be mocked", []string{"Option 1", "Option 2"})
	if err != nil {
		t.Errorf("MockPromptChoice returned an error: %v", err)
	}
	if choice != 42 {
		t.Errorf("MockPromptChoice failed: expected 42, got %d", choice)
	}

	// Restore original
	restoreChoice()

	// Test mocking PromptInput
	restoreInput := MockPromptInput("mocked input")

	input := PromptInput("This should be mocked")
	if input != "mocked input" {
		t.Errorf("MockPromptInput failed: expected 'mocked input', got '%s'", input)
	}

	// Restore original
	restoreInput()

	// Test mocking Confirm
	restoreConfirm := MockConfirm(false)

	confirmed := Confirm("This should be mocked")
	if confirmed != false {
		t.Errorf("MockConfirm failed: expected false, got %v", confirmed)
	}

	// Restore original
	restoreConfirm()

	// Test DefaultMockSetup
	restoreAll := DefaultMockSetup()

	choice, _ = PromptChoice("Default mock", []string{"Option"})
	if choice != 1 {
		t.Errorf("DefaultMockSetup for PromptChoice failed: expected 1, got %d", choice)
	}

	input = PromptInput("Default mock")
	if input != "test-input" {
		t.Errorf("DefaultMockSetup for PromptInput failed: expected 'test-input', got '%s'", input)
	}

	confirmed = Confirm("Default mock")
	if !confirmed {
		t.Errorf("DefaultMockSetup for Confirm failed: expected true, got false")
	}

	// Restore all originals
	restoreAll()
}

func TestDisableLoggingForTests(t *testing.T) {
	// Enable debug and disable quiet initially
	SetDebugMode(true)
	SetQuietMode(false)

	// Now disable logging
	restore := DisableLoggingForTests()

	if IsDebugEnabled {
		t.Error("DisableLoggingForTests failed: Debug still enabled")
	}

	if !IsQuietEnabled {
		t.Error("DisableLoggingForTests failed: Quiet not enabled")
	}

	// Restore original settings
	restore()

	if !IsDebugEnabled {
		t.Error("DisableLoggingForTests restore failed: Debug not enabled")
	}

	if IsQuietEnabled {
		t.Error("DisableLoggingForTests restore failed: Quiet still enabled")
	}
}
