package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/internal/platform"
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
	seen := make(map[string]bool)

	for _, binding := range cfg.GetBindings() {
		restoreRepoRemote(binding.Path, seen, &fixed, &failed)
	}

	for _, workspace := range cfg.GetWorkspaces() {
		workspaceFixed, workspaceFailed := scanAndFixRepos(workspace.Path)
		fixed = appendUnique(fixed, workspaceFixed...)
		failed = appendUnique(failed, workspaceFailed...)
	}

	return fixed, failed
}

func scanAndFixRepos(startPath string) (fixed []string, failed []string) {
	scanDirs := []string{startPath}

	commonDirs := []string{"Documents", "Projects", "repos", "src", "code", "work", "dev", "git"}
	for _, dir := range commonDirs {
		fullPath := filepath.Join(startPath, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			scanDirs = append(scanDirs, fullPath)
		}
	}

	visited := make(map[string]bool)

	for _, scanDir := range scanDirs {
		filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
				return filepath.SkipDir
			}

			skipDirs := []string{"node_modules", "vendor", ".cache", ".local", "snap", ".npm", ".cargo"}
			for _, skip := range skipDirs {
				if info.Name() == skip {
					return filepath.SkipDir
				}
			}

			if info.IsDir() && info.Name() == ".git" {
				repoPath := filepath.Dir(path)

				if visited[repoPath] {
					return filepath.SkipDir
				}

				restoreRepoRemote(repoPath, visited, &fixed, &failed)

				return filepath.SkipDir // Don't descend into .git
			}

			return nil
		})
	}

	return fixed, failed
}

func restoreRepoRemote(repoPath string, visited map[string]bool, fixed, failed *[]string) {
	if visited[repoPath] {
		return
	}
	visited[repoPath] = true

	url, err := getRepoRemoteURL(repoPath)
	if err != nil || url == "" || !strings.Contains(url, "github.com-") {
		return
	}

	newURL, err := convertToStandardURL(url)
	if err != nil {
		*failed = appendUnique(*failed, repoPath)
		return
	}

	if err := setRepoRemoteURL(repoPath, "origin", newURL); err != nil {
		*failed = appendUnique(*failed, repoPath)
	} else {
		*fixed = appendUnique(*fixed, repoPath)
	}
}

func getRepoRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func setRepoRemoteURL(repoPath, remote, url string) error {
	cmd := exec.Command("git", "-C", repoPath, "remote", "set-url", remote, url)
	return cmd.Run()
}

func removeSSHConfigEntries() error {
	sshConfigPath, err := platform.GetSSHConfigPath()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(sshConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inBgitSection := false

	for _, line := range lines {
		if strings.Contains(line, "BEGIN BGIT MANAGED") || strings.Contains(line, "BEGIN BRGIT MANAGED") {
			inBgitSection = true
			continue
		}
		if strings.Contains(line, "END BGIT MANAGED") || strings.Contains(line, "END BRGIT MANAGED") {
			inBgitSection = false
			continue
		}
		if !inBgitSection {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	newContent = strings.TrimRight(newContent, "\n") + "\n"

	return os.WriteFile(sshConfigPath, []byte(newContent), 0600)
}

func restoreGlobalHooksPath(cfg *config.Config) error {
	currentPath, err := getGlobalHooksPath()
	if err != nil {
		return err
	}

	if currentPath == "" {
		return nil
	}

	if !isBgitHooksPath(currentPath) {
		return nil
	}

	if cfg != nil && cfg.PreviousHooksPathSet {
		if cfg.PreviousHooksPath == "" {
			return unsetGlobalConfig("core.hooksPath")
		}
		return setGlobalHooksPath(cfg.PreviousHooksPath)
	}

	return unsetGlobalConfig("core.hooksPath")
}

func isBgitHooksPath(path string) bool {
	cleaned := filepath.ToSlash(filepath.Clean(path))
	if strings.Contains(cleaned, "/.bgit/hooks") || strings.HasSuffix(cleaned, ".bgit/hooks") {
		return true
	}

	hookPath := filepath.Join(path, prePushHookName)
	content, err := os.ReadFile(hookPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "BEGIN BGIT MANAGED") ||
		strings.Contains(string(content), "BEGIN BRGIT MANAGED")
}

func restoreGitIdentity(cfg *config.Config) {
	if cfg == nil {
		ui.Warning("bgit config was unavailable; cannot determine previous Git identity.")
		return
	}

	if cfg.PreviousGitIdentitySet {
		if err := restoreGitConfigValue("user.name", cfg.PreviousGitName); err != nil {
			ui.Warning(fmt.Sprintf("Could not restore git user.name: %v", err))
		}
		if err := restoreGitConfigValue("user.email", cfg.PreviousGitEmail); err != nil {
			ui.Warning(fmt.Sprintf("Could not restore git user.email: %v", err))
		}
		ui.Success("Git identity restored")
		return
	}

	name, email, err := getGlobalGitIdentity()
	if err != nil {
		ui.Warning(fmt.Sprintf("Could not inspect Git identity: %v", err))
		return
	}

	for _, user := range cfg.Users {
		if name == user.Name && email == user.Email {
			ui.Warning(fmt.Sprintf("Global Git identity still matches bgit user '%s'.", user.Alias))
			fmt.Println("bgit cannot know the previous identity because this older config did not store it.")
			fmt.Println("Set it manually, for example:")
			fmt.Println(`  git config --global user.name "Your Name"`)
			fmt.Println(`  git config --global user.email "you@example.com"`)
			return
		}
	}

	ui.Info("No previous Git identity backup found; current Git identity was left unchanged.")
}

func restoreGitConfigValue(key, value string) error {
	if value == "" {
		return unsetGlobalConfig(key)
	}

	cmd := exec.Command("git", "config", "--global", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(output), err)
	}
	return nil
}

func unsetGlobalConfig(key string) error {
	cmd := exec.Command("git", "config", "--global", "--unset", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 5 {
			return nil
		}
		return fmt.Errorf("%s: %w", string(output), err)
	}
	return nil
}

func getGlobalGitIdentity() (name, email string, err error) {
	name, err = getGlobalConfigValue("user.name")
	if err != nil {
		return "", "", err
	}

	email, err = getGlobalConfigValue("user.email")
	if err != nil {
		return "", "", err
	}

	return name, email, nil
}

func getGlobalConfigValue(key string) (string, error) {
	cmd := exec.Command("git", "config", "--global", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func removeGeneratedSSHKeys(cfg *config.Config) {
	if cfg == nil {
		ui.Warning("bgit config was unavailable; cannot determine generated SSH keys.")
		return
	}

	removed := 0
	for _, user := range cfg.Users {
		if user.SSHKeyPath == "" || !isBgitGeneratedKey(user.SSHKeyPath) {
			continue
		}

		for _, path := range []string{user.SSHKeyPath, user.SSHKeyPath + ".pub"} {
			if err := os.Remove(path); err != nil {
				if !os.IsNotExist(err) {
					ui.Warning(fmt.Sprintf("Could not remove %s: %v", path, err))
				}
				continue
			}
			removed++
			ui.Success(fmt.Sprintf("Removed %s", path))
		}
	}

	if removed == 0 {
		ui.Info("No bgit-generated SSH keys found.")
	}
}

func isBgitGeneratedKey(path string) bool {
	base := filepath.Base(path)
	return strings.HasPrefix(base, "bgit_") || strings.HasPrefix(base, "brgit_")
}

func appendUnique(items []string, more ...string) []string {
	seen := make(map[string]bool, len(items)+len(more))
	for _, item := range items {
		seen[item] = true
	}
	for _, item := range more {
		if item == "" || seen[item] {
			continue
		}
		items = append(items, item)
		seen[item] = true
	}
	return items
}
