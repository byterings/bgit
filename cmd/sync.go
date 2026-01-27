package cmd

import (
	"fmt"
	"os"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/git"
	"github.com/byterings/bgit/internal/identity"
	"github.com/byterings/bgit/internal/platform"
	"github.com/byterings/bgit/internal/ssh"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var (
	autoFix bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Validate and sync bgit configuration",
	Long: `Check if Git and SSH configurations match the effective bgit user.

The effective user is determined by:
1. Workspace (if inside a workspace folder)
2. Binding (if repo is bound to a user)
3. Global active user (fallback)

Optionally fix any mismatches found.`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVarP(&autoFix, "fix", "f", false, "Automatically fix issues without prompting")
}

func runSync(cmd *cobra.Command, args []string) error {
	// Check if bgit is initialized
	exists, err := config.ConfigExists()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("bgit not initialized. Run 'bgit init' first")
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get effective identity (respects workspace/binding)
	resolution, err := identity.GetEffectiveResolution(cfg)
	if err != nil {
		return fmt.Errorf("failed to resolve identity: %w", err)
	}

	if resolution == nil || resolution.User == nil {
		ui.Info("No active user set")
		fmt.Println("Set one with: bgit use <alias>")
		return nil
	}

	activeUser := resolution.User

	// Show context info
	sourceInfo := ""
	switch resolution.Source {
	case identity.SourceWorkspace:
		sourceInfo = fmt.Sprintf(" (workspace: %s)", resolution.Path)
	case identity.SourceBinding:
		sourceInfo = " (bound repo)"
	case identity.SourceGlobal:
		sourceInfo = ""
	}

	fmt.Printf("Validating identity: %s%s\n", resolution.Alias, sourceInfo)
	fmt.Printf("Checking configuration for: %s (%s)\n\n", activeUser.GitHubUsername, activeUser.Email)

	issues := []string{}

	// Check Git config
	fmt.Println("Checking Git config...")
	gitName, gitEmail, err := git.GetGlobalUser()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get Git config: %v", err))
		issues = append(issues, "git_config_error")
	} else {
		if gitName != activeUser.Name {
			ui.Error(fmt.Sprintf("Git user.name mismatch: got '%s', expected '%s'", gitName, activeUser.Name))
			issues = append(issues, "git_name_mismatch")
		} else {
			ui.Success("Git user.name matches")
		}

		if gitEmail != activeUser.Email {
			ui.Error(fmt.Sprintf("Git user.email mismatch: got '%s', expected '%s'", gitEmail, activeUser.Email))
			issues = append(issues, "git_email_mismatch")
		} else {
			ui.Success("Git user.email matches")
		}
	}

	// Check SSH key
	if activeUser.SSHKeyPath != "" {
		fmt.Println("\nChecking SSH key...")
		if _, err := os.Stat(activeUser.SSHKeyPath); os.IsNotExist(err) {
			ui.Error(fmt.Sprintf("SSH key not found: %s", activeUser.SSHKeyPath))
			issues = append(issues, "ssh_key_missing")
		} else {
			ui.Success("SSH key exists")

			// Check permissions (Unix only)
			ok, err := platform.CheckFilePermissions(activeUser.SSHKeyPath)
			if err == nil && !ok {
				info, _ := os.Stat(activeUser.SSHKeyPath)
				mode := info.Mode()
				ui.Error(fmt.Sprintf("SSH key has insecure permissions: %s", mode))
				issues = append(issues, "ssh_key_permissions")
			} else if err == nil {
				ui.Success("SSH key permissions OK")
			}
		}

		// Check public key
		pubKeyPath := activeUser.SSHKeyPath + ".pub"
		if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
			ui.Error(fmt.Sprintf("SSH public key not found: %s", pubKeyPath))
			issues = append(issues, "ssh_pubkey_missing")
		} else {
			ui.Success("SSH public key exists")
		}
	}

	fmt.Println()

	if len(issues) == 0 {
		ui.Success("All checks passed! Configuration is in sync.")
		return nil
	}

	// Issues found
	fmt.Printf("\033[31mFound %d issue(s)\033[0m\n\n", len(issues))

	// Determine if we should fix
	fix := autoFix
	if !autoFix {
		// Ask if user wants to fix
		prompted, err := ui.PromptConfirmation("Fix these issues automatically?")
		if err != nil {
			return err
		}
		fix = prompted
	}

	if !fix {
		fmt.Println("\nNo changes made. Run 'bgit sync --fix' to auto-fix.")
		return nil
	}

	// Apply fixes
	fmt.Println("\nApplying fixes...")

	for _, issue := range issues {
		switch issue {
		case "git_name_mismatch", "git_email_mismatch", "git_config_error":
			if err := git.SetGlobalUser(activeUser.Name, activeUser.Email); err != nil {
				ui.Error(fmt.Sprintf("Failed to fix Git config: %v", err))
			} else {
				ui.Success("Fixed Git config")
			}

		case "ssh_key_permissions":
			if err := platform.FixFilePermissions(activeUser.SSHKeyPath); err != nil {
				ui.Error(fmt.Sprintf("Failed to fix SSH key permissions: %v", err))
			} else {
				ui.Success("Fixed SSH key permissions")
			}
		}
	}

	// Update SSH config
	if err := ssh.UpdateSSHConfig(cfg.Users); err != nil {
		ui.Error(fmt.Sprintf("Failed to update SSH config: %v", err))
	} else {
		ui.Success("Updated SSH config")
	}

	fmt.Println()
	ui.Success("Sync complete!")

	return nil
}
