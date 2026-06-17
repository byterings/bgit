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

	mode := coreidentity.LookupByAlias
	if useByUsername {
		mode = coreidentity.LookupByUsername
	} else if useByEmail {
		mode = coreidentity.LookupByEmail
	}

	result, err := coreidentity.ActivateUser(cfg, identifier, mode)
	if err != nil {
		return fmt.Errorf("%w\nRun: bgit list", err)
	}
	user := result.User
	if result.KeyLoaded {
		ui.Info("SSH key loaded into agent")
	}
	resynced, syncFailures := syncBoundReposForAlias(cfg, user.Alias, user.Name, user.Email)

	ui.Success(fmt.Sprintf("Switched to identity: %s (%s)", user.Alias, user.Email))
	if resynced > 0 {
		ui.Success(fmt.Sprintf("Re-synced %d bound repos for '%s'", resynced, user.Alias))
	}
	for _, syncFailure := range syncFailures {
		ui.Warning(syncFailure)
	}

	cwd, err := os.Getwd()
	if err == nil {
		resolution, _ := coreidentity.ResolveIdentity(cfg, cwd)
		if resolution != nil && resolution.Alias != user.Alias {
			fmt.Println()
			switch resolution.Source {
			case coreidentity.SourceWorkspace:
				ui.Warning(fmt.Sprintf("Note: Current directory is inside workspace '%s'", resolution.Path))
				ui.Info(fmt.Sprintf("bgit commands here will use '%s' identity", resolution.Alias))
			case coreidentity.SourceBinding:
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

func syncBoundReposForAlias(cfg *config.Config, alias, name, email string) (int, []string) {
	resynced := 0
	failures := []string{}

	for _, binding := range cfg.GetBindings() {
		if binding.User != alias {
			continue
		}
		if err := git.SetLocalUser(binding.Path, name, email); err != nil {
			failures = append(failures, fmt.Sprintf("Could not sync bound repo '%s': %v", binding.Path, err))
			continue
		}
		resynced++
	}

	return resynced, failures
}
