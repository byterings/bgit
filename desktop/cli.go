package main

import (
	"os"
	"path/filepath"
	"strings"

	coresetup "github.com/byterings/bgit/core/setup"
)

// isCLIInvocation reports whether the process was started with real
// command-line arguments. macOS passes a "-psn_<id>" argument when an app is
// launched from Finder; that is not a CLI call.
func isCLIInvocation(args []string) bool {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-psn_") {
			return true
		}
	}
	return false
}

// hookHelperPath returns the path the managed pre-push hook should call when
// `bgit` is not on PATH: this desktop binary.
func hookHelperPath() string {
	// Inside an AppImage the executable lives in a temporary mount that
	// changes on every launch; APPIMAGE is the stable path of the image file.
	if appImage := strings.TrimSpace(os.Getenv("APPIMAGE")); appImage != "" {
		return appImage
	}

	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return exe
}

// configureHookHelper makes every setup path reached from this binary write a
// hook that can fall back to it.
func configureHookHelper() {
	coresetup.SetDefaultHookHelper(hookHelperPath())
}
