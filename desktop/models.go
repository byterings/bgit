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

type IdentityRequest struct {
	Alias          string `json:"alias"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	GitHubUsername string `json:"githubUsername"`
	SSHKeyPath     string `json:"sshKeyPath"`
	GenerateSSHKey bool   `json:"generateSSHKey"`
}

type UpdateIdentityRequest struct {
	Alias          string `json:"alias"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	GitHubUsername string `json:"githubUsername"`
	SSHKeyPath     string `json:"sshKeyPath"`
}

type DeleteIdentityRequest struct {
	Alias      string `json:"alias"`
	DeleteKeys bool   `json:"deleteKeys"`
}

type IdentityActionResult struct {
	Message string         `json:"message"`
	Status  *DesktopStatus `json:"status"`
}

type DoctorStatus struct {
	Configured bool            `json:"configured"`
	Errors     int             `json:"errors"`
	Warnings   int             `json:"warnings"`
	Sections   []DoctorSection `json:"sections"`
}

type DoctorSection struct {
	Title  string            `json:"title"`
	Checks []DoctorCheckView `json:"checks"`
}

type DoctorCheckView struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"`
}
