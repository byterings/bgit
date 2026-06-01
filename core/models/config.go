package models

// User represents a Git identity.
type User struct {
	Alias          string `toml:"alias"`
	Name           string `toml:"name"`
	Email          string `toml:"email"`
	GitHubUsername string `toml:"github_username"`
	SSHKeyPath     string `toml:"ssh_key_path"`
}

// Workspace represents a directory that auto-binds to a user identity.
type Workspace struct {
	Path string `toml:"path"`
	User string `toml:"user"`
}

// Binding represents a specific repository bound to a user identity.
type Binding struct {
	Path string `toml:"path"`
	User string `toml:"user"`
}
