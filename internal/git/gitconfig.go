package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// SetGlobalUser sets the global Git user name and email
func SetGlobalUser(name, email string) error {
	// Set user.name
	if err := runGitConfig("user.name", name); err != nil {
		return fmt.Errorf("failed to set git user.name: %w", err)
	}

	// Set user.email
	if err := runGitConfig("user.email", email); err != nil {
		return fmt.Errorf("failed to set git user.email: %w", err)
	}

	return nil
}

// GetGlobalUser returns the current global Git user name and email
func GetGlobalUser() (name, email string, err error) {
	name, err = getGitConfig("user.name")
	if err != nil {
		return "", "", fmt.Errorf("failed to get git user.name: %w", err)
	}

	email, err = getGitConfig("user.email")
	if err != nil {
		return "", "", fmt.Errorf("failed to get git user.email: %w", err)
	}

	return name, email, nil
}

// runGitConfig runs git config --global to set a value
func runGitConfig(key, value string) error {
	cmd := exec.Command("git", "config", "--global", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config failed: %s: %w", string(output), err)
	}
	return nil
}

// getGitConfig gets a git config value
func getGitConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--global", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		// If key doesn't exist, return empty string
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// IsGitInstalled checks if git is installed
func IsGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}
