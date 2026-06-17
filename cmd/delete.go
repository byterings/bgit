package cmd

import (
	"fmt"
	"os"

	"github.com/byterings/bgit/core/config"
	coreidentity "github.com/byterings/bgit/core/identity"
	"github.com/byterings/bgit/internal/git"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <alias>",
	Short: "Delete a user identity",
	Long:  `Remove a user identity from bgit configuration and optionally delete SSH keys.`,
	Args:  cobra.ExactArgs(1),
	Example: `  bgit delete work
  bgit delete personal`,
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	if err := autoInit(); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	user := cfg.FindUser(identifier)
	if user == nil {
		return fmt.Errorf("user '%s' not found", identifier)
	}

	bindingCount := 0
	for _, binding := range cfg.GetBindings() {
		if binding.User == user.Alias {
			bindingCount++
		}
	}
	workspaceCount := 0
	for _, workspace := range cfg.GetWorkspaces() {
		if workspace.User == user.Alias {
			workspaceCount++
		}
	}

	prompt := fmt.Sprintf("Delete user '%s' (%s)?", user.Alias, user.Email)
	if bindingCount > 0 || workspaceCount > 0 {
		prompt = fmt.Sprintf(
			"Delete user '%s' (%s)? This will also remove %d binding(s) and %d workspace binding(s).",
			user.Alias,
			user.Email,
			bindingCount,
			workspaceCount,
		)
	}

	confirmed, err := ui.PromptConfirmation(prompt)
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("Operation cancelled.")
		return nil
	}

	deleteKeys := false
	if user.SSHKeyPath != "" {
		deleteKeys, err = ui.PromptConfirmation(fmt.Sprintf("Also delete SSH key files (%s)?", user.SSHKeyPath))
		if err != nil {
			return err
		}
	}

	result, err := coreidentity.DeleteUser(cfg, identifier)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.ActiveCleared {
		ui.Info("Active user cleared")
	}
	for _, repoPath := range result.RemovedBindings {
		if err := git.UnsetLocalUser(repoPath); err != nil {
			ui.Warning(fmt.Sprintf("Could not clear repository Git identity for %s: %v", repoPath, err))
		}
	}
	if len(result.RemovedBindings) > 0 {
		ui.Warning(fmt.Sprintf("Removed %d binding(s) for '%s'", len(result.RemovedBindings), user.Alias))
	}
	if len(result.RemovedWorkspaces) > 0 {
		ui.Warning(fmt.Sprintf("Removed %d workspace binding(s) for '%s'", len(result.RemovedWorkspaces), user.Alias))
	}

	if deleteKeys && user.SSHKeyPath != "" {
		if err := os.Remove(user.SSHKeyPath); err != nil {
			ui.Warning(fmt.Sprintf("Could not delete private key: %v", err))
		} else {
			ui.Success(fmt.Sprintf("Deleted: %s", user.SSHKeyPath))
		}

		pubKeyPath := user.SSHKeyPath + ".pub"
		if err := os.Remove(pubKeyPath); err != nil {
			ui.Warning(fmt.Sprintf("Could not delete public key: %v", err))
		} else {
			ui.Success(fmt.Sprintf("Deleted: %s", pubKeyPath))
		}
	}

	ui.Success(fmt.Sprintf("User '%s' deleted", user.Alias))

	if len(cfg.Users) == 0 {
		fmt.Println("\nNo users remaining. Add one with: bgit add")
	}

	return nil
}
