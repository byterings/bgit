package models

// BindingResult describes the outcome of binding a repository to an identity.
type BindingResult struct {
	User            *User
	ExistingBinding *Binding
	Workspace       *Workspace
	NoChange        bool
}

// RepoOwner describes the identity that owns a repository for safety checks.
type RepoOwner struct {
	Alias  string
	Source string
	User   *User
}
