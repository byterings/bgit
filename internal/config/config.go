package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/byterings/bgit/internal/platform"
)

const (
	ConfigFileName    = "config.toml"
	BackupDirName     = "backups"
	LegacyConfigDir   = ".bgit" // Old config directory name for migration
)

// GetConfigDirName returns the config directory name
func GetConfigDirName() string {
	return platform.GetConfigDirName()
}

// GetConfigDir returns the path to the bgit config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, GetConfigDirName()), nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFileName), nil
}

// GetBackupDir returns the path to the backup directory
func GetBackupDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, BackupDirName), nil
}

// ConfigExists checks if the config file exists
// It also attempts migration from legacy bgit config if needed
func ConfigExists() (bool, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(configPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		// Try migrating from legacy bgit config
		migrated, migrateErr := MigrateFromLegacy()
		if migrateErr != nil {
			// Log but don't fail - migration is optional
			fmt.Fprintf(os.Stderr, "Warning: migration from bgit failed: %v\n", migrateErr)
		}
		if migrated {
			// Check again after migration
			_, err = os.Stat(configPath)
			if err == nil {
				return true, nil
			}
		}
		return false, nil
	}
	return false, err
}

// MigrateFromLegacy migrates configuration from the legacy ~/.bgit directory
// to the new ~/.bgit directory. Returns true if migration was performed.
func MigrateFromLegacy() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	oldDir := filepath.Join(home, LegacyConfigDir)
	newDir := filepath.Join(home, GetConfigDirName())

	// Check if old config exists
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return false, nil // No legacy config to migrate
	}

	// Check if new config already exists
	if _, err := os.Stat(newDir); err == nil {
		return false, nil // New config already exists, don't overwrite
	}

	// Perform migration by copying the directory
	fmt.Printf("Migrating configuration from %s to %s...\n", oldDir, newDir)

	if err := copyDir(oldDir, newDir); err != nil {
		return false, fmt.Errorf("failed to migrate config directory: %w", err)
	}

	fmt.Println("Migration complete! Your bgit configuration has been migrated to bgit.")
	fmt.Println("Note: Your existing SSH keys (bgit_*) will continue to work.")
	fmt.Println("      New keys will be created with the bgit_* prefix.")

	return true, nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Create destination directory with secure permissions
	if err := platform.MkdirSecure(dst); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file with secure permissions
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := platform.OpenFileSecure(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}

// CreateConfigDir creates the bgit config directory
func CreateConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	return platform.MkdirSecure(configDir)
}

// CreateBackupDir creates the backup directory
func CreateBackupDir() error {
	backupDir, err := GetBackupDir()
	if err != nil {
		return err
	}
	return platform.MkdirSecure(backupDir)
}

// NewConfig creates a new empty config
func NewConfig() *Config {
	return &Config{
		Version:    "1.0",
		ActiveUser: "",
		Users:      []User{},
	}
}

// LoadConfig loads the config from file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Migration: Set alias to GitHub username if missing
	needsSave := false
	for i := range config.Users {
		if config.Users[i].Alias == "" {
			config.Users[i].Alias = config.Users[i].GitHubUsername
			needsSave = true
		}
	}

	// Migration: Update ActiveUser if it's a GitHub username instead of alias
	if config.ActiveUser != "" {
		// Check if ActiveUser is actually a GitHub username
		user := config.FindUserByUsername(config.ActiveUser)
		if user != nil && user.Alias != "" {
			config.ActiveUser = user.Alias
			needsSave = true
		}
	}

	// Save migrated config
	if needsSave {
		if err := SaveConfig(&config); err != nil {
			return nil, fmt.Errorf("failed to save migrated config: %w", err)
		}
	}

	return &config, nil
}

// SaveConfig saves the config to file
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	f, err := platform.OpenFileSecure(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// FindUser finds a user by alias (primary), GitHub username, or email
func (c *Config) FindUser(identifier string) *User {
	for i := range c.Users {
		if c.Users[i].Alias == identifier ||
			c.Users[i].GitHubUsername == identifier ||
			c.Users[i].Email == identifier {
			return &c.Users[i]
		}
	}
	return nil
}

// FindUserByAlias finds a user by alias only
func (c *Config) FindUserByAlias(alias string) *User {
	for i := range c.Users {
		if c.Users[i].Alias == alias {
			return &c.Users[i]
		}
	}
	return nil
}

// FindUserByUsername finds a user by GitHub username only
func (c *Config) FindUserByUsername(username string) *User {
	for i := range c.Users {
		if c.Users[i].GitHubUsername == username {
			return &c.Users[i]
		}
	}
	return nil
}

// FindUserByEmail finds a user by email only
func (c *Config) FindUserByEmail(email string) *User {
	for i := range c.Users {
		if c.Users[i].Email == email {
			return &c.Users[i]
		}
	}
	return nil
}

// AddUser adds a new user to the config
func (c *Config) AddUser(user User) error {
	// Check for uniqueness
	for _, u := range c.Users {
		if u.Alias == user.Alias {
			return fmt.Errorf("user with alias '%s' already exists", user.Alias)
		}
		if u.Email == user.Email {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		if u.GitHubUsername == user.GitHubUsername {
			return fmt.Errorf("user with GitHub username %s already exists", user.GitHubUsername)
		}
	}
	c.Users = append(c.Users, user)
	return nil
}

// AddWorkspace adds a new workspace to the config
func (c *Config) AddWorkspace(path, userAlias string) error {
	// Check if workspace already exists
	for _, ws := range c.Workspaces {
		if ws.Path == path {
			return fmt.Errorf("workspace at '%s' already exists", path)
		}
	}
	// Verify user exists
	if c.FindUserByAlias(userAlias) == nil {
		return fmt.Errorf("user '%s' not found", userAlias)
	}
	c.Workspaces = append(c.Workspaces, Workspace{Path: path, User: userAlias})
	return nil
}

// RemoveWorkspace removes a workspace by user alias
func (c *Config) RemoveWorkspace(userAlias string) bool {
	for i, ws := range c.Workspaces {
		if ws.User == userAlias {
			c.Workspaces = append(c.Workspaces[:i], c.Workspaces[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveWorkspaceByPath removes a workspace by path
func (c *Config) RemoveWorkspaceByPath(path string) bool {
	for i, ws := range c.Workspaces {
		if ws.Path == path {
			c.Workspaces = append(c.Workspaces[:i], c.Workspaces[i+1:]...)
			return true
		}
	}
	return false
}

// GetWorkspaces returns all configured workspaces
func (c *Config) GetWorkspaces() []Workspace {
	return c.Workspaces
}

// FindWorkspaceByPath finds a workspace that contains the given path
func (c *Config) FindWorkspaceByPath(path string) *Workspace {
	for i, ws := range c.Workspaces {
		if isPathInside(path, ws.Path) {
			return &c.Workspaces[i]
		}
	}
	return nil
}

// AddBinding adds a new repo binding to the config
func (c *Config) AddBinding(path, userAlias string) error {
	// Check if binding already exists, update if so
	for i, b := range c.Bindings {
		if b.Path == path {
			c.Bindings[i].User = userAlias
			return nil
		}
	}
	// Verify user exists
	if c.FindUserByAlias(userAlias) == nil {
		return fmt.Errorf("user '%s' not found", userAlias)
	}
	c.Bindings = append(c.Bindings, Binding{Path: path, User: userAlias})
	return nil
}

// RemoveBinding removes a binding by path
func (c *Config) RemoveBinding(path string) bool {
	for i, b := range c.Bindings {
		if b.Path == path {
			c.Bindings = append(c.Bindings[:i], c.Bindings[i+1:]...)
			return true
		}
	}
	return false
}

// GetBindings returns all configured bindings
func (c *Config) GetBindings() []Binding {
	return c.Bindings
}

// FindBindingByPath finds a binding for the given path
func (c *Config) FindBindingByPath(path string) *Binding {
	for i, b := range c.Bindings {
		if b.Path == path {
			return &c.Bindings[i]
		}
	}
	return nil
}

// CleanupInvalidPaths removes workspaces and bindings for non-existent paths
func (c *Config) CleanupInvalidPaths() int {
	removed := 0

	// Clean workspaces
	validWorkspaces := make([]Workspace, 0, len(c.Workspaces))
	for _, ws := range c.Workspaces {
		if _, err := os.Stat(ws.Path); err == nil {
			validWorkspaces = append(validWorkspaces, ws)
		} else {
			removed++
		}
	}
	c.Workspaces = validWorkspaces

	// Clean bindings
	validBindings := make([]Binding, 0, len(c.Bindings))
	for _, b := range c.Bindings {
		if _, err := os.Stat(b.Path); err == nil {
			validBindings = append(validBindings, b)
		} else {
			removed++
		}
	}
	c.Bindings = validBindings

	return removed
}

// isPathInside checks if childPath is inside parentPath
func isPathInside(childPath, parentPath string) bool {
	// Clean and get absolute paths
	child, err := filepath.Abs(childPath)
	if err != nil {
		return false
	}
	parent, err := filepath.Abs(parentPath)
	if err != nil {
		return false
	}

	// Ensure parent ends with separator for proper prefix matching
	if !strings.HasSuffix(parent, string(filepath.Separator)) {
		parent = parent + string(filepath.Separator)
	}

	return strings.HasPrefix(child+string(filepath.Separator), parent) || child == strings.TrimSuffix(parent, string(filepath.Separator))
}
