package cmd

import (
	"fmt"

	"github.com/byterings/bgit/core/config"
	coreidentity "github.com/byterings/bgit/core/identity"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
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

	// Validate SSH key path
	if err := coressh.ValidateSSHKeyPath(updateSSHKey); err != nil {
		return err
	}

	foundUser, err := coreidentity.UpdateUserSSHKey(cfg, identifier, updateSSHKey)
	if err != nil {
		return fmt.Errorf("%w\nRun: bgit list", err)
	}

	ui.Success(fmt.Sprintf("SSH key updated for '%s'", foundUser.Alias))

	// Show public key to add to GitHub
	pubKeyContent, err := coressh.GetPublicKeyContent(updateSSHKey)
	if err == nil {
		fmt.Println("\nAdd this public key to your GitHub account:")
		fmt.Println("https://github.com/settings/keys")
		fmt.Println("---")
		fmt.Print(pubKeyContent)
		fmt.Println("---")
	}

	return nil
}
