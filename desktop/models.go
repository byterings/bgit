package main

type DesktopStatus struct {
	Configured        bool                   `json:"configured"`
	SetupCompleted    bool                   `json:"setupCompleted"`
	ActiveAlias       string                 `json:"activeAlias"`
	IdentityCount     int                    `json:"identityCount"`
	WorkspaceCount    int                    `json:"workspaceCount"`
	BindingCount      int                    `json:"bindingCount"`
	Identities        []IdentityView         `json:"identities"`
	Bindings          []RepoBindingView      `json:"bindings"`
	ActiveIdentity    *IdentityView          `json:"activeIdentity,omitempty"`
	EffectiveIdentity *EffectiveIdentityView `json:"effectiveIdentity,omitempty"`
}

type IdentityView struct {
	Alias              string `json:"alias"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	GitHubUsername     string `json:"githubUsername"`
	GitHubAvatarURL    string `json:"githubAvatarUrl"`
	SSHKeyPath         string `json:"sshKeyPath"`
	SSHPublicKey       string `json:"sshPublicKey,omitempty"`
	SSHPublicKeyStatus string `json:"sshPublicKeyStatus"`
	SSHKeyStatus       string `json:"sshKeyStatus"`
	Active             bool   `json:"active"`
}

type EffectiveIdentityView struct {
	Alias  string `json:"alias"`
	Source string `json:"source"`
	Path   string `json:"path,omitempty"`
}

type RepoBindingView struct {
	Path            string `json:"path"`
	Alias           string `json:"alias"`
	Name            string `json:"name,omitempty"`
	Email           string `json:"email,omitempty"`
	GitHubUsername  string `json:"githubUsername,omitempty"`
	MissingIdentity bool   `json:"missingIdentity"`
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

type RepoBindingRequest struct {
	Path  string `json:"path"`
	Alias string `json:"alias"`
}

type RepoBindingActionResult struct {
	Message string         `json:"message"`
	Path    string         `json:"path,omitempty"`
	Status  *DesktopStatus `json:"status"`
}

type BackupExportRequest struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	OutputPath      string `json:"outputPath"`
}

type BackupImportRequest struct {
	Password    string `json:"password"`
	ArchivePath string `json:"archivePath"`
}

type BackupActionResult struct {
	Message         string         `json:"message"`
	ArchivePath     string         `json:"archivePath,omitempty"`
	UsersCount      int            `json:"usersCount,omitempty"`
	WorkspacesCount int            `json:"workspacesCount,omitempty"`
	BindingsCount   int            `json:"bindingsCount,omitempty"`
	ActiveUser      string         `json:"activeUser,omitempty"`
	Status          *DesktopStatus `json:"status,omitempty"`
	Doctor          *DoctorStatus  `json:"doctor,omitempty"`
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
