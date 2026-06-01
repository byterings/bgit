package identity

import (
	"fmt"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/git"
)

// LookupMode controls how a user is selected for activation.
type LookupMode int

const (
	LookupByAlias LookupMode = iota
	LookupByUsername
	LookupByEmail
	LookupByAny
)

type DeleteResult = models.DeleteResult
type ActivateResult = models.ActivateResult
type UserOperationResult = models.UserOperationResult

// AddUser adds the given identity and persists the config.
func AddUser(cfg *config.Config, user config.User) (*UserOperationResult, error) {
	if err := cfg.AddUser(user); err != nil {
		return nil, err
	}
	if err := config.SaveConfig(cfg); err != nil {
		return nil, err
	}
	return &UserOperationResult{
		User:    cfg.FindUserByAlias(user.Alias),
		Changed: true,
	}, nil
}

// UpdateUserSSHKey updates a user's SSH key and persists the config and SSH config.
func UpdateUserSSHKey(cfg *config.Config, identifier, sshKeyPath string) (*UserOperationResult, error) {
	foundUser := cfg.FindUser(identifier)
	if foundUser == nil {
		return nil, fmt.Errorf("user '%s' not found", identifier)
	}

	for i := range cfg.Users {
		if cfg.Users[i].Alias == foundUser.Alias {
			cfg.Users[i].SSHKeyPath = sshKeyPath
			break
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		return nil, err
	}
	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return nil, err
	}

	return &UserOperationResult{
		User:    cfg.FindUserByAlias(foundUser.Alias),
		Changed: true,
	}, nil
}

// DeleteUser removes a user from config and persists the change.
func DeleteUser(cfg *config.Config, identifier string) (*DeleteResult, error) {
	user := cfg.FindUser(identifier)
	if user == nil {
		return nil, fmt.Errorf("user '%s' not found", identifier)
	}

	result := &DeleteResult{
		User: *user,
	}

	newUsers := make([]config.User, 0, len(cfg.Users))
	for _, existingUser := range cfg.Users {
		if existingUser.Alias != user.Alias {
			newUsers = append(newUsers, existingUser)
		}
	}
	cfg.Users = newUsers

	if cfg.ActiveUser == user.Alias {
		cfg.ActiveUser = ""
		result.ActiveCleared = true
	}

	if err := config.SaveConfig(cfg); err != nil {
		return nil, err
	}
	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return nil, err
	}

	return result, nil
}

// ActivateUser switches the global Git identity and active bgit identity.
func ActivateUser(cfg *config.Config, identifier string, mode LookupMode) (*ActivateResult, error) {
	user := resolveUser(cfg, identifier, mode)
	if user == nil {
		return nil, fmt.Errorf("user '%s' not found", identifier)
	}

	if err := capturePreviousGitIdentity(cfg); err != nil {
		return nil, err
	}
	if err := git.SetGlobalUser(user.Name, user.Email); err != nil {
		return nil, err
	}
	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return nil, err
	}

	cfg.ActiveUser = user.Alias
	if err := config.SaveConfig(cfg); err != nil {
		return nil, err
	}

	return &ActivateResult{
		User:      user,
		KeyLoaded: coressh.EnsureKeyLoaded(user).Changed,
	}, nil
}

func resolveUser(cfg *config.Config, identifier string, mode LookupMode) *config.User {
	switch mode {
	case LookupByUsername:
		return cfg.FindUserByUsername(identifier)
	case LookupByEmail:
		return cfg.FindUserByEmail(identifier)
	case LookupByAlias:
		return cfg.FindUserByAlias(identifier)
	default:
		return cfg.FindUser(identifier)
	}
}

func capturePreviousGitIdentity(cfg *config.Config) error {
	if cfg.PreviousGitIdentitySet || cfg.ActiveUser != "" {
		return nil
	}

	name, email, err := git.GetGlobalUser()
	if err != nil {
		return err
	}

	cfg.PreviousGitName = name
	cfg.PreviousGitEmail = email
	cfg.PreviousGitIdentitySet = true
	return nil
}
