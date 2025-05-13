package encrypt

import (
	"os/exec"
)

// Use a variable for exec.Command to allow mocking in tests
var execCommand = exec.Command
