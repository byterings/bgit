package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/platform"
	"github.com/byterings/bgit/internal/ui"
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
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVar(&uninstallSkipRepos, "skip-repos", false, "Skip scanning and fixing repositories")
	uninstallCmd.Flags().BoolVar(&uninstallForce, "force", false, "Skip confirmation prompt")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("bgit Uninstall")
	fmt.Println("==============")
	fmt.Println()

	// Confirmation
	if !uninstallForce {
		fmt.Println("This will:")
		fmt.Println("  1. Scan for repositories with bgit remote URLs")
		fmt.Println("  2. Restore them to standard GitHub format")
		fmt.Println("  3. Remove bgit SSH config entries")
		fmt.Println("  4. Remove bgit configuration (~/.bgit)")
		fmt.Println()
		fmt.Print("Continue? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Uninstall cancelled.")
			return nil
		}
		fmt.Println()
	}

	var fixedRepos []string
	var failedRepos []string

	// Step 1: Find and fix repositories
	if !uninstallSkipRepos {
		fmt.Println("Step 1: Scanning for repositories...")
		homeDir, err := os.UserHomeDir()
		if err != nil {
			ui.Error("Failed to get home directory")
		} else {
			fixedRepos, failedRepos = scanAndFixRepos(homeDir)
		}
		fmt.Println()
	} else {
		fmt.Println("Step 1: Skipped (--skip-repos)")
		fmt.Println()
	}

	// Step 2: Remove SSH config entries
	fmt.Println("Step 2: Removing SSH config entries...")
	if err := removeSSHConfigEntries(); err != nil {
		ui.Error(fmt.Sprintf("Failed to remove SSH config: %v", err))
	} else {
		ui.Success("SSH config entries removed")
	}
	fmt.Println()

	// Step 3: Remove bgit config
	fmt.Println("Step 3: Removing bgit configuration...")
	configDir, err := config.GetConfigDir()
	if err == nil {
		if err := os.RemoveAll(configDir); err != nil {
			ui.Error(fmt.Sprintf("Failed to remove config: %v", err))
		} else {
			ui.Success(fmt.Sprintf("Removed %s", configDir))
		}
	}
	fmt.Println()

	// Summary
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

// scanAndFixRepos scans for git repos with bgit URLs and fixes them
func scanAndFixRepos(startPath string) (fixed []string, failed []string) {
	// Common directories to scan
	scanDirs := []string{
		startPath,
	}

	// Add common project directories
	commonDirs := []string{"Documents", "Projects", "repos", "src", "code", "work", "dev", "git"}
	for _, dir := range commonDirs {
		fullPath := filepath.Join(startPath, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			scanDirs = append(scanDirs, fullPath)
		}
	}

	// Track visited directories to avoid duplicates
	visited := make(map[string]bool)

	bgitPattern := regexp.MustCompile(`github\.com-`)

	for _, scanDir := range scanDirs {
		filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			// Skip hidden directories (except .git)
			if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
				return filepath.SkipDir
			}

			// Skip common non-project directories
			skipDirs := []string{"node_modules", "vendor", ".cache", ".local", "snap", ".npm", ".cargo"}
			for _, skip := range skipDirs {
				if info.Name() == skip {
					return filepath.SkipDir
				}
			}

			// Look for .git directories
			if info.IsDir() && info.Name() == ".git" {
				repoPath := filepath.Dir(path)

				// Skip if already visited
				if visited[repoPath] {
					return filepath.SkipDir
				}
				visited[repoPath] = true

				// Check if remote uses bgit format
				url, err := getRepoRemoteURL(repoPath)
				if err != nil || url == "" {
					return filepath.SkipDir
				}

				if bgitPattern.MatchString(url) {
					// Fix this repo
					newURL, err := convertToStandardURL(url)
					if err != nil {
						failed = append(failed, repoPath)
						return filepath.SkipDir
					}

					if err := setRepoRemoteURL(repoPath, "origin", newURL); err != nil {
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

// getRepoRemoteURL gets remote URL for a specific repo
func getRepoRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// setRepoRemoteURL sets remote URL for a specific repo
func setRepoRemoteURL(repoPath, remote, url string) error {
	cmd := exec.Command("git", "-C", repoPath, "remote", "set-url", remote, url)
	return cmd.Run()
}

// removeSSHConfigEntries removes bgit-managed SSH config entries
func removeSSHConfigEntries() error {
	sshConfigPath, err := platform.GetSSHConfigPath()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(sshConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No SSH config, nothing to do
		}
		return err
	}

	// Remove the bgit-managed section
	lines := strings.Split(string(content), "\n")
	var newLines []string
	inBgitSection := false

	for _, line := range lines {
		if strings.Contains(line, "BEGIN BRGIT MANAGED") {
			inBgitSection = true
			continue
		}
		if strings.Contains(line, "END BRGIT MANAGED") {
			inBgitSection = false
			continue
		}
		if !inBgitSection {
			newLines = append(newLines, line)
		}
	}

	// Write back
	newContent := strings.Join(newLines, "\n")
	// Remove extra blank lines at the end
	newContent = strings.TrimRight(newContent, "\n") + "\n"

	return os.WriteFile(sshConfigPath, []byte(newContent), 0600)
}
