package logging

// Store original functions for later restoration
var (
	originalPromptChoice = promptChoiceFunc
	originalPromptInput  = promptInputFunc
	originalConfirm      = confirmFunc
)

// MockPromptChoice replaces the PromptChoice function for testing
func MockPromptChoice(defaultChoice int) func() {
	// Replace the function
	promptChoiceFunc = func(prompt string, choices []string) (int, error) {
		return defaultChoice, nil
	}

	// Return a function to restore the original
	return func() {
		promptChoiceFunc = originalPromptChoice
	}
}

// MockPromptInput replaces the PromptInput function for testing
func MockPromptInput(returnValue string) func() {
	// Replace the function
	promptInputFunc = func(prompt string) string {
		return returnValue
	}

	// Return a function to restore the original
	return func() {
		promptInputFunc = originalPromptInput
	}
}

// MockConfirm replaces the Confirm function for testing
func MockConfirm(returnValue bool) func() {
	// Replace the function
	confirmFunc = func(prompt string) bool {
		return returnValue
	}

	// Return a function to restore the original
	return func() {
		confirmFunc = originalConfirm
	}
}

// DefaultMockSetup sets up all prompt functions with sensible defaults for testing
func DefaultMockSetup() func() {
	choiceRestore := MockPromptChoice(1) // Default to first choice
	inputRestore := MockPromptInput("test-input")
	confirmRestore := MockConfirm(true) // Default to confirming

	return func() {
		// Restore all original functions
		choiceRestore()
		inputRestore()
		confirmRestore()
	}
}

// DisableLoggingForTests temporarily disables all logging output for tests
func DisableLoggingForTests() func() {
	// Save original settings
	oldDebug := IsDebugEnabled
	oldQuiet := IsQuietEnabled

	// Disable all logging
	IsDebugEnabled = false
	IsQuietEnabled = true

	return func() {
		// Restore original settings
		IsDebugEnabled = oldDebug
		IsQuietEnabled = oldQuiet
	}
}
