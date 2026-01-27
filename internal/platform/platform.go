package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetSSHDir returns the SSH directory path for the current platform
func GetSSHDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".ssh"), nil
}

// GetSSHConfigPath returns the SSH config file path for the current platform
func GetSSHConfigPath() (string, error) {
	sshDir, err := GetSSHDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(sshDir, "config"), nil
}

// MkdirSecure creates a directory with appropriate permissions for the platform
func MkdirSecure(path string) error {
	if runtime.GOOS == "windows" {
		// Windows doesn't use Unix permissions
		return os.MkdirAll(path, 0755)
	}
	// Unix/Linux: use restrictive permissions
	return os.MkdirAll(path, 0700)
}

// CreateFileSecure creates a file with appropriate permissions for the platform
func CreateFileSecure(path string, data []byte) error {
	if runtime.GOOS == "windows" {
		// Windows doesn't use Unix permissions
		return os.WriteFile(path, data, 0644)
	}
	// Unix/Linux: use restrictive permissions
	return os.WriteFile(path, data, 0600)
}

// OpenFileSecure opens a file for writing with appropriate permissions
func OpenFileSecure(path string, flag int) (*os.File, error) {
	if runtime.GOOS == "windows" {
		return os.OpenFile(path, flag, 0644)
	}
	// Unix/Linux: use restrictive permissions
	return os.OpenFile(path, flag, 0600)
}

// CheckFilePermissions checks if a file has secure permissions (Unix only)
// Returns true if permissions are OK, false if they need fixing
func CheckFilePermissions(path string) (bool, error) {
	if runtime.GOOS == "windows" {
		// Windows doesn't use Unix permissions, always return true
		return true, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	mode := info.Mode()
	// Check if other users can read/write (0077)
	if mode&0077 != 0 {
		return false, nil
	}
	return true, nil
}

// FixFilePermissions sets secure permissions on a file (Unix only)
func FixFilePermissions(path string) error {
	if runtime.GOOS == "windows" {
		// Windows doesn't use Unix permissions, no-op
		return nil
	}
	return os.Chmod(path, 0600)
}

// GetPermissionFixCommand returns the appropriate command to fix file permissions
func GetPermissionFixCommand(path string) string {
	if runtime.GOOS == "windows" {
		return "File permissions are not applicable on Windows"
	}
	return fmt.Sprintf("chmod 600 %s", path)
}

// HasCommand checks if a command is available in PATH
func HasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ExpandTilde expands ~ to home directory in path
func ExpandTilde(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if len(path) == 1 {
		return home, nil
	}

	// Handle ~/rest/of/path
	if path[1] == os.PathSeparator || path[1] == '/' {
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}

// GetEditorSuggestion returns the suggested text editor command for the platform
func GetEditorSuggestion() string {
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "nano"
}

// GetSSHKeygenPath returns the path or command name for ssh-keygen
func GetSSHKeygenPath() string {
	return "ssh-keygen"
}

// NormalizePathForSSHConfig converts a path to forward slashes for SSH config
// SSH config files expect forward slashes even on Windows
func NormalizePathForSSHConfig(path string) string {
	if runtime.GOOS == "windows" {
		return filepath.ToSlash(path)
	}
	return path
}

// GetConfigDirName returns the config directory name for the platform
func GetConfigDirName() string {
	// Use .bgit for all platforms for simplicity
	// On Windows, this won't be hidden but it's consistent across platforms
	return ".bgit"
}

// GetPlatformName returns a user-friendly platform name
func GetPlatformName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return runtime.GOOS
	}
}

// GetExampleSSHKeyPath returns an example SSH key path for the platform
func GetExampleSSHKeyPath(username string) string {
	sshDir, err := GetSSHDir()
	if err != nil {
		if runtime.GOOS == "windows" {
			return fmt.Sprintf("%%USERPROFILE%%\\.ssh\\bgit_%s", username)
		}
		return fmt.Sprintf("~/.ssh/bgit_%s", username)
	}
	return filepath.Join(sshDir, fmt.Sprintf("bgit_%s", username))
}

// GetConfigFilePath returns an example config file path for the platform
func GetConfigFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		if runtime.GOOS == "windows" {
			return fmt.Sprintf("%%USERPROFILE%%\\.bgit\\config.toml")
		}
		return "~/.bgit/config.toml"
	}
	return filepath.Join(home, GetConfigDirName(), "config.toml")
}
