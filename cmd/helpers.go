package cmd

import (
	"fmt"

	"github.com/byterings/bgit/core/config"
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

func loadConfigReadOnly() (*config.Config, bool, error) {
	exists, err := config.ConfigExists()
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, true, err
	}

	return cfg, true, nil
}

func printNotConfigured() {
	fmt.Println("bgit is not configured.")
	fmt.Println("Run: bgit setup")
}
