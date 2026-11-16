//go:build windows

package main

import (
	"os"
	"syscall"
)

// attachParentConsole connects the windowed desktop binary to the console it
// was started from, so CLI output is visible in cmd.exe and PowerShell.
//
// When Git runs the pre-push hook it already hands the process valid output
// handles (pipes); in that case nothing needs to change.
func attachParentConsole() {
	if handle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE); err == nil && handle != 0 && handle != syscall.InvalidHandle {
		return
	}

	// ATTACH_PARENT_PROCESS is (DWORD)-1.
	const attachParentProcess = ^uint32(0)
	attachConsole := syscall.NewLazyDLL("kernel32.dll").NewProc("AttachConsole")
	if result, _, _ := attachConsole.Call(uintptr(attachParentProcess)); result == 0 {
		return
	}

	if out, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
		os.Stdout = out
		os.Stderr = out
	}
	if in, err := os.OpenFile("CONIN$", os.O_RDONLY, 0); err == nil {
		os.Stdin = in
	}
}
