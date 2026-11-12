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
)

//go:embed all:frontend/dist
var assets embed.FS

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
		Height:             760,
		MinWidth:           860,
		MinHeight:          560,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
