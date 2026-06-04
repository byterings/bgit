package main

import (
	"os"

	"github.com/byterings/bgit/core/config"
	coreidentity "github.com/byterings/bgit/core/identity"
)

func loadDesktopStatus() (*DesktopStatus, error) {
	exists, err := config.ConfigExists()
	if err != nil {
		return nil, err
	}

	status := &DesktopStatus{Configured: exists}
	if !exists {
		return status, nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	status.SetupCompleted = cfg.SetupCompleted
	status.ActiveAlias = cfg.ActiveUser
	status.IdentityCount = len(cfg.Users)
	status.WorkspaceCount = len(cfg.Workspaces)
	status.BindingCount = len(cfg.Bindings)

	for _, user := range cfg.Users {
		view := identityViewFromUser(user, cfg.ActiveUser)
		status.Identities = append(status.Identities, view)
		if user.Alias == cfg.ActiveUser {
			active := view
			status.ActiveIdentity = &active
		}
	}

	resolution, err := coreidentity.GetEffectiveResolution(cfg)
	if err != nil {
		return nil, err
	}
	if resolution != nil {
		status.EffectiveIdentity = &EffectiveIdentityView{
			Alias:  resolution.Alias,
			Source: string(resolution.Source),
			Path:   resolution.Path,
		}
	}

	return status, nil
}

func identityViewFromUser(user config.User, activeAlias string) IdentityView {
	view := IdentityView{
		Alias:          user.Alias,
		Name:           user.Name,
		Email:          user.Email,
		GitHubUsername: user.GitHubUsername,
		SSHKeyPath:     user.SSHKeyPath,
		Active:         user.Alias == activeAlias,
	}

	if user.SSHKeyPath == "" {
		view.SSHKeyStatus = "not_configured"
		return view
	}
	if _, err := os.Stat(user.SSHKeyPath); err != nil {
		if os.IsNotExist(err) {
			view.SSHKeyStatus = "missing"
		} else {
			view.SSHKeyStatus = "unknown"
		}
		return view
	}

	view.SSHKeyStatus = "available"
	return view
}
