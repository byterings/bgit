package models

// GitFileChange is one changed path in a repository working tree.
type GitFileChange struct {
	Path string `json:"path"`
	// OrigPath is the previous path of a renamed or copied file.
	OrigPath string `json:"origPath"`
	// Kind is "modified", "added", "deleted", "renamed", "copied",
	// "untracked", "conflicted", or "typechange".
	Kind string `json:"kind"`
	// Staged is true when the index already holds a change for this path.
	Staged bool `json:"staged"`
	// Unstaged is true when the working tree differs from the index.
	Unstaged bool `json:"unstaged"`
}

// GitStatus describes the branch and working tree state of a repository.
type GitStatus struct {
	// Branch is the current branch name, or "" when HEAD is detached.
	Branch   string `json:"branch"`
	Detached bool   `json:"detached"`
	// HeadOID is the short commit id of HEAD, or "" before the first commit.
	HeadOID string `json:"headOid"`
	// Unborn is true before the first commit on the current branch.
	Unborn bool `json:"unborn"`

	Upstream    string `json:"upstream"`
	HasUpstream bool   `json:"hasUpstream"`
	Ahead       int    `json:"ahead"`
	Behind      int    `json:"behind"`

	Files []GitFileChange `json:"files"`
	// Truncated is true when the repository has more changes than were listed.
	Truncated bool `json:"truncated"`

	MergeInProgress  bool `json:"mergeInProgress"`
	RebaseInProgress bool `json:"rebaseInProgress"`
	// Conflicts counts files Git still lists as conflicted.
	Conflicts int `json:"conflicts"`
	// UnresolvedConflicts counts conflicted files that still contain conflict
	// markers. When it reaches zero the merge can be committed.
	UnresolvedConflicts int `json:"unresolvedConflicts"`
}

// GitDiffLine is one line of a parsed unified diff.
type GitDiffLine struct {
	// Kind is "hunk", "add", "del", "context", or "note".
	Kind    string `json:"kind"`
	Text    string `json:"text"`
	OldLine int    `json:"oldLine"`
	NewLine int    `json:"newLine"`
}

// GitDiff is the change to one file.
type GitDiff struct {
	Path      string        `json:"path"`
	Binary    bool          `json:"binary"`
	Truncated bool          `json:"truncated"`
	Additions int           `json:"additions"`
	Deletions int           `json:"deletions"`
	Lines     []GitDiffLine `json:"lines"`
}

// GitBranch is a local or remote branch.
type GitBranch struct {
	Name     string `json:"name"`
	Remote   bool   `json:"remote"`
	Current  bool   `json:"current"`
	Upstream string `json:"upstream"`
	OID      string `json:"oid"`
	// CommitTime is the committer time of the branch tip, in Unix seconds.
	CommitTime int64 `json:"commitTime"`
}

// GitCommit is one entry of the commit history.
type GitCommit struct {
	OID         string `json:"oid"`
	ShortOID    string `json:"shortOid"`
	AuthorName  string `json:"authorName"`
	AuthorEmail string `json:"authorEmail"`
	// Time is the author time in Unix seconds.
	Time    int64  `json:"time"`
	Summary string `json:"summary"`
	Body    string `json:"body"`
	// Unpushed is true when the commit is not on the upstream branch yet.
	Unpushed bool `json:"unpushed"`
}

// GitCommitRequest describes a commit to create.
type GitCommitRequest struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	// Paths limits the commit to these files. Empty commits every change.
	Paths []string `json:"paths"`
	// AuthorName and AuthorEmail set the commit identity explicitly, so the
	// commit never depends on whatever Git config happens to be in effect.
	AuthorName  string `json:"authorName"`
	AuthorEmail string `json:"authorEmail"`
}

// GitCommitResult describes a created commit.
type GitCommitResult struct {
	OID      string `json:"oid"`
	ShortOID string `json:"shortOid"`
	Summary  string `json:"summary"`
}

// GitSyncResult describes the outcome of a fetch, pull, or push.
type GitSyncResult struct {
	// Output is Git's own summary of what happened.
	Output string `json:"output"`
	// SetUpstream is true when a push published the branch for the first time.
	SetUpstream bool `json:"setUpstream"`
}

// GitRemote is one configured remote of a repository.
type GitRemote struct {
	Name     string `json:"name"`
	FetchURL string `json:"fetchUrl"`
	PushURL  string `json:"pushUrl"`
}
