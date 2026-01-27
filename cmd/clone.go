package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/identity"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <url> [directory]",
	Short: "Clone a repository with the correct SSH configuration",
	Long: `Clone a GitHub repository using the active user's SSH configuration.

Accepts any GitHub URL format (HTTPS or SSH) and automatically converts it
to use the correct SSH host alias for the active user.`,
	Example: `  # Clone using HTTPS URL
  bgit clone https://github.com/user/repo.git

  # Clone using SSH URL
  bgit clone git@github.com:user/repo.git

  # Clone to specific directory
  bgit clone https://github.com/user/repo.git my-folder`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	url := args[0]
	var directory string
	if len(args) > 1 {
		directory = args[1]
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

	// Resolve effective identity (workspace > binding > global)
	resolution, err := identity.GetEffectiveResolution(cfg)
	if err != nil || resolution == nil || resolution.User == nil {
		// Fall back to checking global active user
		if cfg.ActiveUser == "" {
			return fmt.Errorf("no active user set\nRun: bgit use <alias>")
		}
		resolution = &identity.Resolution{
			User:   cfg.FindUserByAlias(cfg.ActiveUser),
			Alias:  cfg.ActiveUser,
			Source: identity.SourceGlobal,
		}
		if resolution.User == nil {
			return fmt.Errorf("active user '%s' not found in config", cfg.ActiveUser)
		}
	}

	activeUser := resolution.User

	// Show identity source if not global
	if resolution.Source != identity.SourceGlobal {
		sourceInfo := ""
		switch resolution.Source {
		case identity.SourceWorkspace:
			sourceInfo = fmt.Sprintf(" (workspace: %s)", resolution.Path)
		case identity.SourceBinding:
			sourceInfo = " (bound repo)"
		}
		ui.Info(fmt.Sprintf("Using identity from %s%s", resolution.Source, sourceInfo))
	}

	// Check if SSH key is configured
	if activeUser.SSHKeyPath == "" {
		ui.Warning("No SSH key configured for this user")
		fmt.Println("Clone may fail. Run: bgit update " + activeUser.Alias + " --ssh-key <path>")
		fmt.Println()
	} else {
		// Ensure SSH agent has the key loaded
		ensureSSHAgentForClone(activeUser)
	}

	// Convert URL to bgit format (uses GitHub username for SSH host)
	convertedURL, err := convertToBgitURL(url, activeUser.GitHubUsername)
	if err != nil {
		return err
	}

	fmt.Printf("Cloning as: %s\n", activeUser.Alias)
	fmt.Printf("URL: %s\n\n", convertedURL)

	// Build git clone command
	gitArgs := []string{"clone", convertedURL}
	if directory != "" {
		gitArgs = append(gitArgs, directory)
	}

	// Execute git clone
	gitCmd := exec.Command("git", gitArgs...)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	gitCmd.Stdin = os.Stdin

	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	fmt.Println()
	ui.Success("Repository cloned successfully!")

	return nil
}

// ensureSSHAgentForClone ensures SSH key is loaded for cloning
func ensureSSHAgentForClone(user *config.User) {
	if runtime.GOOS == "windows" {
		// Start ssh-agent service silently
		startCmd := exec.Command("powershell", "-Command", "Start-Service ssh-agent")
		startCmd.Run()

		// Set to automatic startup
		autoCmd := exec.Command("powershell", "-Command", "Set-Service -Name ssh-agent -StartupType Automatic")
		autoCmd.Run()
	}

	// Check if key is already loaded
	listCmd := exec.Command("ssh-add", "-l")
	output, _ := listCmd.Output()

	// If key not in agent, add it
	if user.SSHKeyPath != "" && !strings.Contains(string(output), user.SSHKeyPath) {
		addCmd := exec.Command("ssh-add", user.SSHKeyPath)
		addCmd.Run()
	}
}

// convertToBgitURL converts any GitHub URL to bgit's SSH format
// sshHostUser is the GitHub username used for the SSH host (github.com-<sshHostUser>)
func convertToBgitURL(url string, sshHostUser string) (string, error) {
	// Pattern for HTTPS: https://github.com/user/repo.git
	httpsPattern := regexp.MustCompile(`^https?://github\.com/([^/]+)/(.+?)(?:\.git)?$`)

	// Pattern for SSH: git@github.com:user/repo.git
	sshPattern := regexp.MustCompile(`^git@github\.com:([^/]+)/(.+?)(?:\.git)?$`)

	// Pattern for already converted: git@github.com-user:user/repo.git
	bgitPattern := regexp.MustCompile(`^git@github\.com-([^:]+):([^/]+)/(.+?)(?:\.git)?$`)

	var repoOwner, repoName string

	if matches := httpsPattern.FindStringSubmatch(url); matches != nil {
		repoOwner = matches[1]
		repoName = matches[2]
	} else if matches := sshPattern.FindStringSubmatch(url); matches != nil {
		repoOwner = matches[1]
		repoName = matches[2]
	} else if matches := bgitPattern.FindStringSubmatch(url); matches != nil {
		// Already in bgit format, update host user if different
		repoOwner = matches[2]
		repoName = matches[3]
	} else {
		return "", fmt.Errorf("unrecognized URL format: %s\nExpected GitHub HTTPS or SSH URL", url)
	}

	// Remove .git suffix if present
	repoName = strings.TrimSuffix(repoName, ".git")

	// sshHostUser is the GitHub username that matches SSH config: Host github.com-<sshHostUser>
	return fmt.Sprintf("git@github.com-%s:%s/%s.git", sshHostUser, repoOwner, repoName), nil
}
