//go:build !devharness

package main

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	// The default window needs about this much room, including the title bar
	// and the desktop's own panels.
	comfortableScreenWidth  = 1300
	comfortableScreenHeight = 860
)

// fitWindowToScreen maximises the window on small screens. On a 1366x768
// laptop the default window would be taller than the screen and its bottom
// edge, with buttons on it, would be out of reach.
func fitWindowToScreen(ctx context.Context) {
	screens, err := wailsruntime.ScreenGetAll(ctx)
	if err != nil {
		return
	}
	for _, screen := range screens {
		if !screen.IsCurrent {
			continue
		}
		if screen.Size.Width < comfortableScreenWidth || screen.Size.Height < comfortableScreenHeight {
			wailsruntime.WindowMaximise(ctx)
		}
		return
	}
}
