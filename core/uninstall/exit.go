package uninstall

import (
	"errors"
	"os/exec"
)

// exitCode returns the exit code of a failed command, or -1 when the command
// did not run to completion.
func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
