package cmd

import (
	"fmt"
	"os"

	"github.com/byterings/bgit/core/config"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/git"
	"github.com/byterings/bgit/internal/identity"
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

	if !git.IsGitInstalled() {
		return fmt.Errorf("git is not installed")
	}

	if err := autoInit(); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var user *config.User
	if useByUsername {
		user = cfg.FindUserByUsername(identifier)
	} else if useByEmail {
		user = cfg.FindUserByEmail(identifier)
	} else {
		user = cfg.FindUser(identifier)
	}

	if user == nil {
		return fmt.Errorf("user '%s' not found\nRun: bgit list", identifier)
	}

	if err := capturePreviousGitIdentity(cfg); err != nil {
		ui.Warning(fmt.Sprintf("Could not back up current Git identity: %v", err))
	}

	if err := git.SetGlobalUser(user.Name, user.Email); err != nil {
		return fmt.Errorf("failed to update git config: %w", err)
	}

	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return fmt.Errorf("failed to update SSH config: %w", err)
	}

	cfg.ActiveUser = user.Alias
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if user.SSHKeyPath != "" {
		if coressh.EnsureKeyLoaded(user) {
			ui.Info("SSH key loaded into agent")
		}
	}

	ui.Success(fmt.Sprintf("Switched to identity: %s (%s)", user.Alias, user.Email))

	cwd, err := os.Getwd()
	if err == nil {
		resolution, _ := identity.ResolveIdentity(cfg, cwd)
		if resolution != nil && resolution.Alias != user.Alias {
			fmt.Println()
			switch resolution.Source {
			case identity.SourceWorkspace:
				ui.Warning(fmt.Sprintf("Note: Current directory is inside workspace '%s'", resolution.Path))
				ui.Info(fmt.Sprintf("bgit commands here will use '%s' identity", resolution.Alias))
			case identity.SourceBinding:
				ui.Warning("Note: Current repository is bound to a different identity")
				ui.Info(fmt.Sprintf("bgit commands here will use '%s' identity", resolution.Alias))
			}
		}
	}

	if user.SSHKeyPath != "" {
		fmt.Println("\nClone repos: bgit clone <url>")
		fmt.Println("Fix existing: bgit remote fix")
	}

	return nil
}

func capturePreviousGitIdentity(cfg *config.Config) error {
	if cfg.PreviousGitIdentitySet || cfg.ActiveUser != "" {
		return nil
	}

	name, email, err := git.GetGlobalUser()
	if err != nil {
		return err
	}

	cfg.PreviousGitName = name
	cfg.PreviousGitEmail = email
	cfg.PreviousGitIdentitySet = true
	return nil
}
