package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// IsDebugEnabled controls debug logging (exported for tests)
	IsDebugEnabled bool
	// IsQuietEnabled controls minimal output (exported for tests)
	IsQuietEnabled bool

	// Function variables that can be swapped for testing
	promptChoiceFunc = defaultPromptChoice
	promptInputFunc  = defaultPromptInput
	confirmFunc      = defaultConfirm
)

// SetDebugMode enables or disables debug logging
func SetDebugMode(debug bool) {
	IsDebugEnabled = debug
}

// SetQuietMode enables or disables minimal output
func SetQuietMode(quiet bool) {
	IsQuietEnabled = quiet
}

// Debug logs a debug message (only if debug mode is enabled)
func Debug(format string, args ...interface{}) {
	if IsDebugEnabled {
		fmt.Fprintf(os.Stdout, "[DEBUG] "+format+"\n", args...)
	}
}

// Info logs an informational message (unless quiet mode is enabled)
func Info(format string, args ...interface{}) {
	if !IsQuietEnabled {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

// Success logs a success message (unless quiet mode is enabled)
func Success(format string, args ...interface{}) {
	if !IsQuietEnabled {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

// Error logs an error message (always shown)
func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// Fatal logs an error message and exits
func Fatal(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}

// isTestMode checks if we're running in test mode
func isTestMode() bool {
	testMode := os.Getenv("TEST_MODE")
	return testMode == "1" || testMode == "true" || Testing()
}

// Testing checks if the code is running as part of a Go test
func Testing() bool {
	// Get the call stack
	pc := make([]uintptr, 10)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])

	// Look for test patterns in the call stack
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		// Check if the file path contains "_test.go"
		if strings.Contains(filepath.Base(frame.File), "_test.go") {
			return true
		}

		// Check if the function name contains "Test"
		if strings.Contains(frame.Function, ".Test") {
			return true
		}
	}

	return false
}

// Default implementation of PromptChoice
func defaultPromptChoice(prompt string, choices []string) (int, error) {
	// In test mode, avoid actual prompts
	if isTestMode() {
		// Default to first choice in test mode
		return 1, nil
	}

	fmt.Println(prompt)
	for i, choice := range choices {
		fmt.Printf("%d. %s\n", i+1, choice)
	}
	var response int
	fmt.Print("Choose option: ")
	_, err := fmt.Scanln(&response)
	if err != nil {
		return 0, err
	}
	if response < 1 || response > len(choices) {
		return 0, fmt.Errorf("invalid choice: %d", response)
	}
	return response, nil
}

// Default implementation of PromptInput
func defaultPromptInput(prompt string) string {
	// In test mode, avoid actual prompts
	if isTestMode() {
		// Return empty string in test mode
		return ""
	}

	var response string
	fmt.Print(prompt + ": ")
	fmt.Scanln(&response)
	return response
}

// Default implementation of Confirm
func defaultConfirm(prompt string) bool {
	// In test mode, avoid actual prompts
	if isTestMode() {
		// Default to confirming in test mode
		return true
	}

	var response string
	fmt.Printf("%s [y/N]: ", prompt)
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
}

// Public functions that use the swappable implementations

// PromptChoice prompts the user for a numbered choice
func PromptChoice(prompt string, choices []string) (int, error) {
	return promptChoiceFunc(prompt, choices)
}

// PromptInput prompts the user for input
func PromptInput(prompt string) string {
	return promptInputFunc(prompt)
}

// Confirm prompts the user for confirmation
func Confirm(prompt string) bool {
	return confirmFunc(prompt)
}
