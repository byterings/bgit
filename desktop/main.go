//go:build !devharness

package main

import (
	"embed"
	"os"
	"strings"

	"github.com/byterings/bgit/cmd"
	"github.com/byterings/bgit/internal/platform"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	configureHookHelper()

	// With arguments the desktop binary behaves as the bgit CLI. The managed
	// pre-push hook relies on this (`<app> check --from-hook`) so push safety
	// works for users who never installed the CLI.
	if isCLIInvocation(os.Args[1:]) {
		attachParentConsole()
		cmd.Execute()
		return
	}

	platform.SetBackgroundMode(true)
	app := NewApp()

	// BGIT_DESKTOP_LOG=trace prints every call between the window and the
	// backend to standard output, for diagnosing a window that does not load.
	logLevel := logger.ERROR
	if strings.EqualFold(os.Getenv("BGIT_DESKTOP_LOG"), "trace") {
		logLevel = logger.TRACE
	}

	err := wails.Run(&options.App{
		LogLevelProduction: logLevel,
		Title:              "bgit",
		Width:              1180,
		Height:             720,
		MinWidth:           900,
		MinHeight:          560,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// The dark theme's background, so the window never flashes white
		// before the page has loaded.
		BackgroundColour: &options.RGBA{R: 11, G: 15, B: 25, A: 255},
		Linux: &linux.Options{
			Icon: appIcon,
			// Matches the installed desktop entry, so the desktop groups the
			// window under the bgit launcher icon.
			ProgramName:      "bgit-desktop",
			WebviewGpuPolicy: linuxGpuPolicy(),
		},
		OnStartup:  app.startup,
		OnDomReady: app.domReady,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}

// linuxGpuPolicy chooses how the Linux web view draws. Software drawing is
// the default because GPU drawing leaves the window blank on some graphics
// drivers. BGIT_DESKTOP_GPU=on lets the web view use the GPU when a page
// asks for it, which can make scrolling smoother where it works.
func linuxGpuPolicy() linux.WebviewGpuPolicy {
	if strings.EqualFold(os.Getenv("BGIT_DESKTOP_GPU"), "on") {
		return linux.WebviewGpuPolicyOnDemand
	}
	return linux.WebviewGpuPolicyNever
}
