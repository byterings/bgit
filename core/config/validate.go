package config

import (
	"fmt"
	"strings"
)

const (
	CurrentConfigVersion = "1.0"
	LegacyConfigVersion  = ""
)

// ValidateConfig checks that the config is structurally valid and supported.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if err := validateVersion(cfg.Version); err != nil {
		return err
	}

	aliases := make(map[string]struct{}, len(cfg.Users))
	emails := make(map[string]struct{}, len(cfg.Users))
	usernames := make(map[string]struct{}, len(cfg.Users))

	for _, user := range cfg.Users {
		if strings.TrimSpace(user.Alias) == "" {
			return fmt.Errorf("user alias is required")
		}
		if strings.TrimSpace(user.Name) == "" {
			return fmt.Errorf("user '%s' is missing name", user.Alias)
		}
		if strings.TrimSpace(user.Email) == "" {
			return fmt.Errorf("user '%s' is missing email", user.Alias)
		}
		if strings.TrimSpace(user.GitHubUsername) == "" {
			return fmt.Errorf("user '%s' is missing github_username", user.Alias)
		}

		if _, exists := aliases[user.Alias]; exists {
			return fmt.Errorf("duplicate user alias '%s'", user.Alias)
		}
		aliases[user.Alias] = struct{}{}

		if _, exists := emails[user.Email]; exists {
			return fmt.Errorf("duplicate user email '%s'", user.Email)
		}
		emails[user.Email] = struct{}{}

		if _, exists := usernames[user.GitHubUsername]; exists {
			return fmt.Errorf("duplicate GitHub username '%s'", user.GitHubUsername)
		}
		usernames[user.GitHubUsername] = struct{}{}
	}

	if cfg.ActiveUser != "" {
		if _, exists := aliases[cfg.ActiveUser]; !exists {
			return fmt.Errorf("active user '%s' not found in config", cfg.ActiveUser)
		}
	}

	workspacePaths := make(map[string]struct{}, len(cfg.Workspaces))
	for _, workspace := range cfg.Workspaces {
		if strings.TrimSpace(workspace.Path) == "" {
			return fmt.Errorf("workspace path is required")
		}
		if strings.TrimSpace(workspace.User) == "" {
			return fmt.Errorf("workspace '%s' is missing user", workspace.Path)
		}
		if _, exists := aliases[workspace.User]; !exists {
			return fmt.Errorf("workspace '%s' references unknown user '%s'", workspace.Path, workspace.User)
		}
		if _, exists := workspacePaths[workspace.Path]; exists {
			return fmt.Errorf("duplicate workspace path '%s'", workspace.Path)
		}
		workspacePaths[workspace.Path] = struct{}{}
	}

	bindingPaths := make(map[string]struct{}, len(cfg.Bindings))
	for _, binding := range cfg.Bindings {
		if strings.TrimSpace(binding.Path) == "" {
			return fmt.Errorf("binding path is required")
		}
		if strings.TrimSpace(binding.User) == "" {
			return fmt.Errorf("binding '%s' is missing user", binding.Path)
		}
		if _, exists := aliases[binding.User]; !exists {
			return fmt.Errorf("binding '%s' references unknown user '%s'", binding.Path, binding.User)
		}
		if _, exists := bindingPaths[binding.Path]; exists {
			return fmt.Errorf("duplicate binding path '%s'", binding.Path)
		}
		bindingPaths[binding.Path] = struct{}{}
	}

	return nil
}

func validateVersion(version string) error {
	switch version {
	case LegacyConfigVersion, CurrentConfigVersion:
		return nil
	default:
		return fmt.Errorf("unsupported config version '%s'", version)
	}
}
