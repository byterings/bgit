//go:build devharness

// Desktop backend tests run with the devharness tag so they compile without
// the Wails window (and its WebKit/GTK dependency):
//
//	go test -tags devharness ./desktop/
package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/gitops"
	"github.com/byterings/bgit/internal/testenv"
)

var testCtx = context.Background()

// setupDesktop prepares an isolated machine with bgit set up and two
// identities: "work" (active) and "home".
func setupDesktop(t *testing.T) (home string, github *testenv.FakeGitHub) {
	t.Helper()
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("ssh-keygen is not installed")
	}

	home = testenv.Isolate(t)
	github = testenv.NewFakeGitHub(t, home)
	// The managed pre-push hook calls `bgit`; keep it away from any real
	// bgit installed on this machine.
	testenv.StubBgit(t, home, 0)

	if _, err := runDesktopSetup(); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	for _, identity := range []IdentityRequest{
		{Alias: "work", Name: "Work Dev", Email: "dev@work.example", GitHubUsername: "work-gh", GenerateSSHKey: true},
		{Alias: "home", Name: "Home Dev", Email: "dev@home.example", GitHubUsername: "home-gh", GenerateSSHKey: true},
	} {
		if _, err := addDesktopIdentity(identity); err != nil {
			t.Fatalf("add identity %s failed: %v", identity.Alias, err)
		}
	}
	return home, github
}

func TestSetupAndFirstIdentity(t *testing.T) {
	home, _ := setupDesktop(t)

	state, err := loadOnboardingState()
	if err != nil {
		t.Fatalf("onboarding state failed: %v", err)
	}
	if state.Needed || !state.SetupCompleted || !state.PushProtection || state.IdentityCount != 2 {
		t.Fatalf("unexpected onboarding state: %+v", state)
	}
	// The first identity added becomes active; the second does not replace it.
	if state.ActiveAlias != "work" {
		t.Fatalf("active alias = %q, want work", state.ActiveAlias)
	}
	if got := testenv.Git(t, home, "config", "--global", "user.email"); got != "dev@work.example" {
		t.Fatalf("global user.email = %q", got)
	}

	// Re-adding an identity whose key file still exists reuses the key.
	if _, err := deleteDesktopIdentity(DeleteIdentityRequest{Alias: "home", DeleteKeys: false}); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	result, err := addDesktopIdentity(IdentityRequest{Alias: "home", Name: "Home Dev", Email: "dev@home.example", GitHubUsername: "home-gh", GenerateSSHKey: true})
	if err != nil {
		t.Fatalf("re-add failed: %v", err)
	}
	if !strings.Contains(result.Message, "reused the existing SSH key") {
		t.Fatalf("expected the key to be reused, message = %q", result.Message)
	}
}

func TestNormalizeGitHubURL(t *testing.T) {
	valid := map[string]string{
		"https://github.com/org/app.git":            "https://github.com/org/app",
		"https://github.com/org/app/":               "https://github.com/org/app",
		"https://github.com/org/app/tree/main/docs": "https://github.com/org/app",
		"https://www.github.com/org/app":            "https://github.com/org/app",
		"github.com/org/app":                        "https://github.com/org/app",
		"  org/app  ":                               "https://github.com/org/app",
		"git@github.com:org/app.git":                "git@github.com:org/app.git",
	}
	for input, want := range valid {
		got, err := normalizeGitHubURL(input)
		if err != nil || got != want {
			t.Errorf("normalizeGitHubURL(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"", "https://gitlab.com/org/app", "https://github.com/org", "just-a-word", "org/app/extra"} {
		if got, err := normalizeGitHubURL(input); err == nil {
			t.Errorf("normalizeGitHubURL(%q) = %q, expected an error", input, got)
		}
	}
}

func TestCloneBindsAndAlignsRepository(t *testing.T) {
	home, github := setupDesktop(t)
	github.Seed(github.AddRepo("home-gh", "org", "app"))
	projects := filepath.Join(home, "projects")
	if err := os.MkdirAll(projects, 0755); err != nil {
		t.Fatal(err)
	}

	result, err := cloneDesktopRepository(testCtx, CloneRequest{URL: "https://github.com/org/app", ParentDir: projects, Alias: "home"})
	if err != nil {
		t.Fatalf("clone failed: %v", err)
	}
	repo := result.Repository
	if repo == nil || repo.Health != repoHealthOK || repo.Alias != "home" || repo.Source != repoSourceBinding {
		t.Fatalf("unexpected repository after clone: %+v", repo)
	}
	if repo.Path != filepath.Join(projects, "app") || repo.Branch != "main" {
		t.Fatalf("unexpected path or branch: %+v", repo)
	}
	// The clone is bound to "home" while "work" is the active identity.
	if got := testenv.Git(t, repo.Path, "config", "--local", "user.email"); got != "dev@home.example" {
		t.Fatalf("repository user.email = %q", got)
	}
	if got := testenv.Git(t, repo.Path, "remote", "get-url", "origin"); got != "git@github.com-home-gh:org/app.git" {
		t.Fatalf("origin = %q", got)
	}

	// Cloning into the same place again must fail before touching anything.
	if _, err := cloneDesktopRepository(testCtx, CloneRequest{URL: "org/app", ParentDir: projects, Alias: "home"}); err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("expected a not-empty error, got %v", err)
	}
	// An identity without access gets an authentication error, not a hang.
	_, err = cloneDesktopRepository(testCtx, CloneRequest{URL: "org/app", ParentDir: projects, FolderName: "as-work", Alias: "work"})
	if gitops.KindOf(err) != gitops.KindAuth {
		t.Fatalf("expected an auth error, got kind=%q err=%v", gitops.KindOf(err), err)
	}
	for _, request := range []CloneRequest{
		{URL: "https://gitlab.com/org/app", ParentDir: projects, Alias: "home"},
		{URL: "org/app", ParentDir: projects, Alias: "nobody"},
		{URL: "org/app", ParentDir: filepath.Join(home, "does-not-exist"), Alias: "home"},
		{URL: "org/app", ParentDir: projects, FolderName: "../escape", Alias: "home"},
	} {
		if _, err := cloneDesktopRepository(testCtx, request); err == nil {
			t.Errorf("clone request %+v should have failed", request)
		}
	}
}

func TestAddExistingRepositoryFixesIdentityAndRemote(t *testing.T) {
	home, _ := setupDesktop(t)
	repo := testenv.InitRepo(t, home, "existing")
	testenv.Git(t, repo, "remote", "add", "origin", "https://github.com/org/existing.git")

	result, err := addDesktopRepository(testCtx, RepoBindingRequest{Path: repo, Alias: "home"})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if result.Repository.Health != repoHealthOK || len(result.Repository.Problems) != 0 {
		t.Fatalf("repository should be healthy after adding: %+v", result.Repository)
	}
	if got := testenv.Git(t, repo, "remote", "get-url", "origin"); got != "git@github.com-home-gh:org/existing.git" {
		t.Fatalf("origin = %q", got)
	}

	// Changing the identity must move the remote to the new identity's key too.
	result, err = addDesktopRepository(testCtx, RepoBindingRequest{Path: repo, Alias: "work"})
	if err != nil {
		t.Fatalf("change identity failed: %v", err)
	}
	if !strings.Contains(result.Message, "changed from identity 'home' to 'work'") {
		t.Fatalf("message = %q", result.Message)
	}
	if got := testenv.Git(t, repo, "remote", "get-url", "origin"); got != "git@github.com-work-gh:org/existing.git" {
		t.Fatalf("origin after change = %q", got)
	}

	if _, err := addDesktopRepository(testCtx, RepoBindingRequest{Path: home, Alias: "work"}); err == nil {
		t.Fatal("adding a folder that is not a repository must fail")
	}
}

func TestRepositoryListHealthAndFix(t *testing.T) {
	home, _ := setupDesktop(t)

	// A workspace for "home" containing a repository bgit has never touched.
	workspace := filepath.Join(home, "home-projects")
	inWorkspace := testenv.InitRepo(t, workspace, "blog")
	testenv.Git(t, inWorkspace, "remote", "add", "origin", "git@github.com:me/blog.git")
	// Dependency and hidden folders must not be scanned.
	testenv.InitRepo(t, filepath.Join(workspace, "node_modules"), "dep")
	testenv.InitRepo(t, filepath.Join(workspace, ".cache"), "hidden")
	if _, err := addDesktopWorkspace(WorkspaceRequest{Path: workspace, Alias: "home"}); err != nil {
		t.Fatalf("add workspace failed: %v", err)
	}

	gitlab := testenv.InitRepo(t, home, "on-gitlab")
	testenv.Git(t, gitlab, "remote", "add", "origin", "https://gitlab.com/org/on-gitlab.git")
	if _, err := addDesktopRepository(testCtx, RepoBindingRequest{Path: gitlab, Alias: "work"}); err != nil {
		t.Fatalf("add gitlab repo failed: %v", err)
	}

	gone := testenv.InitRepo(t, home, "gone")
	if _, err := addDesktopRepository(testCtx, RepoBindingRequest{Path: gone, Alias: "work"}); err != nil {
		t.Fatalf("add repo failed: %v", err)
	}
	if err := os.RemoveAll(gone); err != nil {
		t.Fatal(err)
	}

	list, err := listDesktopRepositories(testCtx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	byName := map[string]RepositoryView{}
	for _, repo := range list.Repositories {
		byName[repo.Name] = repo
	}
	if len(byName) != 3 {
		t.Fatalf("expected 3 repositories, got %d: %+v", len(byName), list.Repositories)
	}

	blog := byName["blog"]
	if blog.Source != repoSourceWorkspace || blog.Alias != "home" || blog.Health != repoHealthAttention || !blog.CanFix || len(blog.Problems) != 2 {
		t.Fatalf("workspace repository: %+v", blog)
	}
	if byName["on-gitlab"].Health != repoHealthBlocked || byName["on-gitlab"].CanFix {
		t.Fatalf("gitlab repository: %+v", byName["on-gitlab"])
	}
	if byName["gone"].Health != repoHealthMissing || byName["gone"].Exists {
		t.Fatalf("missing repository: %+v", byName["gone"])
	}
	if list.NeedAttention != 3 {
		t.Fatalf("NeedAttention = %d, want 3", list.NeedAttention)
	}

	fixed, err := fixAllDesktopRepositories(testCtx)
	if err != nil {
		t.Fatalf("fix all failed: %v", err)
	}
	if !strings.Contains(fixed.Message, "Fixed 1 repository") || fixed.NeedAttention != 2 {
		t.Fatalf("fix all result: message=%q attention=%d", fixed.Message, fixed.NeedAttention)
	}
	if got := testenv.Git(t, inWorkspace, "remote", "get-url", "origin"); got != "git@github.com-home-gh:me/blog.git" {
		t.Fatalf("workspace repo origin = %q", got)
	}
	if got := testenv.Git(t, inWorkspace, "config", "--local", "user.email"); got != "dev@home.example" {
		t.Fatalf("workspace repo user.email = %q", got)
	}

	// A repository whose folder is gone can still be removed from the list.
	if _, err := removeDesktopRepository(gone); err != nil {
		t.Fatalf("remove missing repository failed: %v", err)
	}
	// Operating on a repository bgit does not manage is refused.
	stranger := testenv.InitRepo(t, home, "stranger")
	if _, err := fixDesktopRepository(testCtx, stranger); err == nil {
		t.Fatal("fixing an unmanaged repository must fail")
	}
	if _, err := getDesktopRepoState(testCtx, stranger); err == nil {
		t.Fatal("opening an unmanaged repository must fail")
	}
}

func TestWorkspaces(t *testing.T) {
	home, _ := setupDesktop(t)
	workspace := filepath.Join(home, "work-projects")
	nested := filepath.Join(workspace, "nested")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	if _, err := addDesktopWorkspace(WorkspaceRequest{Path: workspace, Alias: "work"}); err != nil {
		t.Fatalf("add workspace failed: %v", err)
	}
	for _, request := range []WorkspaceRequest{
		{Path: workspace, Alias: "home"},
		{Path: nested, Alias: "home"},
		{Path: home, Alias: "home"},
		{Path: filepath.Join(home, "missing"), Alias: "home"},
		{Path: nested, Alias: "nobody"},
	} {
		if _, err := addDesktopWorkspace(request); err == nil {
			t.Errorf("workspace request %+v should have failed", request)
		}
	}

	status, err := loadDesktopStatus()
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Workspaces) != 1 || status.Workspaces[0].Alias != "work" || !status.Workspaces[0].Exists {
		t.Fatalf("workspaces = %+v", status.Workspaces)
	}

	if _, err := removeDesktopWorkspace(workspace); err != nil {
		t.Fatalf("remove workspace failed: %v", err)
	}
	if _, err := removeDesktopWorkspace(workspace); err == nil {
		t.Fatal("removing a workspace twice must fail")
	}
	if _, err := os.Stat(nested); err != nil {
		t.Fatalf("removing a workspace must not delete folders: %v", err)
	}
}

func TestCommitPushPullThroughIdentity(t *testing.T) {
	home, github := setupDesktop(t)
	github.Seed(github.AddRepo("work-gh", "org", "app"))
	projects := filepath.Join(home, "projects")
	if err := os.MkdirAll(projects, 0755); err != nil {
		t.Fatal(err)
	}
	cloned, err := cloneDesktopRepository(testCtx, CloneRequest{URL: "org/app", ParentDir: projects, Alias: "work"})
	if err != nil {
		t.Fatalf("clone failed: %v", err)
	}
	repo := cloned.Repository.Path

	testenv.WriteFile(t, repo, "README.md", "line one\nedited\nline three\n")
	testenv.WriteFile(t, repo, "notes.txt", "notes\n")

	state, err := getDesktopRepoState(testCtx, repo)
	if err != nil || len(state.Status.Files) != 2 {
		t.Fatalf("state: %+v err=%v", state, err)
	}
	diff, err := getDesktopFileDiff(testCtx, FileDiffRequest{Path: repo, File: "README.md"})
	if err != nil || diff.Additions != 1 || diff.Deletions != 1 {
		t.Fatalf("diff: %+v err=%v", diff, err)
	}

	// Sabotage the repository identity: the commit must still use "work".
	testenv.Git(t, repo, "config", "--local", "user.email", "wrong@example.com")
	committed, err := commitDesktopChanges(testCtx, CommitChangesRequest{Path: repo, Summary: "Edit readme", Files: []string{"README.md"}})
	if err != nil || !committed.OK {
		t.Fatalf("commit: %+v err=%v", committed, err)
	}
	if got := testenv.Git(t, repo, "log", "-1", "--format=%ae"); got != "dev@work.example" {
		t.Fatalf("commit author email = %q", got)
	}
	if len(committed.State.Status.Files) != 1 || committed.State.Status.Ahead != 1 {
		t.Fatalf("state after commit: %+v", committed.State.Status)
	}

	// Push is refused with a fixable safety result while the identity is wrong.
	pushed, err := pushDesktopRepository(testCtx, repo)
	if err != nil || pushed.OK || pushed.Kind != gitops.KindSafety {
		t.Fatalf("push with mismatch: %+v err=%v", pushed, err)
	}
	if _, err := fixDesktopRepository(testCtx, repo); err != nil {
		t.Fatalf("fix failed: %v", err)
	}
	pushed, err = pushDesktopRepository(testCtx, repo)
	if err != nil || !pushed.OK || pushed.State.Status.Ahead != 0 {
		t.Fatalf("push after fix: %+v err=%v", pushed, err)
	}

	discarded, err := discardDesktopChanges(testCtx, DiscardChangesRequest{Path: repo, Files: []string{"notes.txt"}})
	if err != nil || !discarded.OK || len(discarded.State.Status.Files) != 0 {
		t.Fatalf("discard: %+v err=%v", discarded, err)
	}

	branch, err := createDesktopBranch(testCtx, BranchRequest{Path: repo, Name: "feature"})
	if err != nil || !branch.OK || branch.State.Status.Branch != "feature" {
		t.Fatalf("create branch: %+v err=%v", branch, err)
	}
	if switched, err := switchDesktopBranch(testCtx, BranchRequest{Path: repo, Name: "main"}); err != nil || !switched.OK {
		t.Fatalf("switch branch: %+v err=%v", switched, err)
	}
	branches, err := listDesktopBranches(testCtx, repo)
	if err != nil || len(branches) < 3 {
		t.Fatalf("branches: %+v err=%v", branches, err)
	}

	history, err := getDesktopHistory(testCtx, HistoryRequest{Path: repo, Limit: 10})
	if err != nil || len(history) != 2 || history[0].Summary != "Edit readme" || history[0].Unpushed {
		t.Fatalf("history: %+v err=%v", history, err)
	}
	files, err := getDesktopCommitFiles(testCtx, CommitFilesRequest{Path: repo, Commit: history[0].OID})
	if err != nil || len(files) != 1 || files[0].Path != "README.md" {
		t.Fatalf("commit files: %+v err=%v", files, err)
	}
	if _, err := getDesktopFileDiff(testCtx, FileDiffRequest{Path: repo, File: "README.md", Commit: history[0].OID}); err != nil {
		t.Fatalf("commit diff failed: %v", err)
	}

	if fetched, err := fetchDesktopRepository(testCtx, repo); err != nil || !fetched.OK {
		t.Fatalf("fetch: %+v err=%v", fetched, err)
	}
	if pulled, err := pullDesktopRepository(testCtx, repo); err != nil || !pulled.OK {
		t.Fatalf("pull: %+v err=%v", pulled, err)
	}
}

func TestPushRunsTheManagedHook(t *testing.T) {
	home, github := setupDesktop(t)
	github.Seed(github.AddRepo("work-gh", "org", "app"))
	cloned, err := cloneDesktopRepository(testCtx, CloneRequest{URL: "org/app", ParentDir: home, Alias: "work"})
	if err != nil {
		t.Fatalf("clone failed: %v", err)
	}
	repo := cloned.Repository.Path

	marker := testenv.StubBgit(t, home, 0)

	testenv.WriteFile(t, repo, "change.txt", "x\n")
	if committed, err := commitDesktopChanges(testCtx, CommitChangesRequest{Path: repo, Summary: "Change"}); err != nil || !committed.OK {
		t.Fatalf("commit: %+v err=%v", committed, err)
	}
	if pushed, err := pushDesktopRepository(testCtx, repo); err != nil || !pushed.OK {
		t.Fatalf("push: %+v err=%v", pushed, err)
	}
	ran, err := os.ReadFile(marker)
	if err != nil || strings.TrimSpace(string(ran)) != "check --from-hook" {
		t.Fatalf("the managed pre-push hook did not run bgit check: %q err=%v", ran, err)
	}

	// When the check fails, the push is reported as blocked by safety checks.
	testenv.StubBgit(t, home, 1)
	testenv.WriteFile(t, repo, "change.txt", "y\n")
	if committed, err := commitDesktopChanges(testCtx, CommitChangesRequest{Path: repo, Summary: "Change again"}); err != nil || !committed.OK {
		t.Fatalf("commit: %+v err=%v", committed, err)
	}
	blocked, err := pushDesktopRepository(testCtx, repo)
	if err != nil || blocked.OK || blocked.Kind != gitops.KindSafety {
		t.Fatalf("expected a safety block, got %+v err=%v", blocked, err)
	}
	if blocked.State == nil || blocked.State.Status.Ahead != 1 {
		t.Fatalf("the blocked commit must still be unpushed: %+v", blocked.State)
	}
}

func TestStatusIncludesWorkspaces(t *testing.T) {
	home, _ := setupDesktop(t)
	if _, err := addDesktopWorkspace(WorkspaceRequest{Path: home, Alias: "home"}); err != nil {
		t.Fatalf("add workspace failed: %v", err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Workspaces) != 1 || cfg.Workspaces[0].User != "home" {
		t.Fatalf("config workspaces = %+v", cfg.Workspaces)
	}
}

func TestUninstallFromDesktop(t *testing.T) {
	home, github := setupDesktop(t)
	github.Seed(github.AddRepo("work-gh", "org", "app"))
	cloned, err := cloneDesktopRepository(testCtx, CloneRequest{URL: "org/app", ParentDir: home, Alias: "work"})
	if err != nil {
		t.Fatalf("clone failed: %v", err)
	}
	repo := cloned.Repository.Path

	result, err := uninstallDesktop(UninstallRequest{RemoveKeys: true})
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}
	if !result.ConfigRemoved || len(result.Warnings) != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.RestoredRepos) != 1 || result.RestoredRepos[0] != repo {
		t.Fatalf("restored repos = %v", result.RestoredRepos)
	}
	if len(result.RemovedKeys) != 4 {
		t.Fatalf("removed keys = %v", result.RemovedKeys)
	}
	if got := testenv.Git(t, repo, "remote", "get-url", "origin"); got != "git@github.com:org/app.git" {
		t.Fatalf("origin after uninstall = %q", got)
	}

	// The app is back to its first-run state and nothing recreates the config.
	state, err := loadOnboardingState()
	if err != nil {
		t.Fatal(err)
	}
	if state.Configured || state.PushProtection || !state.Needed {
		t.Fatalf("state after uninstall: %+v", state)
	}
	if _, err := os.Stat(filepath.Join(home, ".bgit")); !os.IsNotExist(err) {
		t.Fatal("the settings folder must not come back after reading state")
	}
}

func TestIsCLIInvocation(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"check", "--from-hook"}, true},
		{[]string{"status"}, true},
		// Launched from Finder on macOS.
		{[]string{"-psn_0_12345"}, false},
		// Flags passed by Wails in dev mode and during binding generation.
		{[]string{"-tsprefix", "x"}, false},
		{[]string{"--help"}, false},
	}
	for _, tc := range cases {
		if got := isCLIInvocation(tc.args); got != tc.want {
			t.Errorf("isCLIInvocation(%v) = %v, want %v", tc.args, got, tc.want)
		}
	}
}
