package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/git"
	"github.com/byterings/bgit/internal/ssh"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var (
	useByUsername bool
	useByEmail    bool
)

var useCmd = &cobra.Command{
	Use:   "use <alias>",
	Short: "Switch to a different Git identity",
	Long:  `Switch to a different Git identity by alias, username, or email.`,
	Args:  cobra.ExactArgs(1),
	Example: `  bgit use work              # By alias (default)
  bgit use -u john-work      # By GitHub username
  bgit use -m john@work.com  # By email`,
	RunE: runUse,
}

func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.Flags().BoolVarP(&useByUsername, "username", "u", false, "Find user by GitHub username")
	useCmd.Flags().BoolVarP(&useByEmail, "email", "m", false, "Find user by email")
}

func runUse(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	// Check if git is installed
	if !git.IsGitInstalled() {
		return fmt.Errorf("git is not installed")
	}

	// Auto-initialize if needed
	if err := autoInit(); err != nil {
		return err
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find user based on flags
	var user *config.User
	if useByUsername {
		user = cfg.FindUserByUsername(identifier)
	} else if useByEmail {
		user = cfg.FindUserByEmail(identifier)
	} else {
		// Default: find by alias, username, or email
		user = cfg.FindUser(identifier)
	}

	if user == nil {
		return fmt.Errorf("user '%s' not found\nRun: bgit list", identifier)
	}

	fmt.Printf("Switching to: %s (%s)\n", user.Alias, user.Email)

	// Update Git global config
	if err := git.SetGlobalUser(user.Name, user.Email); err != nil {
		return fmt.Errorf("failed to update git config: %w", err)
	}

	// Update SSH config
	if err := ssh.UpdateSSHConfig(cfg.Users); err != nil {
		return fmt.Errorf("failed to update SSH config: %w", err)
	}

	// Update active user in bgit config (store alias)
	cfg.ActiveUser = user.Alias
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Auto-setup SSH agent on Windows
	if user.SSHKeyPath != "" {
		ensureSSHAgent(user)
	}

	ui.Success("Identity switched successfully")

	if user.SSHKeyPath != "" {
		fmt.Println("\nClone repos: bgit clone <url>")
		fmt.Println("Fix existing: bgit remote fix")
	}

	return nil
}

// ensureSSHAgent checks if SSH agent is running and adds the user's key
// This runs silently - only shows messages if there's an issue
func ensureSSHAgent(user *config.User) {
	if runtime.GOOS == "windows" {
		// Start ssh-agent service silently
		startCmd := exec.Command("powershell", "-Command", "Start-Service ssh-agent")
		startCmd.Run() // Ignore errors - may already be running

		// Set to automatic startup
		autoCmd := exec.Command("powershell", "-Command", "Set-Service -Name ssh-agent -StartupType Automatic")
		autoCmd.Run() // Ignore errors - may require admin
	}

	// Check if key is already loaded
	listCmd := exec.Command("ssh-add", "-l")
	output, _ := listCmd.Output()

	// If key not in agent, add it
	if user.SSHKeyPath != "" && !strings.Contains(string(output), user.SSHKeyPath) {
		addCmd := exec.Command("ssh-add", user.SSHKeyPath)
		if err := addCmd.Run(); err == nil {
			ui.Info("SSH key loaded into agent")
		}
	}
}
