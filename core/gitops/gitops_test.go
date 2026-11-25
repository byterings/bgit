package gitops

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/byterings/bgit/internal/testenv"
)

var ctx = context.Background()

const (
	authorName  = "Work Dev"
	authorEmail = "dev@work.example"
)

// newRepo creates a repository with one commit containing README.md.
func newRepo(t *testing.T) string {
	t.Helper()
	home := testenv.Isolate(t)
	repo := testenv.InitRepo(t, home, "repo")
	testenv.Git(t, repo, "config", "user.name", "Seed User")
	testenv.Git(t, repo, "config", "user.email", "seed@example.com")
	testenv.WriteFile(t, repo, "README.md", "line one\nline two\nline three\n")
	testenv.Git(t, repo, "add", "-A")
	testenv.Git(t, repo, "commit", "-q", "-m", "initial commit")
	return repo
}

func commitAll(t *testing.T, repo, summary string) *CommitResult {
	t.Helper()
	result, err := Commit(ctx, repo, CommitRequest{Summary: summary, AuthorName: authorName, AuthorEmail: authorEmail})
	if err != nil {
		t.Fatalf("commit %q failed: %v", summary, err)
	}
	return result
}

func fileKinds(status *Status) map[string]string {
	kinds := map[string]string{}
	for _, file := range status.Files {
		kinds[file.Path] = file.Kind
	}
	return kinds
}

func TestStatusUnbornRepository(t *testing.T) {
	home := testenv.Isolate(t)
	repo := testenv.InitRepo(t, home, "empty")
	testenv.WriteFile(t, repo, "a.txt", "hello\n")

	status, err := GetStatus(ctx, repo)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !status.Unborn || status.Branch != "main" || status.HeadOID != "" {
		t.Fatalf("unexpected unborn status: %+v", status)
	}
	if kinds := fileKinds(status); kinds["a.txt"] != "untracked" {
		t.Fatalf("a.txt should be untracked: %v", kinds)
	}

	// The first commit and its diff must work without a HEAD.
	diff, err := DiffFile(ctx, repo, "a.txt")
	if err != nil || diff.Additions != 1 {
		t.Fatalf("diff of new file in empty repo: %+v err=%v", diff, err)
	}
	commitAll(t, repo, "first")
	status, _ = GetStatus(ctx, repo)
	if status.Unborn || len(status.Files) != 0 || status.HeadOID == "" {
		t.Fatalf("status after first commit: %+v", status)
	}
}

func TestStatusReportsEveryKindOfChange(t *testing.T) {
	repo := newRepo(t)
	testenv.WriteFile(t, repo, "keep.txt", "keep\n")
	testenv.WriteFile(t, repo, "gone.txt", "gone\n")
	testenv.WriteFile(t, repo, "old name.txt", "rename me\nsame content\n")
	testenv.Git(t, repo, "add", "-A")
	testenv.Git(t, repo, "commit", "-q", "-m", "more files")

	testenv.WriteFile(t, repo, "README.md", "line one\nline 2\nline three\n")
	testenv.WriteFile(t, repo, "dir/new file.txt", "new\n")
	if err := os.Remove(filepath.Join(repo, "gone.txt")); err != nil {
		t.Fatal(err)
	}
	testenv.Git(t, repo, "mv", "old name.txt", "new name.txt")
	testenv.WriteFile(t, repo, "staged.txt", "staged\n")
	testenv.Git(t, repo, "add", "staged.txt")

	status, err := GetStatus(ctx, repo)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	kinds := fileKinds(status)
	want := map[string]string{
		"README.md":        "modified",
		"dir/new file.txt": "untracked",
		"gone.txt":         "deleted",
		"new name.txt":     "renamed",
		"staged.txt":       "added",
	}
	for path, kind := range want {
		if kinds[path] != kind {
			t.Errorf("%s: kind = %q, want %q (all: %v)", path, kinds[path], kind, kinds)
		}
	}
	if len(status.Files) != len(want) {
		t.Errorf("got %d files, want %d: %v", len(status.Files), len(want), kinds)
	}
	for _, file := range status.Files {
		if file.Path == "new name.txt" && file.OrigPath != "old name.txt" {
			t.Errorf("rename OrigPath = %q", file.OrigPath)
		}
	}
	if status.Branch != "main" || status.Detached || status.HasUpstream {
		t.Errorf("unexpected branch info: %+v", status)
	}
}

func TestDiffFile(t *testing.T) {
	repo := newRepo(t)

	testenv.WriteFile(t, repo, "README.md", "line one\nline 2\nline three\nline four\n")
	diff, err := DiffFile(ctx, repo, "README.md")
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if diff.Additions != 2 || diff.Deletions != 1 || diff.Binary {
		t.Fatalf("unexpected counts: %+v", diff)
	}
	var added, deleted *DiffLine
	for i := range diff.Lines {
		line := &diff.Lines[i]
		if line.Kind == "add" && line.Text == "line 2" {
			added = line
		}
		if line.Kind == "del" && line.Text == "line two" {
			deleted = line
		}
	}
	if added == nil || added.NewLine != 2 || deleted == nil || deleted.OldLine != 2 {
		t.Fatalf("line numbers wrong: add=%+v del=%+v", added, deleted)
	}

	// Staged changes are part of the same "since last commit" diff.
	testenv.Git(t, repo, "add", "README.md")
	staged, err := DiffFile(ctx, repo, "README.md")
	if err != nil || staged.Additions != 2 {
		t.Fatalf("staged diff: %+v err=%v", staged, err)
	}

	testenv.WriteFile(t, repo, "notes/todo.txt", "one\ntwo\n")
	untracked, err := DiffFile(ctx, repo, "notes/todo.txt")
	if err != nil || untracked.Additions != 2 || untracked.Deletions != 0 {
		t.Fatalf("untracked diff: %+v err=%v", untracked, err)
	}

	if err := os.WriteFile(filepath.Join(repo, "image.bin"), []byte{0, 1, 2, 0, 255, 0, 3}, 0644); err != nil {
		t.Fatal(err)
	}
	binary, err := DiffFile(ctx, repo, "image.bin")
	if err != nil || !binary.Binary {
		t.Fatalf("binary diff: %+v err=%v", binary, err)
	}

	if _, err := DiffFile(ctx, repo, "../outside.txt"); err == nil {
		t.Fatal("a path outside the repository must be rejected")
	}
	if _, err := DiffFile(ctx, repo, "/etc/passwd"); err == nil {
		t.Fatal("an absolute path must be rejected")
	}
}

func TestCommitAllUsesGivenIdentity(t *testing.T) {
	repo := newRepo(t)
	testenv.WriteFile(t, repo, "README.md", "changed\n")
	testenv.WriteFile(t, repo, "new.txt", "new\n")

	result, err := Commit(ctx, repo, CommitRequest{
		Summary:     "Update readme",
		Description: "First line of body.\n\nSecond paragraph.",
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
	})
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	if len(result.ShortOID) != 7 {
		t.Fatalf("ShortOID = %q", result.ShortOID)
	}

	// The repository config says "Seed User"; the commit must use the identity
	// that was passed in, for both author and committer.
	if got := testenv.Git(t, repo, "log", "-1", "--format=%an <%ae> / %cn <%ce>"); got != "Work Dev <dev@work.example> / Work Dev <dev@work.example>" {
		t.Fatalf("commit identity = %q", got)
	}
	if got := testenv.Git(t, repo, "log", "-1", "--format=%B"); got != "Update readme\n\nFirst line of body.\n\nSecond paragraph." {
		t.Fatalf("commit message = %q", got)
	}
	status, _ := GetStatus(ctx, repo)
	if len(status.Files) != 0 {
		t.Fatalf("working tree should be clean: %v", fileKinds(status))
	}
}

func TestCommitSelectedPathsLeavesTheRestAlone(t *testing.T) {
	repo := newRepo(t)
	testenv.WriteFile(t, repo, "tracked.txt", "v1\n")
	testenv.WriteFile(t, repo, "other.txt", "v1\n")
	testenv.Git(t, repo, "add", "-A")
	testenv.Git(t, repo, "commit", "-q", "-m", "add files")

	testenv.WriteFile(t, repo, "tracked.txt", "v2\n")
	testenv.WriteFile(t, repo, "other.txt", "v2\n")
	testenv.WriteFile(t, repo, "brand new.txt", "new\n")
	testenv.WriteFile(t, repo, "left out.txt", "new\n")
	// A change staged outside the app must not sneak into a partial commit.
	testenv.Git(t, repo, "add", "other.txt")

	_, err := Commit(ctx, repo, CommitRequest{
		Summary:     "Partial",
		Paths:       []string{"tracked.txt", "brand new.txt"},
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
	})
	if err != nil {
		t.Fatalf("partial commit failed: %v", err)
	}

	committed := testenv.Git(t, repo, "show", "--format=", "--name-only", "HEAD")
	if committed != "brand new.txt\ntracked.txt" {
		t.Fatalf("committed files = %q", committed)
	}

	kinds := fileKinds(mustStatus(t, repo))
	if len(kinds) != 2 || kinds["other.txt"] != "modified" || kinds["left out.txt"] != "untracked" {
		t.Fatalf("remaining changes = %v", kinds)
	}
	if staged := testenv.Git(t, repo, "diff", "--cached", "--name-only"); staged != "other.txt" {
		t.Fatalf("other.txt should still be staged, staged = %q", staged)
	}
}

func TestCommitValidation(t *testing.T) {
	repo := newRepo(t)

	if _, err := Commit(ctx, repo, CommitRequest{Summary: "  ", AuthorName: authorName, AuthorEmail: authorEmail}); err == nil {
		t.Fatal("an empty summary must be rejected")
	}
	if _, err := Commit(ctx, repo, CommitRequest{Summary: "x"}); err == nil {
		t.Fatal("a missing identity must be rejected")
	}
	if _, err := Commit(ctx, repo, CommitRequest{Summary: "x", AuthorName: authorName, AuthorEmail: authorEmail}); err == nil {
		t.Fatal("committing with no changes must fail")
	}
	testenv.WriteFile(t, repo, "a.txt", "a\n")
	if _, err := Commit(ctx, repo, CommitRequest{Summary: "x", Paths: []string{"../a.txt"}, AuthorName: authorName, AuthorEmail: authorEmail}); err == nil {
		t.Fatal("a path outside the repository must be rejected")
	}
}

func mustStatus(t *testing.T, repo string) *Status {
	t.Helper()
	status, err := GetStatus(ctx, repo)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	return status
}

func TestDiscard(t *testing.T) {
	repo := newRepo(t)
	testenv.WriteFile(t, repo, "delete me.txt", "original\n")
	testenv.WriteFile(t, repo, "rename me.txt", "rename content\nmore\n")
	testenv.WriteFile(t, repo, "untouched.txt", "original\n")
	testenv.Git(t, repo, "add", "-A")
	testenv.Git(t, repo, "commit", "-q", "-m", "add files")

	testenv.WriteFile(t, repo, "README.md", "edited\n")
	testenv.WriteFile(t, repo, "untracked.txt", "new\n")
	testenv.WriteFile(t, repo, "staged new.txt", "new\n")
	testenv.Git(t, repo, "add", "staged new.txt")
	if err := os.Remove(filepath.Join(repo, "delete me.txt")); err != nil {
		t.Fatal(err)
	}
	testenv.Git(t, repo, "mv", "rename me.txt", "renamed.txt")
	testenv.WriteFile(t, repo, "untouched.txt", "keep this edit\n")

	err := Discard(ctx, repo, []string{"README.md", "untracked.txt", "staged new.txt", "delete me.txt", "renamed.txt"})
	if err != nil {
		t.Fatalf("discard failed: %v", err)
	}

	kinds := fileKinds(mustStatus(t, repo))
	if len(kinds) != 1 || kinds["untouched.txt"] != "modified" {
		t.Fatalf("only untouched.txt should remain changed, got %v", kinds)
	}
	for path, want := range map[string]string{
		"README.md":     "line one\nline two\nline three\n",
		"delete me.txt": "original\n",
		"rename me.txt": "rename content\nmore\n",
	} {
		content, err := os.ReadFile(filepath.Join(repo, path))
		if err != nil || string(content) != want {
			t.Errorf("%s = %q (err %v), want %q", path, content, err, want)
		}
	}
	for _, path := range []string{"untracked.txt", "staged new.txt", "renamed.txt"} {
		if _, err := os.Stat(filepath.Join(repo, path)); !os.IsNotExist(err) {
			t.Errorf("%s should have been removed", path)
		}
	}

	if err := Discard(ctx, repo, []string{"../../etc/hosts"}); err == nil {
		t.Fatal("a path outside the repository must be rejected")
	}
	if err := Discard(ctx, repo, nil); err == nil {
		t.Fatal("discarding nothing must be rejected")
	}
}

func TestBranches(t *testing.T) {
	repo := newRepo(t)

	if err := CreateBranch(ctx, repo, "feature/login"); err != nil {
		t.Fatalf("create branch failed: %v", err)
	}
	if status := mustStatus(t, repo); status.Branch != "feature/login" {
		t.Fatalf("current branch = %q", status.Branch)
	}
	if err := CreateBranch(ctx, repo, "feature/login"); err == nil {
		t.Fatal("creating an existing branch must fail")
	}
	for _, bad := range []string{"", "-x", "has space", "bad..name", "ends.lock"} {
		if err := CreateBranch(ctx, repo, bad); err == nil {
			t.Errorf("branch name %q must be rejected", bad)
		}
	}

	if err := SwitchBranch(ctx, repo, "main"); err != nil {
		t.Fatalf("switch failed: %v", err)
	}
	if err := SwitchBranch(ctx, repo, "missing"); KindOf(err) != KindNotFound {
		t.Fatalf("switching to a missing branch: kind=%q err=%v", KindOf(err), err)
	}

	branches, err := ListBranches(ctx, repo)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	names := map[string]bool{}
	current := ""
	for _, branch := range branches {
		names[branch.Name] = true
		if branch.Current {
			current = branch.Name
		}
	}
	if !names["main"] || !names["feature/login"] || current != "main" {
		t.Fatalf("branches = %+v", branches)
	}

	// Switching must refuse to overwrite uncommitted work.
	testenv.Git(t, repo, "checkout", "-q", "feature/login")
	testenv.WriteFile(t, repo, "README.md", "feature version\n")
	commitAll(t, repo, "feature change")
	testenv.Git(t, repo, "checkout", "-q", "main")
	testenv.WriteFile(t, repo, "README.md", "uncommitted on main\n")
	if err := SwitchBranch(ctx, repo, "feature/login"); KindOf(err) != KindDirty {
		t.Fatalf("expected a dirty-tree error, got kind=%q err=%v", KindOf(err), err)
	}
}

func TestLogAndCommitDetails(t *testing.T) {
	repo := newRepo(t)
	testenv.WriteFile(t, repo, "README.md", "line one\nline two\nline three\nline four\n")
	testenv.WriteFile(t, repo, "added.txt", "a\n")
	second := commitAll(t, repo, "Second commit")

	commits, err := Log(ctx, repo, 10, 0)
	if err != nil {
		t.Fatalf("log failed: %v", err)
	}
	if len(commits) != 2 || commits[0].Summary != "Second commit" || commits[1].Summary != "initial commit" {
		t.Fatalf("unexpected log: %+v", commits)
	}
	if commits[0].AuthorEmail != authorEmail || commits[0].Time == 0 || !commits[0].Unpushed {
		t.Fatalf("unexpected commit fields: %+v", commits[0])
	}

	paged, _ := Log(ctx, repo, 1, 1)
	if len(paged) != 1 || paged[0].Summary != "initial commit" {
		t.Fatalf("paging failed: %+v", paged)
	}

	files, err := CommitFiles(ctx, repo, second.OID)
	if err != nil {
		t.Fatalf("commit files failed: %v", err)
	}
	got := map[string]string{}
	for _, file := range files {
		got[file.Path] = file.Kind
	}
	if len(got) != 2 || got["README.md"] != "modified" || got["added.txt"] != "added" {
		t.Fatalf("commit files = %v", got)
	}

	diff, err := DiffCommitFile(ctx, repo, second.OID, "README.md")
	if err != nil || diff.Additions != 1 || diff.Deletions != 0 {
		t.Fatalf("commit diff: %+v err=%v", diff, err)
	}

	// The root commit has no parent and is compared with an empty tree.
	rootFiles, err := CommitFiles(ctx, repo, commits[1].OID)
	if err != nil || len(rootFiles) != 1 || rootFiles[0].Kind != "added" {
		t.Fatalf("root commit files: %+v err=%v", rootFiles, err)
	}

	for _, bad := range []string{"HEAD", "--all", "abc", "main; rm -rf /"} {
		if _, err := CommitFiles(ctx, repo, bad); err == nil {
			t.Errorf("revision %q must be rejected", bad)
		}
	}
}

// newRemotePair returns a bare remote and a clone of it with one pushed commit.
func newRemotePair(t *testing.T) (home, remote, clone string) {
	t.Helper()
	home = testenv.Isolate(t)
	remote = filepath.Join(home, "remote.git")
	testenv.Git(t, home, "init", "-q", "--bare", "-b", "main", remote)

	clone = filepath.Join(home, "clone-a")
	if err := Clone(ctx, remote, clone); err != nil {
		t.Fatalf("clone failed: %v", err)
	}
	testenv.WriteFile(t, clone, "README.md", "line one\nline two\nline three\n")
	commitAll(t, clone, "initial commit")
	if _, err := Push(ctx, clone); err != nil {
		t.Fatalf("first push failed: %v", err)
	}
	return home, remote, clone
}

func TestCloneAndPushPublishesBranch(t *testing.T) {
	_, remote, clone := newRemotePair(t)

	status := mustStatus(t, clone)
	if !status.HasUpstream || status.Upstream != "origin/main" || status.Ahead != 0 || status.Behind != 0 {
		t.Fatalf("status after first push: %+v", status)
	}
	if got := testenv.Git(t, remote, "log", "--format=%s", "main"); got != "initial commit" {
		t.Fatalf("remote log = %q", got)
	}

	testenv.WriteFile(t, clone, "README.md", "changed\n")
	commitAll(t, clone, "second")
	if status := mustStatus(t, clone); status.Ahead != 1 {
		t.Fatalf("ahead = %d, want 1", status.Ahead)
	}
	commits, _ := Log(ctx, clone, 10, 0)
	if !commits[0].Unpushed || commits[1].Unpushed {
		t.Fatalf("unpushed flags wrong: %+v", commits)
	}

	result, err := Push(ctx, clone)
	if err != nil || result.SetUpstream {
		t.Fatalf("second push: %+v err=%v", result, err)
	}
	if status := mustStatus(t, clone); status.Ahead != 0 {
		t.Fatalf("ahead after push = %d", status.Ahead)
	}

	// A new branch is published on its first push.
	if err := CreateBranch(ctx, clone, "feature"); err != nil {
		t.Fatal(err)
	}
	result, err = Push(ctx, clone)
	if err != nil || !result.SetUpstream {
		t.Fatalf("branch publish: %+v err=%v", result, err)
	}
	if status := mustStatus(t, clone); status.Upstream != "origin/feature" {
		t.Fatalf("upstream = %q", status.Upstream)
	}
}

func TestCloneRejectsOccupiedDestination(t *testing.T) {
	home, remote, _ := newRemotePair(t)
	occupied := filepath.Join(home, "occupied")
	testenv.WriteFile(t, occupied, "file.txt", "x\n")

	err := Clone(ctx, remote, occupied)
	if err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("expected a non-empty destination error, got %v", err)
	}
	if err := Clone(ctx, "--upload-pack=evil", filepath.Join(home, "x")); err == nil {
		t.Fatal("an address starting with '-' must be rejected")
	}
}

func TestFetchPullAndRejectedPush(t *testing.T) {
	home, remote, cloneA := newRemotePair(t)
	cloneB := filepath.Join(home, "clone-b")
	if err := Clone(ctx, remote, cloneB); err != nil {
		t.Fatalf("second clone failed: %v", err)
	}

	testenv.WriteFile(t, cloneB, "from-b.txt", "b\n")
	commitAll(t, cloneB, "commit from b")
	if _, err := Push(ctx, cloneB); err != nil {
		t.Fatalf("push from b failed: %v", err)
	}

	// A does not know about B's commit until it fetches.
	if _, err := Fetch(ctx, cloneA); err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if status := mustStatus(t, cloneA); status.Behind != 1 || status.Ahead != 0 {
		t.Fatalf("after fetch: ahead=%d behind=%d", status.Ahead, status.Behind)
	}

	// Pushing while behind is rejected with a clear category.
	testenv.WriteFile(t, cloneA, "from-a.txt", "a\n")
	commitAll(t, cloneA, "commit from a")
	if _, err := Push(ctx, cloneA); KindOf(err) != KindRejected {
		t.Fatalf("expected a rejected push, got kind=%q err=%v", KindOf(err), err)
	}

	// Both sides changed different files: pull merges without prompting.
	if _, err := Pull(ctx, cloneA, authorName, authorEmail); err != nil {
		t.Fatalf("pull failed: %v", err)
	}
	status := mustStatus(t, cloneA)
	if status.Behind != 0 || status.MergeInProgress || len(status.Files) != 0 {
		t.Fatalf("after pull: %+v", status)
	}
	if _, err := os.Stat(filepath.Join(cloneA, "from-b.txt")); err != nil {
		t.Fatalf("pulled file missing: %v", err)
	}
	// The merge commit carries the identity that was passed in.
	if got := testenv.Git(t, cloneA, "log", "-1", "--format=%cn <%ce>"); got != "Work Dev <dev@work.example>" {
		t.Fatalf("merge commit identity = %q", got)
	}
	if _, err := Push(ctx, cloneA); err != nil {
		t.Fatalf("push after pull failed: %v", err)
	}
}

func TestPullConflictCanBeAborted(t *testing.T) {
	home, remote, cloneA := newRemotePair(t)
	cloneB := filepath.Join(home, "clone-b")
	if err := Clone(ctx, remote, cloneB); err != nil {
		t.Fatal(err)
	}

	testenv.WriteFile(t, cloneB, "README.md", "line one\nB CHANGE\nline three\n")
	commitAll(t, cloneB, "b edits line two")
	if _, err := Push(ctx, cloneB); err != nil {
		t.Fatal(err)
	}

	testenv.WriteFile(t, cloneA, "README.md", "line one\nA CHANGE\nline three\n")
	commitAll(t, cloneA, "a edits line two")

	if _, err := Pull(ctx, cloneA, authorName, authorEmail); KindOf(err) != KindConflict {
		t.Fatalf("expected a conflict, got kind=%q err=%v", KindOf(err), err)
	}
	status := mustStatus(t, cloneA)
	if !status.MergeInProgress || status.Conflicts != 1 || status.UnresolvedConflicts != 1 || fileKinds(status)["README.md"] != "conflicted" {
		t.Fatalf("conflict state not reported: %+v", status)
	}
	if _, err := Commit(ctx, cloneA, CommitRequest{Summary: "x", AuthorName: authorName, AuthorEmail: authorEmail}); KindOf(err) != KindConflict {
		t.Fatalf("committing with conflicts must fail with a conflict error, got %v", err)
	}

	if err := AbortMerge(ctx, cloneA); err != nil {
		t.Fatalf("abort failed: %v", err)
	}
	status = mustStatus(t, cloneA)
	if status.MergeInProgress || len(status.Files) != 0 {
		t.Fatalf("after abort: %+v", status)
	}
	content, _ := os.ReadFile(filepath.Join(cloneA, "README.md"))
	if string(content) != "line one\nA CHANGE\nline three\n" {
		t.Fatalf("local commit content lost: %q", content)
	}
}

func TestSyncNeedsRemoteAndUpstream(t *testing.T) {
	repo := newRepo(t)

	if _, err := Fetch(ctx, repo); KindOf(err) != KindNotFound {
		t.Fatalf("fetch without a remote: kind=%q err=%v", KindOf(err), err)
	}
	if _, err := Push(ctx, repo); KindOf(err) != KindNotFound {
		t.Fatalf("push without a remote: kind=%q err=%v", KindOf(err), err)
	}

	bare := filepath.Join(filepath.Dir(repo), "bare.git")
	testenv.Git(t, repo, "init", "-q", "--bare", bare)
	testenv.Git(t, repo, "remote", "add", "origin", bare)
	if _, err := Pull(ctx, repo, authorName, authorEmail); KindOf(err) != KindNoUpstream {
		t.Fatalf("pull without an upstream: kind=%q err=%v", KindOf(err), err)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		output string
		kind   string
	}{
		{"git@github.com: Permission denied (publickey).\nfatal: Could not read from remote repository.", KindAuth},
		{"ERROR: Repository not found.\nfatal: Could not read from remote repository.", KindAuth},
		{"ssh: Could not resolve hostname github.com-work: Name or service not known", KindNetwork},
		{"bgit: push blocked by safety checks\nerror: failed to push some refs to 'x'", KindSafety},
		{" ! [rejected]        main -> main (fetch first)\nerror: failed to push some refs", KindRejected},
		{"error: Your local changes to the following files would be overwritten by merge:", KindDirty},
		{"CONFLICT (content): Merge conflict in README.md\nAutomatic merge failed", KindConflict},
		{"fatal: The current branch x has no upstream branch.", KindNoUpstream},
		{"*** Please tell me who you are.\nfatal: unable to auto-detect email address (got 'x@y.(none)')", KindIdentity},
		{"fatal: something nobody expected", KindUnknown},
	}
	for _, tc := range cases {
		gitErr := classify("", tc.output)
		if gitErr.Kind != tc.kind {
			t.Errorf("output %q: kind = %q, want %q", tc.output, gitErr.Kind, tc.kind)
		}
		if gitErr.Message == "" {
			t.Errorf("output %q: message is empty", tc.output)
		}
	}
	if got := classify("", "fatal: something nobody expected").Message; got != "Something nobody expected" {
		t.Errorf("fallback message = %q", got)
	}
}

func TestResolvedConflictCanBeCommitted(t *testing.T) {
	home, remote, cloneA := newRemotePair(t)
	cloneB := filepath.Join(home, "clone-b")
	if err := Clone(ctx, remote, cloneB); err != nil {
		t.Fatal(err)
	}
	testenv.WriteFile(t, cloneB, "README.md", "line one\nB CHANGE\nline three\n")
	commitAll(t, cloneB, "b edits line two")
	if _, err := Push(ctx, cloneB); err != nil {
		t.Fatal(err)
	}
	testenv.WriteFile(t, cloneA, "README.md", "line one\nA CHANGE\nline three\n")
	commitAll(t, cloneA, "a edits line two")
	if _, err := Pull(ctx, cloneA, authorName, authorEmail); KindOf(err) != KindConflict {
		t.Fatalf("expected a conflict, got %v", err)
	}

	// The user edits the file in their editor and removes the markers. They
	// never run `git add`; the app must still see the conflict as resolved.
	testenv.WriteFile(t, cloneA, "README.md", "line one\nA AND B CHANGE\nline three\n")
	status := mustStatus(t, cloneA)
	if status.Conflicts != 1 || status.UnresolvedConflicts != 0 || !status.MergeInProgress {
		t.Fatalf("resolved conflict not detected: %+v", status)
	}

	commitAll(t, cloneA, "Merge remote changes")
	status = mustStatus(t, cloneA)
	if status.MergeInProgress || len(status.Files) != 0 || status.Ahead != 2 {
		t.Fatalf("after committing the merge: %+v", status)
	}
	if parents := testenv.Git(t, cloneA, "log", "-1", "--format=%P"); len(strings.Fields(parents)) != 2 {
		t.Fatalf("the commit should be a merge with two parents, got %q", parents)
	}
	if _, err := Push(ctx, cloneA); err != nil {
		t.Fatalf("push after merge failed: %v", err)
	}
}

func TestListRemotes(t *testing.T) {
	repo := newRepo(t)

	remotes, err := ListRemotes(ctx, repo)
	if err != nil || len(remotes) != 0 {
		t.Fatalf("a repository without remotes: %+v err=%v", remotes, err)
	}

	testenv.Git(t, repo, "remote", "add", "origin", "git@github.com-work-gh:org/app.git")
	testenv.Git(t, repo, "remote", "add", "upstream", "https://github.com/other/app.git")
	testenv.Git(t, repo, "remote", "set-url", "--push", "upstream", "https://github.com/other/app-push.git")

	remotes, err = ListRemotes(ctx, repo)
	if err != nil || len(remotes) != 2 {
		t.Fatalf("remotes: %+v err=%v", remotes, err)
	}
	if remotes[0].Name != "origin" || remotes[0].FetchURL != "git@github.com-work-gh:org/app.git" || remotes[0].PushURL != remotes[0].FetchURL {
		t.Fatalf("origin: %+v", remotes[0])
	}
	if remotes[1].Name != "upstream" || remotes[1].FetchURL != "https://github.com/other/app.git" || remotes[1].PushURL != "https://github.com/other/app-push.git" {
		t.Fatalf("upstream: %+v", remotes[1])
	}
}

func TestDeleteBranch(t *testing.T) {
	repo := newRepo(t)

	// A branch with nothing new can be deleted without force.
	testenv.Git(t, repo, "branch", "merged")
	if err := DeleteBranch(ctx, repo, "merged", false); err != nil {
		t.Fatalf("deleting a merged branch failed: %v", err)
	}

	// A branch with its own commit needs force, and says why.
	if err := CreateBranch(ctx, repo, "feature"); err != nil {
		t.Fatal(err)
	}
	testenv.WriteFile(t, repo, "feature.txt", "x\n")
	commitAll(t, repo, "feature work")

	if err := DeleteBranch(ctx, repo, "feature", false); err == nil {
		t.Fatal("the current branch must not be deleted")
	}
	if err := SwitchBranch(ctx, repo, "main"); err != nil {
		t.Fatal(err)
	}
	if err := DeleteBranch(ctx, repo, "feature", false); KindOf(err) != KindUnmerged {
		t.Fatalf("expected an unmerged error, got kind=%q err=%v", KindOf(err), err)
	}
	if err := DeleteBranch(ctx, repo, "feature", true); err != nil {
		t.Fatalf("forced delete failed: %v", err)
	}

	branches, _ := ListBranches(ctx, repo)
	for _, branch := range branches {
		if branch.Name == "feature" || branch.Name == "merged" {
			t.Fatalf("branch %s still exists", branch.Name)
		}
	}
	for _, bad := range []string{"", "-D", "missing"} {
		if err := DeleteBranch(ctx, repo, bad, false); err == nil {
			t.Errorf("deleting %q must fail", bad)
		}
	}
}
