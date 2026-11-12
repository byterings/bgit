package models

// SetupResult describes what one-time setup installed or replaced.
type SetupResult struct {
	// HooksDir is the bgit-managed global hooks directory.
	HooksDir string
	// HookPath is the managed pre-push hook file.
	HookPath string
	// ReplacedHooksPath is the previous global core.hooksPath when setup
	// replaced one that was not managed by bgit. Empty otherwise.
	ReplacedHooksPath string
}

// CheckExpectation describes which identity a repository is expected to use
// for push safety checks.
type CheckExpectation struct {
	// OwnerAlias is the repository owner resolved from binding, workspace, or
	// remote. Empty when the repository has no owner.
	OwnerAlias string
	// Alias is the identity the repository must use: the owner when there is
	// one, otherwise the global active user. Empty when neither exists.
	Alias string
	// Source is "binding", "workspace", "remote", or "active".
	Source string
	// User is nil when Alias does not exist in the config.
	User *User
}

// RepoCheckReport is a non-interactive push safety report for one repository.
type RepoCheckReport struct {
	RepoRoot string

	ExpectedAlias  string
	ExpectedSource string
	ExpectedUser   *User

	// Issue is set when the expected identity cannot be determined.
	Issue string

	GitName         string
	GitEmail        string
	IdentityChecked bool
	IdentityMatches bool

	RemoteURL         string
	ExpectedRemoteURL string
	// RemoteChecked is false when there is no origin remote or its URL is not
	// a supported GitHub URL. RemoteIssue explains why.
	RemoteChecked bool
	RemoteMatches bool
	RemoteIssue   string
}

// OK reports whether the repository would pass the push safety check.
func (r *RepoCheckReport) OK() bool {
	return r != nil && r.Issue == "" && r.IdentityChecked && r.IdentityMatches && r.RemoteChecked && r.RemoteMatches
}

// RepoFixResult describes what a repository fix changed.
type RepoFixResult struct {
	IdentityFixed bool
	// IdentityScope is "repository" or "global".
	IdentityScope string
	RemoteFixed   bool
	Report        *RepoCheckReport
}
