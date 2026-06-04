package cmd

import (
	"fmt"

	coreexport "github.com/byterings/bgit/core/export"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Create a bgit backup archive",
	Long: `Create an encrypted .bgit archive containing the current bgit backup payload structure.

This command preserves the existing backup archive layout and wraps it in the
encrypted export envelope used by bgit.`,
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	cfg, exists, err := loadConfigReadOnly()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if !exists {
		printNotConfigured()
		return nil
	}

	ui.Info("This password is required to import this backup later. If you forget it, the archive cannot be restored.")
	password, err := ui.PromptPasswordConfirmation("Export password:", "Confirm export password:")
	if err != nil {
		return fmt.Errorf("failed to read export password: %w", err)
	}

	result, err := coreexport.CreateArchive(cfg, version, password)
	if err != nil {
		return fmt.Errorf("failed to create export archive: %w", err)
	}

	ui.Success("Created bgit export archive")
	ui.Info(result.Path)
	return nil
}
