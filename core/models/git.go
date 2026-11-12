package models

// GitFileChange is one changed path in a repository working tree.
type GitFileChange struct {
	Path string
	// OrigPath is the previous path of a renamed or copied file.
	OrigPath string
	// Kind is "modified", "added", "deleted", "renamed", "copied",
	// "untracked", "conflicted", or "typechange".
	Kind string
	// Staged is true when the index already holds a change for this path.
	Staged bool
	// Unstaged is true when the working tree differs from the index.
	Unstaged bool
}

// GitStatus describes the branch and working tree state of a repository.
type GitStatus struct {
	// Branch is the current branch name, or "" when HEAD is detached.
	Branch   string
	Detached bool
	// HeadOID is the short commit id of HEAD, or "" before the first commit.
	HeadOID string
	// Unborn is true before the first commit on the current branch.
	Unborn bool

	Upstream    string
	HasUpstream bool
	Ahead       int
	Behind      int

	Files []GitFileChange
	// Truncated is true when the repository has more changes than were listed.
	Truncated bool

	MergeInProgress  bool
	RebaseInProgress bool
	Conflicts        int
}

// GitDiffLine is one line of a parsed unified diff.
type GitDiffLine struct {
	// Kind is "hunk", "add", "del", "context", or "note".
	Kind    string
	Text    string
	OldLine int
	NewLine int
}

// GitDiff is the change to one file.
type GitDiff struct {
	Path      string
	Binary    bool
	Truncated bool
	Additions int
	Deletions int
	Lines     []GitDiffLine
}

// GitBranch is a local or remote branch.
type GitBranch struct {
	Name     string
	Remote   bool
	Current  bool
	Upstream string
	OID      string
	// CommitTime is the committer time of the branch tip, in Unix seconds.
	CommitTime int64
}

// GitCommit is one entry of the commit history.
type GitCommit struct {
	OID         string
	ShortOID    string
	AuthorName  string
	AuthorEmail string
	// Time is the author time in Unix seconds.
	Time    int64
	Summary string
	Body    string
	// Unpushed is true when the commit is not on the upstream branch yet.
	Unpushed bool
}

// GitCommitRequest describes a commit to create.
type GitCommitRequest struct {
	Summary     string
	Description string
	// Paths limits the commit to these files. Empty commits every change.
	Paths []string
	// AuthorName and AuthorEmail set the commit identity explicitly, so the
	// commit never depends on whatever Git config happens to be in effect.
	AuthorName  string
	AuthorEmail string
}

// GitCommitResult describes a created commit.
type GitCommitResult struct {
	OID      string
	ShortOID string
	Summary  string
}

// GitSyncResult describes the outcome of a fetch, pull, or push.
type GitSyncResult struct {
	// Output is Git's own summary of what happened.
	Output string
	// SetUpstream is true when a push published the branch for the first time.
	SetUpstream bool
}
