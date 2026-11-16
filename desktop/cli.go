package main

import (
	"os"
	"path/filepath"
	"strings"

	coresetup "github.com/byterings/bgit/core/setup"
)

// isCLIInvocation reports whether the process was started as the bgit CLI,
// for example `<app> check --from-hook` from the managed pre-push hook.
//
// Only a leading subcommand counts. Arguments that start with "-" are left
// to the window: macOS passes "-psn_<id>" when an app is launched from
// Finder, and Wails passes its own flags in dev mode and when it generates
// frontend bindings.
func isCLIInvocation(args []string) bool {
	return len(args) > 0 && !strings.HasPrefix(args[0], "-")
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

	// macOS runs an app opened straight from a download out of a random
	// read-only location. That path stops existing when the app quits, so
	// the installed copy is used when there is one.
	if strings.Contains(exe, "/AppTranslocation/") {
		if _, bundlePath, found := strings.Cut(exe, ".app/"); found {
			installed := filepath.Join("/Applications", filepath.Base(strings.TrimSuffix(exe, "/"+bundlePath)), bundlePath)
			if info, err := os.Stat(installed); err == nil && !info.IsDir() {
				return installed
			}
		}
	}
	return exe
}

// configureHookHelper makes every setup path reached from this binary write a
// hook that can fall back to it.
func configureHookHelper() {
	coresetup.SetDefaultHookHelper(hookHelperPath())
}
