package ssh

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/internal/platform"
)

const (
	bgitManagedStart   = "# ---- BEGIN BGIT MANAGED ----"
	bgitManagedEnd     = "# ---- END BGIT MANAGED ----"
	legacyManagedStart = "# ---- BEGIN BRGIT MANAGED ----"
	legacyManagedEnd   = "# ---- END BRGIT MANAGED ----"
)

// GetSSHConfigPath returns the path to the SSH config file.
func GetSSHConfigPath() (string, error) {
	return platform.GetSSHConfigPath()
}

// UpdateSSHConfig updates the SSH config with bgit-managed entries.
func UpdateSSHConfig(users []config.User) error {
	configPath, err := GetSSHConfigPath()
	if err != nil {
		return err
	}

	sshDir := filepath.Dir(configPath)
	if err := platform.MkdirSecure(sshDir); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	existingContent, err := readSSHConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read SSH config: %w", err)
	}

	cleanedContent := removeManagedSection(existingContent)
	bgitSection := generateManagedSection(users)

	var newContent strings.Builder
	if cleanedContent != "" {
		newContent.WriteString(cleanedContent)
		if !strings.HasSuffix(cleanedContent, "\n") {
			newContent.WriteString("\n")
		}
		newContent.WriteString("\n")
	}
	newContent.WriteString(bgitSection)

	if err := platform.CreateFileSecure(configPath, []byte(newContent.String())); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}

	return nil
}

// HasManagedSection reports whether the content contains bgit-managed SSH entries.
func HasManagedSection(content string) bool {
	return strings.Contains(content, bgitManagedStart) || strings.Contains(content, legacyManagedStart)
}

// RemoveManagedSectionFromFile removes the bgit-managed SSH block from the config file if present.
func RemoveManagedSectionFromFile() error {
	configPath, err := GetSSHConfigPath()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cleaned := removeManagedSection(string(content))
	if cleaned != "" {
		cleaned += "\n"
	}

	return os.WriteFile(configPath, []byte(cleaned), 0600)
}

// GetHostForUser returns the SSH host alias for a user.
func GetHostForUser(username string) string {
	return fmt.Sprintf("github.com-%s", username)
}

func readSSHConfig(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func removeManagedSection(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var result strings.Builder
	inManagedSection := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == bgitManagedStart || trimmedLine == legacyManagedStart {
			inManagedSection = true
			continue
		}

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

func generateManagedSection(users []config.User) string {
	var section strings.Builder

	section.WriteString(bgitManagedStart + "\n")
	section.WriteString("# DO NOT EDIT THIS SECTION MANUALLY\n")
	section.WriteString("# This section is managed by bgit\n\n")

	for _, user := range users {
		if user.SSHKeyPath == "" {
			continue
		}

		section.WriteString(fmt.Sprintf("Host %s\n", GetHostForUser(user.GitHubUsername)))
		section.WriteString("  HostName github.com\n")
		section.WriteString("  User git\n")
		section.WriteString(fmt.Sprintf("  IdentityFile %s\n", platform.NormalizePathForSSHConfig(user.SSHKeyPath)))
		section.WriteString("  IdentitiesOnly yes\n\n")
	}

	section.WriteString(bgitManagedEnd + "\n")
	return section.String()
}
