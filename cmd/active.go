package cmd

import (
	"fmt"

	coreidentity "github.com/byterings/bgit/core/identity"
	"github.com/spf13/cobra"
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Show the currently active user",
	Long: `Display which user identity is currently active.

Shows the effective identity for the current directory, which may differ
from the global active user if you're inside a workspace or bound repository.`,
	RunE: runActive,
}

func init() {
	rootCmd.AddCommand(activeCmd)
}

func runActive(cmd *cobra.Command, args []string) error {
	cfg, exists, err := loadConfigReadOnly()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if !exists {
		printNotConfigured()
		return nil
	}

	// Get effective identity (respects workspace/binding)
	resolution, err := coreidentity.GetEffectiveResolution(cfg)
	if err != nil {
		return fmt.Errorf("failed to resolve identity: %w", err)
	}

	if resolution == nil {
		fmt.Println("No active user set")
		fmt.Println("\nSet one with: bgit use <alias>")
		return nil
	}

	activeUser := resolution.User

	// Show source of identity
	sourceInfo := ""
	switch resolution.Source {
	case coreidentity.SourceWorkspace:
		sourceInfo = fmt.Sprintf(" (workspace: %s)", resolution.Path)
	case coreidentity.SourceBinding:
		sourceInfo = " (bound repo)"
	case coreidentity.SourceGlobal:
		sourceInfo = " (global)"
	}

	fmt.Printf("Active user: %s%s\n", resolution.Alias, sourceInfo)
	fmt.Printf("  Name: %s\n", activeUser.Name)
	fmt.Printf("  Email: %s\n", activeUser.Email)
	fmt.Printf("  GitHub: %s\n", activeUser.GitHubUsername)
	if activeUser.SSHKeyPath != "" {
		fmt.Printf("  SSH Key: %s\n", activeUser.SSHKeyPath)
	}

	return nil
}
