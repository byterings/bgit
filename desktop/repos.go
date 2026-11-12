package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	corecheck "github.com/byterings/bgit/core/check"
	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/gitops"
	coreidentity "github.com/byterings/bgit/core/identity"
	corerepo "github.com/byterings/bgit/core/repo"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/git"
	"github.com/byterings/bgit/internal/platform"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	repoSourceBinding   = "binding"
	repoSourceWorkspace = "workspace"

	repoHealthOK        = "ok"
	repoHealthAttention = "attention"
	repoHealthBlocked   = "blocked"
	repoHealthMissing   = "missing"

	// workspaceScanDepth is how many folder levels below a workspace are
	// searched for repositories.
	workspaceScanDepth = 3
	// repoInspectWorkers limits how many repositories are inspected at once.
	repoInspectWorkers = 6
)

type repoEntry struct {
	path       string
	source     string
	sourcePath string
}

// listDesktopRepositories returns every repository bgit knows about: bound
// repositories, plus repositories found inside workspace folders.
func listDesktopRepositories(ctx context.Context) (*RepositoriesResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	entries := knownRepositories(cfg)
	views := make([]RepositoryView, len(entries))

	// Each repository needs a few git calls; a small worker pool keeps the
	// page fast without starting dozens of processes at once.
	var wg sync.WaitGroup
	slots := make(chan struct{}, repoInspectWorkers)
	for i, entry := range entries {
		wg.Add(1)
		go func(i int, entry repoEntry) {
			defer wg.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			views[i] = buildRepositoryView(ctx, cfg, entry)
		}(i, entry)
	}
	wg.Wait()

	result := &RepositoriesResult{Repositories: views}
	for _, view := range views {
		if view.Health != repoHealthOK {
			result.NeedAttention++
		}
	}
	return result, nil
}

// knownRepositories lists bound repositories first, then repositories found
// in workspaces that are not bound. Sorted by folder name within each group.
func knownRepositories(cfg *config.Config) []repoEntry {
	seen := map[string]bool{}
	bound := []repoEntry{}
	for _, binding := range cfg.GetBindings() {
		if seen[binding.Path] {
			continue
		}
		seen[binding.Path] = true
		bound = append(bound, repoEntry{path: binding.Path, source: repoSourceBinding})
	}

	inWorkspaces := []repoEntry{}
	for _, workspace := range cfg.GetWorkspaces() {
		for _, repoPath := range findRepositories(workspace.Path, workspaceScanDepth) {
			if seen[repoPath] {
				continue
			}
			seen[repoPath] = true
			inWorkspaces = append(inWorkspaces, repoEntry{path: repoPath, source: repoSourceWorkspace, sourcePath: workspace.Path})
		}
	}

	byName := func(entries []repoEntry) {
		sort.SliceStable(entries, func(i, j int) bool {
			return strings.ToLower(filepath.Base(entries[i].path)) < strings.ToLower(filepath.Base(entries[j].path))
		})
	}
	byName(bound)
	byName(inWorkspaces)
	return append(bound, inWorkspaces...)
}

// findRepositories returns Git repositories under root, down to maxDepth
// folder levels. It does not look inside a repository, hidden folders, or
// dependency folders.
func findRepositories(root string, maxDepth int) []string {
	found := []string{}
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
			found = append(found, dir)
			return
		}
		if depth >= maxDepth {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() || strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				continue
			}
			walk(filepath.Join(dir, name), depth+1)
		}
	}
	walk(root, 0)
	return found
}

func buildRepositoryView(ctx context.Context, cfg *config.Config, entry repoEntry) RepositoryView {
	view := RepositoryView{
		Path:       entry.path,
		Name:       filepath.Base(entry.path),
		Source:     entry.source,
		SourcePath: entry.sourcePath,
		Problems:   []string{},
		Notes:      []string{},
	}

	if coreidentity.FindGitRoot(entry.path) != entry.path {
		view.Health = repoHealthMissing
		view.Problems = append(view.Problems, "This folder no longer exists or is no longer a Git repository.")
		if binding := cfg.FindBindingByPath(entry.path); binding != nil {
			view.Alias = binding.User
		}
		return view
	}
	view.Exists = true

	report := corecheck.Inspect(cfg, entry.path)
	applyCheckReport(&view, report)

	if status, err := gitops.GetStatus(ctx, entry.path); err == nil {
		view.Branch = status.Branch
		if status.Detached {
			view.Branch = status.HeadOID
		}
		view.Changes = len(status.Files)
		view.Ahead = status.Ahead
		view.Behind = status.Behind
	}
	return view
}

// applyCheckReport turns the push safety report into health, problems the
// user can fix with one click, and problems they cannot.
func applyCheckReport(view *RepositoryView, report *corecheck.Report) {
	view.Alias = report.ExpectedAlias
	view.IdentitySource = report.ExpectedSource
	view.RemoteURL = report.RemoteURL
	view.Health = repoHealthOK

	if report.Issue != "" {
		view.Health = repoHealthBlocked
		view.Problems = append(view.Problems, report.Issue)
		return
	}

	user := report.ExpectedUser
	view.IdentityName = user.Name
	view.IdentityEmail = user.Email
	view.GitHubUsername = user.GitHubUsername

	if report.IdentityChecked && !report.IdentityMatches {
		current := "no name and email"
		if report.GitName != "" || report.GitEmail != "" {
			current = fmt.Sprintf("%s <%s>", report.GitName, report.GitEmail)
		}
		view.Problems = append(view.Problems, fmt.Sprintf("Commits here would use %s instead of %s <%s>.", current, user.Name, user.Email))
		view.CanFix = true
	}

	switch {
	case report.RemoteChecked && !report.RemoteMatches:
		view.Problems = append(view.Problems, fmt.Sprintf("The remote address does not go through the SSH key for '%s', so pushes are blocked.", report.ExpectedAlias))
		view.CanFix = true
	case report.RemoteURL == "":
		view.Notes = append(view.Notes, "No remote is configured, so there is nothing to push to yet.")
	case !report.RemoteChecked:
		// bgit can only route GitHub remotes through an identity's SSH key.
		view.Health = repoHealthBlocked
		view.Problems = append(view.Problems, "The remote is not on GitHub. bgit only manages GitHub remotes, and push protection blocks pushes to other hosts.")
	}

	if view.CanFix && view.Health == repoHealthOK {
		view.Health = repoHealthAttention
	}
}

// resolveKnownRepository validates a repository path coming from the UI. Only
// repositories bgit manages (bound, or inside a workspace) can be operated on.
func resolveKnownRepository(cfg *config.Config, path string) (repoEntry, error) {
	repoRoot, err := normalizeDesktopRepoPath(path)
	if err != nil {
		return repoEntry{}, err
	}
	if cfg.FindBindingByPath(repoRoot) != nil {
		return repoEntry{path: repoRoot, source: repoSourceBinding}, nil
	}
	if workspace := cfg.FindWorkspaceByPath(repoRoot); workspace != nil {
		return repoEntry{path: repoRoot, source: repoSourceWorkspace, sourcePath: workspace.Path}, nil
	}
	return repoEntry{}, fmt.Errorf("this repository is not managed by bgit yet. Add it on the Repositories page first")
}

func repositoryActionResult(ctx context.Context, cfg *config.Config, entry repoEntry, message string) (*RepositoryActionResult, error) {
	status, err := loadDesktopStatus()
	if err != nil {
		return nil, err
	}
	view := buildRepositoryView(ctx, cfg, entry)
	return &RepositoryActionResult{Message: message, Repository: &view, Status: status}, nil
}

// addDesktopRepository binds an existing local repository to an identity and
// aligns its Git identity and remote in one step.
func addDesktopRepository(ctx context.Context, request RepoBindingRequest) (*RepositoryActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}
	repoRoot, err := normalizeDesktopRepoPath(request.Path)
	if err != nil {
		return nil, err
	}
	alias := strings.TrimSpace(request.Alias)
	if alias == "" {
		return nil, fmt.Errorf("choose an identity for this repository")
	}

	result, err := corerepo.BindRepository(cfg, repoRoot, alias, true)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("'%s' now uses identity '%s'", filepath.Base(repoRoot), alias)
	if result.ExistingBinding != nil && result.ExistingBinding.User != alias {
		message = fmt.Sprintf("'%s' changed from identity '%s' to '%s'", filepath.Base(repoRoot), result.ExistingBinding.User, alias)
	}

	// Binding alone is not enough: the repository's Git identity and its
	// remote address must match too, or the next push is blocked.
	if _, err := corecheck.Fix(cfg, repoRoot); err != nil {
		return nil, fmt.Errorf("the repository was bound, but could not be aligned: %w", err)
	}

	return repositoryActionResult(ctx, cfg, repoEntry{path: repoRoot, source: repoSourceBinding}, message)
}

// removeDesktopRepository stops managing a bound repository. The stored path
// is matched as-is, so a repository whose folder was deleted can still be
// removed from the list.
func removeDesktopRepository(path string) (*RepositoryActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	path = strings.TrimSpace(path)
	result, err := corerepo.RemoveBinding(cfg, path)
	if err != nil {
		return nil, err
	}
	if result.Binding == nil {
		return nil, fmt.Errorf("this repository is not bound to an identity")
	}

	message := fmt.Sprintf("'%s' is no longer bound to '%s'", filepath.Base(path), result.Binding.User)
	// Clear the repository's own Git identity so it does not keep committing
	// as the old identity. Skipped when the folder is gone.
	if coreidentity.FindGitRoot(path) == path {
		if err := restoreDesktopRepoGitIdentity(cfg, path); err != nil {
			return nil, err
		}
	}

	status, err := loadDesktopStatus()
	if err != nil {
		return nil, err
	}
	return &RepositoryActionResult{Message: message, Status: status}, nil
}

// fixDesktopRepository aligns one repository with its identity.
func fixDesktopRepository(ctx context.Context, path string) (*RepositoryActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}
	entry, err := resolveKnownRepository(cfg, path)
	if err != nil {
		return nil, err
	}

	fixed, err := corecheck.Fix(cfg, entry.path)
	if err != nil {
		return nil, err
	}

	changes := []string{}
	if fixed.IdentityFixed {
		changes = append(changes, "commit identity")
	}
	if fixed.RemoteFixed {
		changes = append(changes, "remote address")
	}
	message := fmt.Sprintf("'%s' already matched its identity", filepath.Base(entry.path))
	if len(changes) > 0 {
		message = fmt.Sprintf("Fixed the %s of '%s'", strings.Join(changes, " and "), filepath.Base(entry.path))
	}
	return repositoryActionResult(ctx, cfg, entry, message)
}

// fixAllDesktopRepositories aligns every repository that has a fixable problem.
func fixAllDesktopRepositories(ctx context.Context) (*RepositoriesResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	fixed := 0
	failures := []string{}
	for _, entry := range knownRepositories(cfg) {
		if coreidentity.FindGitRoot(entry.path) != entry.path {
			continue
		}
		report := corecheck.Inspect(cfg, entry.path)
		needsFix := report.Issue == "" && ((report.IdentityChecked && !report.IdentityMatches) || (report.RemoteChecked && !report.RemoteMatches))
		if !needsFix {
			continue
		}
		if _, err := corecheck.Fix(cfg, entry.path); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", filepath.Base(entry.path), err))
			continue
		}
		fixed++
	}

	result, err := listDesktopRepositories(ctx)
	if err != nil {
		return nil, err
	}
	result.Message = "Fixed " + countOf(fixed, "repository", "repositories")
	if len(failures) > 0 {
		result.Message += fmt.Sprintf("; %d could not be fixed: %s", len(failures), strings.Join(failures, "; "))
	}
	return result, nil
}

// cloneDesktopRepository clones a GitHub repository with the chosen identity
// and binds the clone to it.
func cloneDesktopRepository(ctx context.Context, request CloneRequest) (*RepositoryActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	alias := strings.TrimSpace(request.Alias)
	user := cfg.FindUserByAlias(alias)
	if user == nil {
		return nil, fmt.Errorf("choose an identity to clone with")
	}
	if user.SSHKeyPath == "" {
		return nil, fmt.Errorf("identity '%s' has no SSH key. Edit the identity and choose a key first", alias)
	}

	githubURL, err := normalizeGitHubURL(request.URL)
	if err != nil {
		return nil, err
	}
	cloneURL, err := corerepo.ConvertToBgitURL(githubURL, user.GitHubUsername)
	if err != nil {
		return nil, fmt.Errorf("this is not a GitHub repository address. Use a link such as https://github.com/owner/repository")
	}

	parentDir, err := normalizeDesktopExistingDirPath(request.ParentDir)
	if err != nil {
		return nil, fmt.Errorf("choose the folder to clone into: %w", err)
	}
	folderName := strings.TrimSpace(request.FolderName)
	if folderName == "" {
		folderName, err = corerepo.InferCloneDirectory(githubURL)
		if err != nil {
			return nil, err
		}
	}
	if folderName != filepath.Base(folderName) || folderName == "." || folderName == ".." {
		return nil, fmt.Errorf("the folder name cannot contain path separators")
	}
	destination := filepath.Join(parentDir, folderName)
	if entries, err := os.ReadDir(destination); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("the folder %s already exists and is not empty", destination)
	}

	// The clone goes through the identity's SSH host alias, which must exist
	// in the managed SSH config before Git connects.
	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return nil, err
	}
	loadKeyIntoAgent(user)

	if err := gitops.Clone(ctx, cloneURL, destination); err != nil {
		return nil, err
	}

	repoRoot := coreidentity.FindGitRoot(destination)
	if repoRoot == "" {
		return nil, fmt.Errorf("the clone finished but no repository was found at %s", destination)
	}
	if _, err := corerepo.BindRepository(cfg, repoRoot, alias, true); err != nil {
		return nil, fmt.Errorf("cloned, but could not bind the repository: %w", err)
	}
	if err := git.SetLocalUser(repoRoot, user.Name, user.Email); err != nil {
		return nil, fmt.Errorf("cloned, but could not set the repository identity: %w", err)
	}

	message := fmt.Sprintf("Cloned '%s' as '%s'", filepath.Base(repoRoot), alias)
	return repositoryActionResult(ctx, cfg, repoEntry{path: repoRoot, source: repoSourceBinding}, message)
}

// normalizeGitHubURL accepts the forms people paste: a full HTTPS or SSH
// address, "github.com/owner/repo", or just "owner/repo".
func normalizeGitHubURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("enter the repository address")
	}
	if strings.HasPrefix(value, "git@") {
		return value, nil
	}

	value = strings.TrimSuffix(value, "/")
	if !strings.Contains(value, "://") {
		if strings.HasPrefix(value, "github.com/") || strings.HasPrefix(value, "www.github.com/") {
			value = "https://" + value
		} else if parts := strings.Split(value, "/"); len(parts) == 2 && parts[0] != "" && parts[1] != "" && !strings.ContainsAny(value, " :") {
			value = "https://github.com/" + value
		}
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("this is not a GitHub repository address. Use a link such as https://github.com/owner/repository")
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Host), "www.")
	if host != "github.com" {
		return "", fmt.Errorf("bgit clones GitHub repositories. '%s' is not on GitHub", parsed.Host)
	}

	// Keep only /owner/repo: links copied from the browser often carry extra
	// parts such as /tree/main or /pull/12.
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segments) < 2 || segments[0] == "" || segments[1] == "" {
		return "", fmt.Errorf("the address must include the owner and the repository name")
	}
	return fmt.Sprintf("https://github.com/%s/%s", segments[0], strings.TrimSuffix(segments[1], ".git")), nil
}

// loadKeyIntoAgent adds the identity's key to a running SSH agent. It is a
// convenience for keys with a passphrase; plain keys work without it.
func loadKeyIntoAgent(user *config.User) {
	if user == nil || !coressh.IsSSHAgentRunning() {
		return
	}
	coressh.EnsureKeyLoaded(user)
}

func chooseDesktopCloneFolder(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("desktop runtime is not ready")
	}
	return wailsruntime.OpenDirectoryDialog(ctx, wailsruntime.OpenDialogOptions{
		Title:                "Choose where to put the repository",
		CanCreateDirectories: true,
	})
}

// revealDesktopRepository opens a managed repository folder in the system
// file manager.
func revealDesktopRepository(ctx context.Context, path string) error {
	if ctx == nil {
		return fmt.Errorf("desktop runtime is not ready")
	}
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return err
	}
	entry, err := resolveKnownRepository(cfg, path)
	if err != nil {
		return err
	}
	wailsruntime.BrowserOpenURL(ctx, (&url.URL{Scheme: "file", Path: filepath.ToSlash(entry.path)}).String())
	return nil
}

// cloneDefaults suggests where a clone for an identity should go: the
// identity's workspace folder when it has one, otherwise the home folder.
func cloneDefaults(alias string) (*CloneDefaults, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}
	defaults := &CloneDefaults{}
	for _, workspace := range cfg.GetWorkspaces() {
		if workspace.User == strings.TrimSpace(alias) {
			if info, err := os.Stat(workspace.Path); err == nil && info.IsDir() {
				defaults.ParentDir = workspace.Path
				defaults.FromWorkspace = true
				return defaults, nil
			}
		}
	}
	if home, err := platform.ExpandTilde("~"); err == nil {
		defaults.ParentDir = home
	}
	return defaults, nil
}

// countOf formats a count with the right singular or plural noun.
func countOf(count int, singular, plural string) string {
	if count == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d %s", count, plural)
}
