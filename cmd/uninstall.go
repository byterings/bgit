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
	"github.com/byterings/bgit/internal/platform"
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

	if !uninstallForce {
		fmt.Println("This will:")
		fmt.Println("  1. Scan for repositories with bgit remote URLs")
		fmt.Println("  2. Restore them to standard GitHub format")
		fmt.Println("  3. Remove bgit SSH config entries")
		fmt.Println("  4. Remove bgit configuration (~/.bgit)")
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

	fmt.Println("Step 2: Removing SSH config entries...")
	if err := removeSSHConfigEntries(); err != nil {
		ui.Error(fmt.Sprintf("Failed to remove SSH config: %v", err))
	} else {
		ui.Success("SSH config entries removed")
	}
	fmt.Println()

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

	newContent := strings.Join(newLines, "\n")
	newContent = strings.TrimRight(newContent, "\n") + "\n"

	return os.WriteFile(sshConfigPath, []byte(newContent), 0600)
}
