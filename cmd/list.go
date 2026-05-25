package cmd

import (
	"fmt"

	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
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
	cfg, exists, err := loadConfigReadOnly()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if !exists {
		printNotConfigured()
		return nil
	}

	// Print users
	ui.PrintUsersList(cfg.Users, cfg.ActiveUser)

	return nil
}
