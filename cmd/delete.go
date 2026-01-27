package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/ssh"
	"github.com/byterings/bgit/internal/ui"
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

	// Auto-initialize if needed
	if err := autoInit(); err != nil {
		return err
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find user
	user := cfg.FindUser(identifier)
	if user == nil {
		return fmt.Errorf("user '%s' not found", identifier)
	}

	// Confirm deletion
	confirmed, err := ui.PromptConfirmation(fmt.Sprintf("Delete user '%s' (%s)?", user.Alias, user.Email))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("Cancelled")
		return nil
	}

	// Ask if user wants to delete SSH keys too
	deleteKeys := false
	if user.SSHKeyPath != "" {
		deleteKeys, err = ui.PromptConfirmation(fmt.Sprintf("Also delete SSH key files (%s)?", user.SSHKeyPath))
		if err != nil {
			return err
		}
	}

	// Remove user from list
	newUsers := []config.User{}
	for _, u := range cfg.Users {
		if u.Alias != user.Alias {
			newUsers = append(newUsers, u)
		}
	}
	cfg.Users = newUsers

	// Clear active user if it was the deleted one
	if cfg.ActiveUser == user.Alias {
		cfg.ActiveUser = ""
		ui.Info("Active user cleared")
	}

	// Delete SSH keys if requested
	if deleteKeys && user.SSHKeyPath != "" {
		// Delete private key
		if err := os.Remove(user.SSHKeyPath); err != nil {
			ui.Info(fmt.Sprintf("Warning: Could not delete private key: %v", err))
		} else {
			ui.Success(fmt.Sprintf("Deleted: %s", user.SSHKeyPath))
		}

		// Delete public key
		pubKeyPath := user.SSHKeyPath + ".pub"
		if err := os.Remove(pubKeyPath); err != nil {
			ui.Info(fmt.Sprintf("Warning: Could not delete public key: %v", err))
		} else {
			ui.Success(fmt.Sprintf("Deleted: %s", pubKeyPath))
		}
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Update SSH config
	if err := ssh.UpdateSSHConfig(cfg.Users); err != nil {
		ui.Info("Warning: Failed to update SSH config")
	}

	ui.Success(fmt.Sprintf("User '%s' deleted", user.Alias))

	if len(cfg.Users) == 0 {
		fmt.Println("\nNo users remaining. Add one with: bgit add")
	}

	return nil
}
