package config

import "github.com/byterings/bgit/core/models"

type User = models.User
type Workspace = models.Workspace
type Binding = models.Binding

// Config represents the bgit configuration.
type Config struct {
	Version                string      `toml:"version"`
	SetupCompleted         bool        `toml:"setup_completed"`
	ActiveUser             string      `toml:"active_user"` // Stores the alias
	PreviousGitIdentitySet bool        `toml:"previous_git_identity_set"`
	PreviousGitName        string      `toml:"previous_git_name"`
	PreviousGitEmail       string      `toml:"previous_git_email"`
	PreviousHooksPathSet   bool        `toml:"previous_hooks_path_set"`
	PreviousHooksPath      string      `toml:"previous_hooks_path"`
	Users                  []User      `toml:"users"`
	Workspaces             []Workspace `toml:"workspaces"`
	Bindings               []Binding   `toml:"bindings"`
}
