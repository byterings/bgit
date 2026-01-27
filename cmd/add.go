package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/platform"
	"github.com/byterings/bgit/internal/ui"
	"github.com/byterings/bgit/internal/user"
)

var (
	addFlagAlias   string
	addFlagName    string
	addFlagEmail   string
	addFlagGitHub  string
	addFlagSSHKey  string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new Git user identity",
	Long:  `Add a new Git user identity with name, email, and SSH key.`,
	Example: `  # Interactive mode
  bgit add

  # Using flags
  bgit add --name "John Doe" --email "john@work.com" --github "john-work"`,
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVar(&addFlagAlias, "alias", "", "Alias for this identity (e.g., work, personal, freelance)")
	addCmd.Flags().StringVar(&addFlagName, "name", "", "Full name for Git commits")
	addCmd.Flags().StringVar(&addFlagEmail, "email", "", "Email address for Git commits")
	addCmd.Flags().StringVar(&addFlagGitHub, "github", "", "GitHub username")
	addCmd.Flags().StringVar(&addFlagSSHKey, "ssh-key", "", "Path to existing SSH private key")
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Auto-initialize if needed
	if err := autoInit(); err != nil {
		return err
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var alias, name, email, githubUsername, sshKeyPath string

	// Get user info (interactive or from flags)
	if addFlagAlias == "" || addFlagName == "" || addFlagEmail == "" || addFlagGitHub == "" {
		// Interactive mode
		fmt.Println("Adding new user identity")
		fmt.Println()

		alias, name, email, githubUsername, err = ui.PromptUserInfo()
		if err != nil {
			return fmt.Errorf("failed to get user info: %w", err)
		}
	} else {
		// Flag mode
		alias = addFlagAlias
		name = addFlagName
		email = addFlagEmail
		githubUsername = addFlagGitHub
	}

	// Handle SSH key
	if addFlagSSHKey != "" && addFlagSSHKey != "skip" {
		// Validate provided key path
		if err := user.ValidateSSHKeyPath(addFlagSSHKey); err != nil {
			return err
		}
		sshKeyPath = addFlagSSHKey
	} else if addFlagSSHKey == "skip" {
		// Skip SSH key setup when using flags
		sshKeyPath = ""
		ui.Info("Skipping SSH key setup")
	} else {
		// Interactive SSH key setup
		choice, err := ui.PromptSSHKeyOption()
		if err != nil {
			return fmt.Errorf("failed to get SSH key option: %w", err)
		}

		if strings.Contains(choice, "Generate new") {
			// Generate new key using system ssh-keygen (more reliable)
			privateKey, _, err := user.GenerateSSHKeySystem(githubUsername)
			if err != nil {
				return fmt.Errorf("failed to generate SSH key: %w", err)
			}

			sshKeyPath = privateKey
			ui.Success(fmt.Sprintf("SSH key generated: %s", privateKey))

			// Show public key content
			pubKeyContent, err := user.GetPublicKeyContent(privateKey)
			if err == nil {
				fmt.Println("\n" + strings.Repeat("-", 70))
				fmt.Println("Add this public key to your GitHub account:")
				fmt.Println("https://github.com/settings/keys")
				fmt.Println(strings.Repeat("-", 70))
				fmt.Print(pubKeyContent)
				fmt.Println(strings.Repeat("-", 70))
			}

		} else if strings.Contains(choice, "Import existing") {
			// Import existing key
			keyPath, err := ui.PromptExistingKeyPath()
			if err != nil {
				return fmt.Errorf("failed to get key path: %w", err)
			}

			if err := user.ValidateSSHKeyPath(keyPath); err != nil {
				return err
			}
			sshKeyPath = keyPath
			ui.Success(fmt.Sprintf("Using existing key: %s", keyPath))

		} else {
			// Skip for now
			sshKeyPath = ""
			ui.Info("SSH key setup skipped")
			fmt.Println("\nTo add SSH key later:")
			fmt.Printf("  1. Generate a key: ssh-keygen -t ed25519 -f %s\n", platform.GetExampleSSHKeyPath(githubUsername))
			fmt.Printf("  2. Edit config: %s %s\n", platform.GetEditorSuggestion(), platform.GetConfigFilePath())
			fmt.Printf("  3. Add: ssh_key_path = \"%s\"\n", platform.GetExampleSSHKeyPath(githubUsername))
			fmt.Printf("  4. Add public key to GitHub: https://github.com/settings/keys\n")
		}
	}

	// Create user
	newUser := config.User{
		Alias:       alias,
		Name:           name,
		Email:          email,
		GitHubUsername: githubUsername,
		SSHKeyPath:     sshKeyPath,
	}

	// Add user to config
	if err := cfg.AddUser(newUser); err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("User '%s' added successfully", alias))
	fmt.Println()
	fmt.Printf("Next: bgit use %s\n", alias)

	return nil
}
