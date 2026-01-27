package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/ui"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all configured user identities",
	Long:    `Display all configured Git user identities and highlight the active one.`,
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Auto-initialize if needed
	if err := autoInit(); err != nil {
		return err
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Print users
	ui.PrintUsersList(cfg.Users, cfg.ActiveUser)

	return nil
}
