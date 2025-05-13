package main

import (
	"fmt"
	"os"

	"simple-sops/internal/cli"
	"simple-sops/pkg/logging"
	"github.com/spf13/cobra"
)

var (
	debug bool
	quiet bool
)

func main() {
	// Initialize root command
	rootCmd := &cobra.Command{
		Use:   "simple-sops",
		Short: "Simple SOPS Helper - Making encryption easier",
		Long:  `A tool to simplify working with SOPS encryption and Age keys`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logging.SetDebugMode(debug)
			logging.SetQuietMode(quiet)
		},
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Minimal output")

	// Register all commands
	cli.RegisterCommands(rootCmd)

	// Special handling for no sub-commands (edit mode)
	if len(os.Args) > 1 && !isCommand(os.Args[1]) && !isFlag(os.Args[1]) {
		// Assume edit mode if first arg is a file
		if fileExists(os.Args[1]) {
			args := []string{"edit"}
			args = append(args, os.Args[1:]...)
			os.Args = append(os.Args[:1], args...)
		}
	}

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// isCommand checks if the argument is a defined command
func isCommand(arg string) bool {
	commands := []string{
		"encrypt", "decrypt", "edit", "set-keys", "config",
		"rm", "clean-config", "get-key", "clear-key", "help",
		"gen-key", "run", // New commands
	}
	for _, cmd := range commands {
		if arg == cmd {
			return true
		}
	}
	return false
}

// isFlag checks if the argument is a flag
func isFlag(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
