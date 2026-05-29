package config

// User represents a Git identity.
type User struct {
	Alias          string `toml:"alias"` // Short name for easy switching (e.g., work, personal)
	Name           string `toml:"name"`
	Email          string `toml:"email"`
	GitHubUsername string `toml:"github_username"`
	SSHKeyPath     string `toml:"ssh_key_path"`
}

// Workspace represents a directory that auto-binds to a user identity.
// All repositories cloned within this directory will use the associated user.
type Workspace struct {
	Path string `toml:"path"` // Absolute path to the workspace directory
	User string `toml:"user"` // User alias
}

// Binding represents a specific repository bound to a user identity.
type Binding struct {
	Path string `toml:"path"` // Absolute path to the repository root
	User string `toml:"user"` // User alias
}

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
