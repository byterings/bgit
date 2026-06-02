package models

// UserOperationResult describes a user mutation or lookup result.
type UserOperationResult struct {
	User    *User
	Changed bool
}

// WorkspaceUsersResult describes the resolved set of workspace users.
type WorkspaceUsersResult struct {
	Users []User
}

// WorkspaceRegistrationResult describes workspace registration outcome.
type WorkspaceRegistrationResult struct {
	Workspace     *Workspace
	Changed       bool
	AlreadyExists bool
}

// WorkspaceRemovalResult describes workspace removal outcome.
type WorkspaceRemovalResult struct {
	Workspace *Workspace
	Changed   bool
}

// BindingRemovalResult describes repository binding removal outcome.
type BindingRemovalResult struct {
	Binding *Binding
	Changed bool
}

// RepoOwnerResult describes repository owner resolution outcome.
type RepoOwnerResult struct {
	Owner *RepoOwner
}

// CommandOutputResult carries textual command output from core helpers.
type CommandOutputResult struct {
	Output string
}

// KeyLoadResult describes SSH agent key loading behavior.
type KeyLoadResult struct {
	Output  string
	Changed bool
}

// ConnectivityResult reports an SSH authentication check against GitHub.
type ConnectivityResult struct {
	Passed   bool
	Alias    string
	Username string
	Fix      string
	Message  string
}

// ConnectivityCheckResult wraps SSH connectivity checks.
type ConnectivityCheckResult struct {
	Results []ConnectivityResult
}

// SSHAgentSetupReport describes key loading results for setup flows.
type SSHAgentSetupReport struct {
	Added  map[string]string
	Failed map[string]string
}
