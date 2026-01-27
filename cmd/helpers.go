package cmd

import (
	"github.com/byterings/bgit/internal/config"
)

// autoInit initializes bgit automatically if not already initialized
func autoInit() error {
	exists, err := config.ConfigExists()
	if err != nil {
		return err
	}

	if !exists {
		// Silently initialize
		if err := config.CreateConfigDir(); err != nil {
			return err
		}

		if err := config.CreateBackupDir(); err != nil {
			return err
		}

		cfg := config.NewConfig()
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}
	}

	return nil
}
