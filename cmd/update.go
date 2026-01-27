package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/ssh"
	"github.com/byterings/bgit/internal/ui"
	"github.com/byterings/bgit/internal/user"
)

var (
	updateSSHKey string
)

var updateCmd = &cobra.Command{
	Use:   "update <alias>",
	Short: "Update a user's SSH key",
	Long:  `Update the SSH key for an existing user.`,
	Args:  cobra.ExactArgs(1),
	Example: `  bgit update work --ssh-key ~/.ssh/id_ed25519
  bgit update personal --ssh-key ~/.ssh/bgit_personal`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateSSHKey, "ssh-key", "", "Path to SSH private key")
	updateCmd.MarkFlagRequired("ssh-key")
}

func runUpdate(cmd *cobra.Command, args []string) error {
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
	foundUser := cfg.FindUser(identifier)
	if foundUser == nil {
		return fmt.Errorf("user '%s' not found\nRun: bgit list", identifier)
	}

	// Validate SSH key path
	if err := user.ValidateSSHKeyPath(updateSSHKey); err != nil {
		return err
	}

	// Update user's SSH key
	for i := range cfg.Users {
		if cfg.Users[i].Alias == foundUser.Alias {
			cfg.Users[i].SSHKeyPath = updateSSHKey
			break
		}
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Update SSH config
	if err := ssh.UpdateSSHConfig(cfg.Users); err != nil {
		return fmt.Errorf("failed to update SSH config: %w", err)
	}

	ui.Success(fmt.Sprintf("SSH key updated for '%s'", foundUser.Alias))

	// Show public key to add to GitHub
	pubKeyContent, err := user.GetPublicKeyContent(updateSSHKey)
	if err == nil {
		fmt.Println("\nAdd this public key to your GitHub account:")
		fmt.Println("https://github.com/settings/keys")
		fmt.Println("---")
		fmt.Print(pubKeyContent)
		fmt.Println("---")
	}

	return nil
}
