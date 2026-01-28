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

	confirmed, err := ui.PromptConfirmation(fmt.Sprintf("Delete user '%s' (%s)?", user.Alias, user.Email))
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

	newUsers := []config.User{}
	for _, u := range cfg.Users {
		if u.Alias != user.Alias {
			newUsers = append(newUsers, u)
		}
	}
	cfg.Users = newUsers

	if cfg.ActiveUser == user.Alias {
		cfg.ActiveUser = ""
		ui.Info("Active user cleared")
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

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if err := ssh.UpdateSSHConfig(cfg.Users); err != nil {
		ui.Info("Warning: Failed to update SSH config")
	}

	ui.Success(fmt.Sprintf("User '%s' deleted", user.Alias))

	if len(cfg.Users) == 0 {
		fmt.Println("\nNo users remaining. Add one with: bgit add")
	}

	return nil
}
