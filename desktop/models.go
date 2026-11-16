package main

import "github.com/byterings/bgit/core/models"

type DesktopStatus struct {
	Configured        bool                   `json:"configured"`
	SetupCompleted    bool                   `json:"setupCompleted"`
	ActiveAlias       string                 `json:"activeAlias"`
	IdentityCount     int                    `json:"identityCount"`
	WorkspaceCount    int                    `json:"workspaceCount"`
	BindingCount      int                    `json:"bindingCount"`
	Identities        []IdentityView         `json:"identities"`
	Bindings          []RepoBindingView      `json:"bindings"`
	Workspaces        []WorkspaceView        `json:"workspaces"`
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

type IdentityDetectionRequest struct {
	Path string `json:"path"`
}

type IdentityDetectionResult struct {
	Path             string `json:"path"`
	RepoRoot         string `json:"repoRoot,omitempty"`
	Alias            string `json:"alias,omitempty"`
	Source           string `json:"source,omitempty"`
	SourcePath       string `json:"sourcePath,omitempty"`
	Name             string `json:"name,omitempty"`
	Email            string `json:"email,omitempty"`
	GitHubUsername   string `json:"githubUsername,omitempty"`
	GitConfigScope   string `json:"gitConfigScope,omitempty"`
	GitName          string `json:"gitName,omitempty"`
	GitEmail         string `json:"gitEmail,omitempty"`
	GitMatches       bool   `json:"gitMatches"`
	GitConfigChecked bool   `json:"gitConfigChecked"`
	SyncAvailable    bool   `json:"syncAvailable"`
	Message          string `json:"message,omitempty"`
}

type IdentityDetectionActionResult struct {
	Message  string                   `json:"message"`
	Detected *IdentityDetectionResult `json:"detected,omitempty"`
	Status   *DesktopStatus           `json:"status,omitempty"`
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

// OnboardingState describes what first-run setup still has to do.
type OnboardingState struct {
	GitInstalled     bool   `json:"gitInstalled"`
	GitVersion       string `json:"gitVersion,omitempty"`
	Configured       bool   `json:"configured"`
	SetupCompleted   bool   `json:"setupCompleted"`
	PushProtection   bool   `json:"pushProtection"`
	HooksPath        string `json:"hooksPath,omitempty"`
	HooksPathManaged bool   `json:"hooksPathManaged"`
	IdentityCount    int    `json:"identityCount"`
	ActiveAlias      string `json:"activeAlias,omitempty"`
	// Needed is true while the wizard still has a required step left.
	Needed bool `json:"needed"`
}

// ConnectionTestResult is the outcome of an SSH connection test to GitHub.
type ConnectionTestResult struct {
	Alias          string `json:"alias"`
	GitHubUsername string `json:"githubUsername"`
	Passed         bool   `json:"passed"`
	Authenticated  string `json:"authenticated,omitempty"`
	Message        string `json:"message"`
	Fix            string `json:"fix,omitempty"`
}

// AppInfo carries build and install details for the Settings page.
type AppInfo struct {
	Version        string `json:"version"`
	HookHelperPath string `json:"hookHelperPath,omitempty"`
}

// RepositoryView is one repository bgit manages, with its identity and health.
type RepositoryView struct {
	Path string `json:"path"`
	Name string `json:"name"`
	// Source is "binding" for a bound repository or "workspace" for one found
	// inside a workspace folder.
	Source     string `json:"source"`
	SourcePath string `json:"sourcePath,omitempty"`

	Alias string `json:"alias"`
	// IdentitySource is how the identity was chosen: binding, workspace,
	// remote, or active.
	IdentitySource string `json:"identitySource,omitempty"`
	IdentityName   string `json:"identityName,omitempty"`
	IdentityEmail  string `json:"identityEmail,omitempty"`
	GitHubUsername string `json:"githubUsername,omitempty"`

	Exists bool `json:"exists"`
	// Health is "ok", "attention" (fixable), "blocked" (not fixable in the
	// app), or "missing".
	Health   string   `json:"health"`
	Problems []string `json:"problems"`
	Notes    []string `json:"notes"`
	CanFix   bool     `json:"canFix"`

	RemoteURL string `json:"remoteUrl,omitempty"`
	Branch    string `json:"branch,omitempty"`
	Changes   int    `json:"changes"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
}

type RepositoriesResult struct {
	Repositories  []RepositoryView `json:"repositories"`
	NeedAttention int              `json:"needAttention"`
	Message       string           `json:"message,omitempty"`
}

type RepositoryActionResult struct {
	Message    string          `json:"message"`
	Repository *RepositoryView `json:"repository,omitempty"`
	Status     *DesktopStatus  `json:"status,omitempty"`
}

type CloneRequest struct {
	URL       string `json:"url"`
	ParentDir string `json:"parentDir"`
	// FolderName overrides the folder created inside ParentDir. Empty uses
	// the repository name.
	FolderName string `json:"folderName"`
	Alias      string `json:"alias"`
}

type CloneDefaults struct {
	ParentDir     string `json:"parentDir"`
	FromWorkspace bool   `json:"fromWorkspace"`
}

// RepoState is everything the repository view needs after an action.
type RepoState struct {
	Repository *RepositoryView   `json:"repository"`
	Status     *models.GitStatus `json:"status"`
}

// GitActionResult is the outcome of a Git operation. Failures Git itself
// reported come back with OK=false and a Kind the UI can act on.
type GitActionResult struct {
	OK      bool       `json:"ok"`
	Kind    string     `json:"kind,omitempty"`
	Message string     `json:"message"`
	Detail  string     `json:"detail,omitempty"`
	State   *RepoState `json:"state,omitempty"`
}

type FileDiffRequest struct {
	Path string `json:"path"`
	File string `json:"file"`
	// Commit shows the change made by that commit instead of the
	// uncommitted change.
	Commit string `json:"commit"`
}

type CommitChangesRequest struct {
	Path        string   `json:"path"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
}

type DiscardChangesRequest struct {
	Path  string   `json:"path"`
	Files []string `json:"files"`
}

type BranchRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type HistoryRequest struct {
	Path  string `json:"path"`
	Limit int    `json:"limit"`
	Skip  int    `json:"skip"`
}

type CommitFilesRequest struct {
	Path   string `json:"path"`
	Commit string `json:"commit"`
}

type WorkspaceView struct {
	Path            string `json:"path"`
	Alias           string `json:"alias"`
	IdentityName    string `json:"identityName,omitempty"`
	IdentityEmail   string `json:"identityEmail,omitempty"`
	MissingIdentity bool   `json:"missingIdentity"`
	Exists          bool   `json:"exists"`
	RepositoryCount int    `json:"repositoryCount"`
}

type WorkspaceRequest struct {
	Path  string `json:"path"`
	Alias string `json:"alias"`
}

type WorkspaceActionResult struct {
	Message string         `json:"message"`
	Status  *DesktopStatus `json:"status,omitempty"`
}

type UninstallRequest struct {
	// RemoveKeys also deletes the SSH keys bgit generated.
	RemoveKeys bool `json:"removeKeys"`
}

// UninstallResult summarises what removing bgit from this computer did.
type UninstallResult struct {
	RestoredRepos   []string `json:"restoredRepos"`
	FailedRepos     []string `json:"failedRepos"`
	RemovedKeys     []string `json:"removedKeys"`
	GitIdentityNote string   `json:"gitIdentityNote,omitempty"`
	Warnings        []string `json:"warnings"`
	ConfigRemoved   bool     `json:"configRemoved"`
}
