package models

// ResolutionSource indicates how the identity was resolved.
type ResolutionSource string

const (
	SourceWorkspace ResolutionSource = "workspace"
	SourceBinding   ResolutionSource = "binding"
	SourceGlobal    ResolutionSource = "global"
)

// Resolution contains the resolved identity and its source.
type Resolution struct {
	User   *User
	Alias  string
	Source ResolutionSource
	Path   string
}

// DeleteResult describes the result of removing an identity.
type DeleteResult struct {
	User          User
	ActiveCleared bool
}

// ActivateResult describes the result of switching identities.
type ActivateResult struct {
	User      *User
	KeyLoaded bool
}
