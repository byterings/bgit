package main

type DesktopStatus struct {
	Configured        bool                   `json:"configured"`
	SetupCompleted    bool                   `json:"setupCompleted"`
	ActiveAlias       string                 `json:"activeAlias"`
	IdentityCount     int                    `json:"identityCount"`
	WorkspaceCount    int                    `json:"workspaceCount"`
	BindingCount      int                    `json:"bindingCount"`
	Identities        []IdentityView         `json:"identities"`
	ActiveIdentity    *IdentityView          `json:"activeIdentity,omitempty"`
	EffectiveIdentity *EffectiveIdentityView `json:"effectiveIdentity,omitempty"`
}

type IdentityView struct {
	Alias          string `json:"alias"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	GitHubUsername string `json:"githubUsername"`
	SSHKeyPath     string `json:"sshKeyPath"`
	SSHKeyStatus   string `json:"sshKeyStatus"`
	Active         bool   `json:"active"`
}

type EffectiveIdentityView struct {
	Alias  string `json:"alias"`
	Source string `json:"source"`
	Path   string `json:"path,omitempty"`
}
