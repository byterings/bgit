package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/ssh"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Safely uninstall bgit and restore all repositories",
	Long: `Safely uninstall bgit by:
1. Finding all git repositories with bgit remote URLs
2. Restoring them to standard GitHub format
3. Removing bgit SSH config entries
4. Removing bgit configuration

This ensures your repositories continue to work after bgit is removed.`,
	Example: `  # Uninstall bgit safely
  bgit uninstall

  # After running this command, manually delete:
  # Linux/macOS: sudo rm /usr/local/bin/bgit
  # Windows: Remove from Add/Remove Programs or delete the install folder`,
	RunE: runUninstall,
}

var (
	uninstallSkipRepos bool
	uninstallForce     bool
	uninstallDryRun    bool
	uninstallVerbose   bool
	uninstallRemoveCfg bool
	uninstallRemoveBin bool
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVar(&uninstallSkipRepos, "skip-repos", false, "Skip scanning and fixing repositories")
	uninstallCmd.Flags().BoolVar(&uninstallForce, "force", false, "Skip confirmation prompt")
	uninstallCmd.Flags().BoolVar(&uninstallDryRun, "dry-run", false, "Show what would be removed or restored without changing files")
	uninstallCmd.Flags().BoolVar(&uninstallVerbose, "verbose", false, "Show detailed cleanup actions")
	uninstallCmd.Flags().BoolVar(&uninstallRemoveCfg, "remove-config", false, "Remove bgit config and saved identities after cleanup")
	uninstallCmd.Flags().BoolVar(&uninstallRemoveBin, "remove-binary", false, "Attempt to remove the bgit binary or install directory when safe")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("bgit Uninstall")
	fmt.Println("==============")
	fmt.Println()

	if !uninstallForce {
		fmt.Println("This will:")
		fmt.Println("  1. Scan for repositories with bgit remote URLs")
		fmt.Println("  2. Restore them to standard GitHub format")
		fmt.Println("  3. Remove bgit SSH config entries")
		fmt.Println("  4. Restore bgit-owned Git configuration")
		fmt.Println("  5. Preserve bgit identities/config unless --remove-config is used")
		if uninstallDryRun {
			fmt.Println()
			ui.Info("Dry run enabled; no files or settings will be changed")
		}
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
	var errors []string

	cfg, cfgErr := config.LoadConfig()
	if cfgErr != nil && uninstallVerbose {
		ui.Warning(fmt.Sprintf("Could not load bgit config for restoration metadata: %v", cfgErr))
	}

	if !uninstallSkipRepos {
		fmt.Println("Step 1: Scanning for repositories...")
		homeDir, err := os.UserHomeDir()
		if err != nil {
			ui.Error("Failed to get home directory")
			errors = append(errors, err.Error())
		} else {
			fixedRepos, failedRepos = scanAndFixRepos(homeDir, uninstallDryRun, uninstallVerbose)
		}
		fmt.Println()
	} else {
		fmt.Println("Step 1: Skipped (--skip-repos)")
		fmt.Println()
	}

	fmt.Println("Step 2: Removing SSH config entries...")
	if err := removeSSHConfigEntries(uninstallDryRun); err != nil {
		ui.Error(fmt.Sprintf("Failed to remove SSH config: %v", err))
		errors = append(errors, err.Error())
	} else {
		if uninstallDryRun {
			ui.Info("Would remove bgit-managed SSH config entries")
		} else {
			ui.Success("SSH config entries removed")
		}
	}
	fmt.Println()

	fmt.Println("Step 3: Removing global hook configuration...")
	if err := restoreManagedHooksPath(cfg, uninstallDryRun); err != nil {
		ui.Warning(fmt.Sprintf("Could not clear global hooks path: %v", err))
		errors = append(errors, err.Error())
	} else {
		if uninstallDryRun {
			ui.Info("Would restore or unset bgit-managed core.hooksPath")
		} else {
			ui.Success("Global hooks path restored")
		}
	}
	fmt.Println()

	fmt.Println("Step 4: Restoring global Git identity...")
	if err := restoreGitIdentity(cfg, uninstallDryRun, uninstallVerbose); err != nil {
		ui.Warning(fmt.Sprintf("Could not restore global Git identity: %v", err))
		errors = append(errors, err.Error())
	}
	fmt.Println()

	fmt.Println("Step 5: Cleaning PATH entries...")
	if err := cleanupInstallPath(uninstallDryRun, uninstallVerbose); err != nil {
		ui.Warning(fmt.Sprintf("Could not clean PATH: %v", err))
		errors = append(errors, err.Error())
	}
	fmt.Println()

	fmt.Println("Step 6: Removing bgit configuration...")
	if uninstallRemoveCfg {
		configDir, err := config.GetConfigDir()
		if err == nil {
			if uninstallDryRun {
				ui.Info(fmt.Sprintf("Would remove %s", configDir))
			} else if err := os.RemoveAll(configDir); err != nil {
				ui.Error(fmt.Sprintf("Failed to remove config: %v", err))
				errors = append(errors, err.Error())
			} else {
				ui.Success(fmt.Sprintf("Removed %s", configDir))
			}
		}
	} else {
		ui.Info("Preserved bgit config and identities. Use --remove-config to delete them.")
	}
	fmt.Println()

	if uninstallRemoveBin {
		fmt.Println("Step 7: Removing bgit binary...")
		if err := removeInstalledBinary(uninstallDryRun, uninstallVerbose); err != nil {
			ui.Warning(fmt.Sprintf("Could not remove binary automatically: %v", err))
			errors = append(errors, err.Error())
		}
		fmt.Println()
	}

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
	if len(errors) > 0 || len(failedRepos) > 0 {
		ui.Warning("bgit uninstall completed with warnings")
	} else if uninstallDryRun {
		ui.Success("bgit uninstall dry run complete")
	} else {
		ui.Success("bgit uninstall complete")
	}
	fmt.Println()
	if !uninstallRemoveBin {
		fmt.Println("Final step - manually remove the bgit binary, or rerun with --remove-binary:")
		if runtime.GOOS == "windows" {
			fmt.Println("  Option 1: Settings → Apps → bgit → Uninstall")
			fmt.Println("  Option 2: Remove-Item \"$env:LOCALAPPDATA\\bgit\" -Recurse -Force")
		} else {
			fmt.Println("  sudo rm /usr/local/bin/bgit")
		}
		fmt.Println()
	}

	return nil
}

func scanAndFixRepos(startPath string, dryRun, verbose bool) (fixed []string, failed []string) {
	scanDirs := []string{startPath}

	commonDirs := []string{"Documents", "Projects", "repos", "src", "code", "work", "dev", "git"}
	for _, dir := range commonDirs {
		fullPath := filepath.Join(startPath, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			scanDirs = append(scanDirs, fullPath)
		}
	}

	visited := make(map[string]bool)
	bgitPattern := regexp.MustCompile(`github\.com-`)

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
				visited[repoPath] = true

				url, err := getRepoRemoteURL(repoPath)
				if err != nil || url == "" {
					return filepath.SkipDir
				}

				if bgitPattern.MatchString(url) {
					newURL, err := convertToStandardURL(url)
					if err != nil {
						failed = append(failed, repoPath)
						return filepath.SkipDir
					}

					if verbose || dryRun {
						fmt.Printf("  %s: %s -> %s\n", repoPath, url, newURL)
					}
					if dryRun {
						fixed = append(fixed, repoPath)
					} else if err := setRepoRemoteURL(repoPath, "origin", newURL); err != nil {
						failed = append(failed, repoPath)
					} else {
						fixed = append(fixed, repoPath)
					}
				}

				return filepath.SkipDir // Don't descend into .git
			}

			return nil
		})
	}

	return fixed, failed
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

func removeSSHConfigEntries(dryRun bool) error {
	if dryRun {
		return nil
	}
	return ssh.RemoveManagedConfig()
}

func restoreGitIdentity(cfg *config.Config, dryRun, verbose bool) error {
	if cfg == nil || !cfg.GitIdentityBackedUp {
		if verbose {
			ui.Info("No backed-up Git identity found; leaving user.name and user.email unchanged")
		}
		return nil
	}

	currentName, currentEmail, err := getGlobalGitIdentity()
	if err != nil {
		return err
	}

	if !matchesKnownBgitIdentity(cfg, currentName, currentEmail) {
		if verbose {
			ui.Info("Current Git identity is not a bgit-managed identity; leaving it unchanged")
		}
		return nil
	}

	if dryRun {
		ui.Info(fmt.Sprintf("Would restore Git identity to %s <%s>", cfg.PreviousGitName, cfg.PreviousGitEmail))
		return nil
	}

	if cfg.PreviousGitName == "" {
		if err := unsetGlobalGitConfig("user.name"); err != nil {
			return err
		}
	} else if err := setGlobalGitConfig("user.name", cfg.PreviousGitName); err != nil {
		return err
	}

	if cfg.PreviousGitEmail == "" {
		if err := unsetGlobalGitConfig("user.email"); err != nil {
			return err
		}
	} else if err := setGlobalGitConfig("user.email", cfg.PreviousGitEmail); err != nil {
		return err
	}

	ui.Success("Global Git identity restored")
	return nil
}

func matchesKnownBgitIdentity(cfg *config.Config, name, email string) bool {
	for _, user := range cfg.Users {
		if user.Name == name && user.Email == email {
			return true
		}
	}
	return false
}

func getGlobalGitIdentity() (string, string, error) {
	name, err := getGlobalGitConfig("user.name")
	if err != nil {
		return "", "", err
	}
	email, err := getGlobalGitConfig("user.email")
	if err != nil {
		return "", "", err
	}
	return name, email, nil
}

func getGlobalGitConfig(key string) (string, error) {
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

func setGlobalGitConfig(key, value string) error {
	cmd := exec.Command("git", "config", "--global", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set %s: %s: %w", key, string(output), err)
	}
	return nil
}

func unsetGlobalGitConfig(key string) error {
	cmd := exec.Command("git", "config", "--global", "--unset", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 5 {
			return nil
		}
		return fmt.Errorf("failed to unset %s: %s: %w", key, string(output), err)
	}
	return nil
}

func cleanupInstallPath(dryRun, verbose bool) error {
	if runtime.GOOS != "windows" {
		if verbose {
			ui.Info("PATH cleanup is only automated for Windows user installs")
		}
		return nil
	}

	installDir := os.Getenv("LOCALAPPDATA")
	if installDir == "" {
		return nil
	}
	installDir = filepath.Join(installDir, "bgit")
	if dryRun {
		ui.Info(fmt.Sprintf("Would remove exact PATH entry: %s", installDir))
		return nil
	}

	script := fmt.Sprintf(`$dir = %q
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($null -ne $path) {
  $target = [IO.Path]::GetFullPath($dir)
  $parts = @()
  foreach ($entry in ($path -split ';')) {
    if (-not $entry) { continue }
    try {
      if ([IO.Path]::GetFullPath($entry.Trim('"')) -ine $target) {
        $parts += $entry
      }
    } catch {
      $parts += $entry
    }
  }
  [Environment]::SetEnvironmentVariable("Path", ($parts -join ';'), "User")
}`, installDir)
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	ui.Success("Removed bgit install directory from user PATH")
	return nil
}

func removeInstalledBinary(dryRun, verbose bool) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exePath, _ = filepath.EvalSymlinks(exePath)
	base := strings.ToLower(filepath.Base(exePath))
	if base != "bgit" && base != "bgit.exe" {
		return fmt.Errorf("current executable is not named bgit: %s", exePath)
	}
	if !isSafeInstallPath(exePath) {
		return fmt.Errorf("refusing to remove unexpected executable path: %s", exePath)
	}
	if dryRun {
		ui.Info(fmt.Sprintf("Would remove %s", exePath))
		return nil
	}
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Windows cannot remove a running executable; use Settings > Apps or uninstall.ps1")
	}
	if err := os.Remove(exePath); err != nil {
		return err
	}
	ui.Success(fmt.Sprintf("Removed %s", exePath))
	return nil
}

func isSafeInstallPath(exePath string) bool {
	dir := filepath.Dir(exePath)
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		return localAppData != "" && filepath.Clean(dir) == filepath.Clean(filepath.Join(localAppData, "bgit"))
	}

	home, _ := os.UserHomeDir()
	safeDirs := []string{"/usr/local/bin", "/opt/homebrew/bin"}
	if home != "" {
		safeDirs = append(safeDirs, filepath.Join(home, ".local", "bin"))
	}
	for _, safeDir := range safeDirs {
		if filepath.Clean(dir) == filepath.Clean(safeDir) {
			return true
		}
	}
	return false
}
