package logging

import (
	"fmt"
	"os"
)

var (
	debugMode bool
	quietMode bool
)

// SetDebugMode enables or disables debug logging
func SetDebugMode(debug bool) {
	debugMode = debug
}

// SetQuietMode enables or disables minimal output
func SetQuietMode(quiet bool) {
	quietMode = quiet
}

// Debug logs a debug message (only if debug mode is enabled)
func Debug(format string, args ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stdout, "[DEBUG] "+format+"\n", args...)
	}
}

// Info logs an informational message (unless quiet mode is enabled)
func Info(format string, args ...interface{}) {
	if !quietMode {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

// Success logs a success message (unless quiet mode is enabled)
func Success(format string, args ...interface{}) {
	if !quietMode {
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

// Confirm prompts the user for confirmation
func Confirm(prompt string) bool {
	var response string
	fmt.Printf("%s [y/N]: ", prompt)
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
}

// PromptChoice prompts the user for a numbered choice
func PromptChoice(prompt string, choices []string) (int, error) {
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

// PromptInput prompts the user for input
func PromptInput(prompt string) string {
	var response string
	fmt.Print(prompt + ": ")
	fmt.Scanln(&response)
	return response
}
