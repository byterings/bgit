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
	Long: `Create a .bgit archive containing the current bgit backup payload structure.

This command only builds the archive structure and manifest. Encryption is added
later as part of the export roadmap.`,
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

	result, err := coreexport.CreateArchive(cfg, version)
	if err != nil {
		return fmt.Errorf("failed to create export archive: %w", err)
	}

	ui.Success("Created bgit export archive")
	ui.Info(result.Path)
	return nil
}
