package keymgmt

import (
	"os/exec"
)

// Variables that can be swapped for testing
var (
	// Use a variable for exec.Command to allow mocking in tests
	execCommand = exec.Command

	// Use a variable for exec.LookPath to allow mocking in tests
	lookPathFunc = exec.LookPath
)
