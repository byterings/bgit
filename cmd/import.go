package cmd

import (
	"fmt"

	coreexport "github.com/byterings/bgit/core/export"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <archive.bgit>",
	Short: "Restore bgit configuration from an encrypted archive",
	Args:  cobra.ExactArgs(1),
	Long: `Restore bgit configuration from an encrypted .bgit archive.

This command reads the encrypted export envelope, prompts for the archive
password, decrypts the preserved backup payload, validates the config, and
restores it to the managed bgit config path.`,
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	password, err := ui.PromptPassword("Import password:")
	if err != nil {
		return fmt.Errorf("failed to read import password: %w", err)
	}

	result, err := coreexport.ImportArchive(args[0], password)
	if err != nil {
		return fmt.Errorf("failed to import archive: %w", err)
	}

	ui.Success("Imported bgit archive")
	ui.Info(fmt.Sprintf("Users restored: %d", result.UsersCount))
	if result.ActiveUser != "" {
		ui.Info(fmt.Sprintf("Active user: %s", result.ActiveUser))
	}
	return nil
}
