package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/byterings/bgit/core/config"
	coresetup "github.com/byterings/bgit/core/setup"
	coressh "github.com/byterings/bgit/core/ssh"
	coreuninstall "github.com/byterings/bgit/core/uninstall"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Safely uninstall bgit and restore all repositories",
	Long: `Safely uninstall bgit by:
1. Restoring repositories with bgit remote URLs to standard GitHub format
2. Removing bgit SSH config entries
3. Restoring or clearing bgit-managed global hook configuration
4. Restoring backed-up Git identity when available
5. Removing bgit configuration

This ensures your repositories continue to work after bgit is removed.`,
	Example: `  # Uninstall bgit safely
  bgit uninstall

  # Also remove bgit-generated SSH keys referenced in config
  bgit uninstall --remove-keys

  # After running this command, manually delete:
  # Linux/macOS: sudo rm /usr/local/bin/bgit
  # Windows: Remove from Add/Remove Programs or delete the install folder`,
	RunE: runUninstall,
}

var (
	uninstallSkipRepos  bool
	uninstallForce      bool
	uninstallRemoveKeys bool
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVar(&uninstallSkipRepos, "skip-repos", false, "Skip scanning and fixing repositories")
	uninstallCmd.Flags().BoolVar(&uninstallForce, "force", false, "Skip confirmation prompt")
	uninstallCmd.Flags().BoolVar(&uninstallRemoveKeys, "remove-keys", false, "Delete bgit-generated SSH keys referenced in config")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("bgit Uninstall")
	fmt.Println("==============")
	fmt.Println()

	if !uninstallForce {
		fmt.Println("This will:")
		fmt.Println("  1. Restore repositories with bgit remote URLs")
		fmt.Println("  2. Restore them to standard GitHub format")
		fmt.Println("  3. Remove bgit SSH config entries")
		fmt.Println("  4. Restore or report global Git identity")
		fmt.Println("  5. Remove bgit configuration (~/.bgit)")
		fmt.Println()

		confirmed, err := ui.PromptConfirmation("Continue?")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Operation cancelled.")
			return nil
		}
		fmt.Println()
	}

	var fixedRepos []string
	var failedRepos []string
	var cfg *config.Config

	loadedCfg, err := config.LoadConfig()
	if err == nil {
		cfg = loadedCfg
	} else {
		ui.Warning(fmt.Sprintf("Could not load bgit config before uninstall: %v", err))
	}

	if !uninstallSkipRepos {
		fmt.Println("Step 1: Restoring repository remotes...")
		if cfg != nil {
			configFixed, configFailed := restoreConfiguredRepos(cfg)
			fixedRepos = append(fixedRepos, configFixed...)
			failedRepos = append(failedRepos, configFailed...)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			ui.Error("Failed to get home directory")
		} else {
			scanFixed, scanFailed := scanAndFixRepos(homeDir)
			fixedRepos = appendUnique(fixedRepos, scanFixed...)
			failedRepos = appendUnique(failedRepos, scanFailed...)
		}
		fmt.Println()
	} else {
		fmt.Println("Step 1: Skipped (--skip-repos)")
		fmt.Println()
	}

	fmt.Println("Step 2: Removing SSH config entries...")
	if err := removeSSHConfigEntries(); err != nil {
		ui.Error(fmt.Sprintf("Failed to remove SSH config: %v", err))
	} else {
		ui.Success("SSH config entries removed")
	}
	fmt.Println()

	fmt.Println("Step 3: Removing global hook configuration...")
	if err := restoreGlobalHooksPath(cfg); err != nil {
		ui.Warning(fmt.Sprintf("Could not clear global hooks path: %v", err))
	} else {
		ui.Success("Global hooks path restored")
	}
	fmt.Println()

	fmt.Println("Step 4: Restoring Git identity...")
	restoreGitIdentity(cfg)
	fmt.Println()

	if uninstallRemoveKeys {
		fmt.Println("Step 5: Removing bgit-generated SSH keys...")
		removeGeneratedSSHKeys(cfg)
		fmt.Println()
	} else {
		fmt.Println("Step 5: Skipping SSH key removal")
		ui.Info("Use --remove-keys to delete bgit-generated SSH keys referenced in config.")
		fmt.Println()
	}

	fmt.Println("Step 6: Removing bgit configuration...")
	configDir, err := config.GetConfigDir()
	if err == nil {
		if err := os.RemoveAll(configDir); err != nil {
			ui.Error(fmt.Sprintf("Failed to remove config: %v", err))
		} else {
			ui.Success(fmt.Sprintf("Removed %s", configDir))
		}
	}
	fmt.Println()

	fmt.Println("==============")
	fmt.Println("Summary")
	fmt.Println("==============")

	if len(fixedRepos) > 0 {
		fmt.Printf("\nRepositories restored (%d):\n", len(fixedRepos))
		for _, repo := range fixedRepos {
			fmt.Printf("  ✓ %s\n", repo)
		}
	}

	if len(failedRepos) > 0 {
		fmt.Printf("\nRepositories failed (%d):\n", len(failedRepos))
		for _, repo := range failedRepos {
			fmt.Printf("  ✗ %s\n", repo)
		}
	}

	fmt.Println()
	ui.Success("bgit uninstall complete!")
	fmt.Println()
	fmt.Println("Final step - manually remove the bgit binary:")
	if runtime.GOOS == "windows" {
		fmt.Println("  Option 1: Settings → Apps → bgit → Uninstall")
		fmt.Println("  Option 2: Remove-Item \"$env:LOCALAPPDATA\\bgit\" -Recurse -Force")
	} else {
		fmt.Println("  sudo rm /usr/local/bin/bgit")
	}
	fmt.Println()

	return nil
}

func restoreConfiguredRepos(cfg *config.Config) (fixed []string, failed []string) {
	return coreuninstall.RestoreConfiguredRepos(cfg)
}

func scanAndFixRepos(startPath string) (fixed []string, failed []string) {
	return coreuninstall.ScanAndRestoreRepos(startPath)
}

func removeSSHConfigEntries() error {
	return coressh.RemoveManagedSectionFromFile()
}

func restoreGlobalHooksPath(cfg *config.Config) error {
	return coresetup.RestoreGlobalHooksPath(cfg)
}

// restoreGitIdentity restores the pre-bgit Git identity and prints the outcome.
func restoreGitIdentity(cfg *config.Config) {
	if cfg == nil {
		ui.Warning("bgit config was unavailable; cannot determine previous Git identity.")
		return
	}

	result := coreuninstall.RestoreGitIdentity(cfg)

	if result.HadBackup {
		if result.NameErr != nil {
			ui.Warning(fmt.Sprintf("Could not restore git user.name: %v", result.NameErr))
		}
		if result.EmailErr != nil {
			ui.Warning(fmt.Sprintf("Could not restore git user.email: %v", result.EmailErr))
		}
		ui.Success("Git identity restored")
		return
	}

	if result.InspectErr != nil {
		ui.Warning(fmt.Sprintf("Could not inspect Git identity: %v", result.InspectErr))
		return
	}

	if result.MatchingUser != nil {
		ui.Warning(fmt.Sprintf("Global Git identity still matches bgit user '%s'.", result.MatchingUser.Alias))
		fmt.Println("bgit cannot know the previous identity because this older config did not store it.")
		fmt.Println("Set it manually, for example:")
		fmt.Println(`  git config --global user.name "Your Name"`)
		fmt.Println(`  git config --global user.email "you@example.com"`)
		return
	}

	ui.Info("No previous Git identity backup found; current Git identity was left unchanged.")
}

// removeGeneratedSSHKeys deletes bgit-generated keys and prints the outcome.
func removeGeneratedSSHKeys(cfg *config.Config) {
	if cfg == nil {
		ui.Warning("bgit config was unavailable; cannot determine generated SSH keys.")
		return
	}

	removed := 0
	for _, removal := range coreuninstall.RemoveGeneratedKeys(cfg) {
		if removal.Err != nil {
			ui.Warning(fmt.Sprintf("Could not remove %s: %v", removal.Path, removal.Err))
			continue
		}
		removed++
		ui.Success(fmt.Sprintf("Removed %s", removal.Path))
	}

	if removed == 0 {
		ui.Info("No bgit-generated SSH keys found.")
	}
}

func appendUnique(items []string, more ...string) []string {
	return coreuninstall.AppendUnique(items, more...)
}
