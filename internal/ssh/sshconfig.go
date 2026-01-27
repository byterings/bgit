package ssh

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/platform"
)

const (
	bgitManagedStart = "# ---- BEGIN BGIT MANAGED ----"
	bgitManagedEnd   = "# ---- END BGIT MANAGED ----"
	// Legacy markers for migration from bgit
	legacyManagedStart = "# ---- BEGIN BRGIT MANAGED ----"
	legacyManagedEnd   = "# ---- END BRGIT MANAGED ----"
)

// GetSSHConfigPath returns the path to the SSH config file
func GetSSHConfigPath() (string, error) {
	return platform.GetSSHConfigPath()
}

// UpdateSSHConfig updates the SSH config with bgit-managed entries
func UpdateSSHConfig(users []config.User) error {
	configPath, err := GetSSHConfigPath()
	if err != nil {
		return err
	}

	// Ensure .ssh directory exists
	sshDir := filepath.Dir(configPath)
	if err := platform.MkdirSecure(sshDir); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Read existing config
	existingContent, err := readSSHConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read SSH config: %w", err)
	}

	// Remove old bgit-managed section
	cleanedContent := removeBgitSection(existingContent)

	// Generate new bgit section
	bgitSection := generateBgitSection(users)

	// Combine content
	var newContent strings.Builder
	if cleanedContent != "" {
		newContent.WriteString(cleanedContent)
		if !strings.HasSuffix(cleanedContent, "\n") {
			newContent.WriteString("\n")
		}
		newContent.WriteString("\n")
	}
	newContent.WriteString(bgitSection)

	// Write updated config
	if err := platform.CreateFileSecure(configPath, []byte(newContent.String())); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}

	return nil
}

// readSSHConfig reads the SSH config file
func readSSHConfig(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// removeBgitSection removes the bgit-managed section from SSH config
// Also removes legacy bgit-managed sections for migration
func removeBgitSection(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var result strings.Builder
	inManagedSection := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for current or legacy start markers
		if trimmedLine == bgitManagedStart || trimmedLine == legacyManagedStart {
			inManagedSection = true
			continue
		}

		// Check for current or legacy end markers
		if trimmedLine == bgitManagedEnd || trimmedLine == legacyManagedEnd {
			inManagedSection = false
			continue
		}

		if !inManagedSection {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimRight(result.String(), "\n")
}

// generateBgitSection generates the bgit-managed SSH config section
func generateBgitSection(users []config.User) string {
	var section strings.Builder

	section.WriteString(bgitManagedStart + "\n")
	section.WriteString("# DO NOT EDIT THIS SECTION MANUALLY\n")
	section.WriteString("# This section is managed by bgit\n")
	section.WriteString("\n")

	for _, user := range users {
		if user.SSHKeyPath == "" {
			continue // Skip users without SSH keys
		}

		section.WriteString(fmt.Sprintf("Host github.com-%s\n", user.GitHubUsername))
		section.WriteString("  HostName github.com\n")
		section.WriteString("  User git\n")
		section.WriteString(fmt.Sprintf("  IdentityFile %s\n", platform.NormalizePathForSSHConfig(user.SSHKeyPath)))
		section.WriteString("  IdentitiesOnly yes\n")
		section.WriteString("\n")
	}

	section.WriteString(bgitManagedEnd + "\n")

	return section.String()
}

// GetHostForUser returns the SSH host alias for a user
func GetHostForUser(username string) string {
	return fmt.Sprintf("github.com-%s", username)
}
