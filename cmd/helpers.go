package cmd

import (
	"fmt"
)

// autoInit initializes bgit automatically and runs first-time setup.
func autoInit() error {
	cfg, _, err := ensureConfigInitialized()
	if err != nil {
		return err
	}

	if !cfg.SetupCompleted {
		if err := applySetup(cfg, true); err != nil {
			return fmt.Errorf("automatic setup failed: %w\nRun: bgit setup", err)
		}
	}

	return nil
}
