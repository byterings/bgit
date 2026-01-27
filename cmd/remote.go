package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage git remotes for bgit",
	Long:  `Commands to manage git remote URLs for bgit compatibility.`,
}

var remoteFixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Convert remote URL to use active user's SSH config",
	Long: `Convert the current repository's origin remote URL to use the active user's SSH host alias.

This allows git push/pull to work with the correct SSH key.`,
	Example: `  # Fix current repo's remote
  bgit use work
  bgit remote fix

  # Now git push works with the work identity`,
	RunE: runRemoteFix,
}

var remoteRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore remote URL to standard GitHub format",
	Long: `Convert the current repository's origin remote URL back to standard GitHub format.

Use this before uninstalling bgit or if you want to use standard git SSH.`,
	Example: `  # Restore current repo's remote
  bgit remote restore

  # Remote is now: git@github.com:user/repo.git`,
	RunE: runRemoteRestore,
}

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(remoteFixCmd)
	remoteCmd.AddCommand(remoteRestoreCmd)
}

func runRemoteFix(cmd *cobra.Command, args []string) error {
	// Check if we're in a git repo
	if !isGitRepo() {
		return fmt.Errorf("not a git repository\nRun this command inside a git repository")
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

	// Check for active user
	if cfg.ActiveUser == "" {
		return fmt.Errorf("no active user set\nRun: bgit use <alias>")
	}

	activeUser := cfg.FindUserByAlias(cfg.ActiveUser)
	if activeUser == nil {
		return fmt.Errorf("active user '%s' not found in config", cfg.ActiveUser)
	}

	// Get current remote URL
	currentURL, err := getRemoteURL("origin")
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	if currentURL == "" {
		return fmt.Errorf("no 'origin' remote found\nAdd a remote first: git remote add origin <url>")
	}

	// Check if repo is already configured for a different user
	existingUsername := extractAliasFromURL(currentURL)
	if existingUsername != "" && existingUsername != activeUser.GitHubUsername {
		ui.Warning(fmt.Sprintf("This repo is configured for GitHub user '%s' but effective user is '%s' (%s)", existingUsername, activeUser.Alias, activeUser.GitHubUsername))
		fmt.Print("Continue anyway? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Operation cancelled.")
			return nil
		}
		fmt.Println()
	}

	// Convert URL (uses GitHub username for SSH host)
	newURL, err := convertToBgitURL(currentURL, activeUser.GitHubUsername)
	if err != nil {
		return err
	}

	if currentURL == newURL {
		ui.Info("Remote URL already configured for " + activeUser.Alias)
		return nil
	}

	// Update remote
	if err := setRemoteURL("origin", newURL); err != nil {
		return fmt.Errorf("failed to update remote: %w", err)
	}

	fmt.Printf("Remote 'origin' updated:\n")
	fmt.Printf("  Old: %s\n", currentURL)
	fmt.Printf("  New: %s\n", newURL)
	fmt.Println()
	ui.Success(fmt.Sprintf("Remote fixed for user '%s'", activeUser.Alias))

	return nil
}

func runRemoteRestore(cmd *cobra.Command, args []string) error {
	// Check if we're in a git repo
	if !isGitRepo() {
		return fmt.Errorf("not a git repository\nRun this command inside a git repository")
	}

	// Get current remote URL
	currentURL, err := getRemoteURL("origin")
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	if currentURL == "" {
		return fmt.Errorf("no 'origin' remote found")
	}

	// Convert back to standard GitHub URL
	newURL, err := convertToStandardURL(currentURL)
	if err != nil {
		return err
	}

	if currentURL == newURL {
		ui.Info("Remote URL is already in standard format")
		return nil
	}

	// Update remote
	if err := setRemoteURL("origin", newURL); err != nil {
		return fmt.Errorf("failed to update remote: %w", err)
	}

	fmt.Printf("Remote 'origin' restored:\n")
	fmt.Printf("  Old: %s\n", currentURL)
	fmt.Printf("  New: %s\n", newURL)
	fmt.Println()
	ui.Success("Remote restored to standard GitHub format")

	return nil
}

// isGitRepo checks if current directory is a git repository
func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// getRemoteURL gets the URL of a remote
func getRemoteURL(remote string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remote)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// setRemoteURL sets the URL of a remote
func setRemoteURL(remote, url string) error {
	cmd := exec.Command("git", "remote", "set-url", remote, url)
	return cmd.Run()
}

// convertToStandardURL converts bgit URL back to standard GitHub SSH URL
func convertToStandardURL(url string) (string, error) {
	// Pattern for bgit format: git@github.com-alias:user/repo.git
	bgitPattern := regexp.MustCompile(`^git@github\.com-[^:]+:([^/]+)/(.+?)(?:\.git)?$`)

	// Pattern for standard SSH (already standard)
	sshPattern := regexp.MustCompile(`^git@github\.com:([^/]+)/(.+?)(?:\.git)?$`)

	// Pattern for HTTPS (already standard)
	httpsPattern := regexp.MustCompile(`^https?://github\.com/`)

	if matches := bgitPattern.FindStringSubmatch(url); matches != nil {
		user := matches[1]
		repo := strings.TrimSuffix(matches[2], ".git")
		return fmt.Sprintf("git@github.com:%s/%s.git", user, repo), nil
	} else if sshPattern.MatchString(url) || httpsPattern.MatchString(url) {
		// Already in standard format
		return url, nil
	}

	return "", fmt.Errorf("unrecognized URL format: %s", url)
}

// extractAliasFromURL extracts the bgit alias from a URL if present
func extractAliasFromURL(url string) string {
	// Pattern for bgit format: git@github.com-alias:user/repo.git
	bgitPattern := regexp.MustCompile(`^git@github\.com-([^:]+):`)
	if matches := bgitPattern.FindStringSubmatch(url); matches != nil {
		return matches[1]
	}
	return ""
}
