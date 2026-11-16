package platform

import (
	"os/exec"
	"sync/atomic"
)

// backgroundMode is enabled by the desktop app. A windowed process has no
// console, so on Windows every child process would otherwise flash its own
// console window. The CLI never enables it, so CLI child processes are
// unaffected.
var backgroundMode atomic.Bool

// SetBackgroundMode marks the current process as a windowed (non-console) app.
func SetBackgroundMode(enabled bool) {
	backgroundMode.Store(enabled)
}

// PrepareCommand applies platform process settings for the current mode and
// returns the same command for chaining.
func PrepareCommand(cmd *exec.Cmd) *exec.Cmd {
	if backgroundMode.Load() {
		hideConsoleWindow(cmd)
	}
	return cmd
}
