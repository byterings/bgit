//go:build !windows

package platform

import "os/exec"

func hideConsoleWindow(cmd *exec.Cmd) {}
