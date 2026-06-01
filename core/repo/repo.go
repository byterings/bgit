package repo

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/byterings/bgit/core/config"
	coreidentity "github.com/byterings/bgit/core/identity"
	"github.com/byterings/bgit/core/models"
)

type BindingResult = models.BindingResult
type RepoOwner = models.RepoOwner
type WorkspaceUsersResult = models.WorkspaceUsersResult
type WorkspaceRegistrationResult = models.WorkspaceRegistrationResult
type WorkspaceRemovalResult = models.WorkspaceRemovalResult
type BindingRemovalResult = models.BindingRemovalResult
type RepoOwnerResult = models.RepoOwnerResult

// ResolveWorkspaceUsers resolves the aliases to concrete users or returns all users.
func ResolveWorkspaceUsers(cfg *config.Config, aliases []string) (*WorkspaceUsersResult, error) {
	if len(aliases) == 0 {
		return &WorkspaceUsersResult{Users: cfg.Users}, nil
	}

	users := make([]config.User, 0, len(aliases))
	for _, alias := range aliases {
		user := cfg.FindUserByAlias(strings.TrimSpace(alias))
		if user == nil {
			return nil, fmt.Errorf("user '%s' not found", strings.TrimSpace(alias))
		}
		users = append(users, *user)
	}

	return &WorkspaceUsersResult{Users: users}, nil
}

// RegisterWorkspace ensures the workspace binding exists and persists the config.
func RegisterWorkspace(cfg *config.Config, folderPath, userAlias string) (*WorkspaceRegistrationResult, error) {
	if err := cfg.AddWorkspace(folderPath, userAlias); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return &WorkspaceRegistrationResult{
				Workspace:     cfg.FindWorkspaceByPath(folderPath),
				AlreadyExists: true,
			}, nil
		}
		return nil, err
	}
	if err := config.SaveConfig(cfg); err != nil {
		return nil, err
	}
	return &WorkspaceRegistrationResult{
		Workspace: cfg.FindWorkspaceByPath(folderPath),
		Changed:   true,
	}, nil
}

// RemoveWorkspace removes the workspace binding for a user alias and persists the config.
func RemoveWorkspace(cfg *config.Config, userAlias string) (*WorkspaceRemovalResult, error) {
	for _, ws := range cfg.GetWorkspaces() {
		if ws.User == userAlias {
			workspace := ws
			if cfg.RemoveWorkspace(userAlias) {
				if err := config.SaveConfig(cfg); err != nil {
					return nil, err
				}
				return &WorkspaceRemovalResult{
					Workspace: &workspace,
					Changed:   true,
				}, nil
			}
			return &WorkspaceRemovalResult{Workspace: &workspace}, nil
		}
	}

	return &WorkspaceRemovalResult{}, nil
}

// BindRepository binds the repository to the given user and persists the config.
func BindRepository(cfg *config.Config, repoRoot, userAlias string, allowOverride bool) (*BindingResult, error) {
	user := cfg.FindUserByAlias(userAlias)
	if user == nil {
		return nil, fmt.Errorf("user '%s' not found", userAlias)
	}

	result := &BindingResult{
		User: user,
	}

	if existingBinding := cfg.FindBindingByPath(repoRoot); existingBinding != nil {
		existing := *existingBinding
		result.ExistingBinding = &existing
		if existingBinding.User == userAlias {
			result.NoChange = true
			return result, nil
		}
		if !allowOverride {
			return result, nil
		}
	}

	if ws := cfg.FindWorkspaceByPath(repoRoot); ws != nil && ws.User != userAlias {
		workspace := *ws
		result.Workspace = &workspace
	}

	if err := cfg.AddBinding(repoRoot, userAlias); err != nil {
		return nil, err
	}
	if err := config.SaveConfig(cfg); err != nil {
		return nil, err
	}

	return result, nil
}

// RemoveBinding removes the repo binding and persists the config.
func RemoveBinding(cfg *config.Config, repoRoot string) (*BindingRemovalResult, error) {
	binding := cfg.FindBindingByPath(repoRoot)
	if binding == nil {
		return &BindingRemovalResult{}, nil
	}

	existing := *binding
	if cfg.RemoveBinding(repoRoot) {
		if err := config.SaveConfig(cfg); err != nil {
			return nil, err
		}
		return &BindingRemovalResult{
			Binding: &existing,
			Changed: true,
		}, nil
	}

	return &BindingRemovalResult{Binding: &existing}, nil
}

// BindClonedRepository binds the inferred clone path to the given user and persists the config.
func BindClonedRepository(cfg *config.Config, cloneURL, directory, userAlias string) (*BindingResult, error) {
	clonePath := directory
	if clonePath == "" {
		inferred, err := InferCloneDirectory(cloneURL)
		if err != nil {
			return nil, err
		}
		clonePath = inferred
	}

	absPath, err := filepath.Abs(clonePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve clone path: %w", err)
	}

	repoRoot := coreidentity.FindGitRoot(absPath)
	if repoRoot == "" {
		return nil, fmt.Errorf("could not locate cloned git repository at %s", absPath)
	}

	return BindRepository(cfg, repoRoot, userAlias, true)
}

// ResolveRepoOwner returns the binding/workspace/remote owner for a repository.
func ResolveRepoOwner(cfg *config.Config, repoRoot, remoteURL string) *RepoOwnerResult {
	if binding := cfg.FindBindingByPath(repoRoot); binding != nil {
		if bindingUser := cfg.FindUserByAlias(binding.User); bindingUser != nil {
			return &RepoOwnerResult{Owner: &RepoOwner{Alias: binding.User, Source: "binding", User: bindingUser}}
		}
	}

	if workspace := cfg.FindWorkspaceByPath(repoRoot); workspace != nil {
		if workspaceUser := cfg.FindUserByAlias(workspace.User); workspaceUser != nil {
			return &RepoOwnerResult{Owner: &RepoOwner{Alias: workspace.User, Source: "workspace", User: workspaceUser}}
		}
	}

	remoteUsername := ExtractAliasFromURL(remoteURL)
	if remoteUsername == "" {
		return nil
	}

	remoteUser := cfg.FindUserByUsername(remoteUsername)
	if remoteUser == nil {
		return &RepoOwnerResult{}
	}

	return &RepoOwnerResult{Owner: &RepoOwner{Alias: remoteUser.Alias, Source: "remote", User: remoteUser}}
}

// ConvertToBgitURL converts a supported GitHub URL into bgit's SSH host-alias form.
func ConvertToBgitURL(url, sshHostUser string) (string, error) {
	httpsPattern := regexp.MustCompile(`^https?://github\.com/([^/]+)/(.+?)(?:\.git)?$`)
	sshPattern := regexp.MustCompile(`^git@github\.com:([^/]+)/(.+?)(?:\.git)?$`)
	bgitPattern := regexp.MustCompile(`^git@github\.com-([^:]+):([^/]+)/(.+?)(?:\.git)?$`)

	var repoOwner, repoName string

	switch {
	case httpsPattern.MatchString(url):
		matches := httpsPattern.FindStringSubmatch(url)
		repoOwner = matches[1]
		repoName = matches[2]
	case sshPattern.MatchString(url):
		matches := sshPattern.FindStringSubmatch(url)
		repoOwner = matches[1]
		repoName = matches[2]
	case bgitPattern.MatchString(url):
		matches := bgitPattern.FindStringSubmatch(url)
		repoOwner = matches[2]
		repoName = matches[3]
	default:
		return "", fmt.Errorf("unrecognized URL format: %s\nExpected GitHub HTTPS or SSH URL", url)
	}

	repoName = strings.TrimSuffix(repoName, ".git")
	return fmt.Sprintf("git@github.com-%s:%s/%s.git", sshHostUser, repoOwner, repoName), nil
}

// ConvertToStandardURL converts a bgit URL back to standard GitHub SSH form.
func ConvertToStandardURL(url string) (string, error) {
	bgitPattern := regexp.MustCompile(`^git@github\.com-[^:]+:([^/]+)/(.+?)(?:\.git)?$`)
	sshPattern := regexp.MustCompile(`^git@github\.com:([^/]+)/(.+?)(?:\.git)?$`)
	httpsPattern := regexp.MustCompile(`^https?://github\.com/`)

	if matches := bgitPattern.FindStringSubmatch(url); matches != nil {
		user := matches[1]
		repo := strings.TrimSuffix(matches[2], ".git")
		return fmt.Sprintf("git@github.com:%s/%s.git", user, repo), nil
	}
	if sshPattern.MatchString(url) || httpsPattern.MatchString(url) {
		return url, nil
	}

	return "", fmt.Errorf("unrecognized URL format: %s", url)
}

// ExtractAliasFromURL extracts the bgit host alias from a URL if present.
func ExtractAliasFromURL(url string) string {
	bgitPattern := regexp.MustCompile(`^git@github\.com-([^:]+):`)
	if matches := bgitPattern.FindStringSubmatch(url); matches != nil {
		return matches[1]
	}
	return ""
}

// InferCloneDirectory infers the directory name Git would create from a URL.
func InferCloneDirectory(url string) (string, error) {
	httpsPattern := regexp.MustCompile(`^https?://github\.com/[^/]+/(.+?)(?:\.git)?$`)
	sshPattern := regexp.MustCompile(`^git@github\.com:[^/]+/(.+?)(?:\.git)?$`)
	bgitPattern := regexp.MustCompile(`^git@github\.com-[^:]+:[^/]+/(.+?)(?:\.git)?$`)

	var repoName string

	switch {
	case httpsPattern.MatchString(url):
		repoName = httpsPattern.FindStringSubmatch(url)[1]
	case sshPattern.MatchString(url):
		repoName = sshPattern.FindStringSubmatch(url)[1]
	case bgitPattern.MatchString(url):
		repoName = bgitPattern.FindStringSubmatch(url)[1]
	default:
		return "", fmt.Errorf("cannot infer clone directory from URL: %s", url)
	}

	repoName = strings.TrimSuffix(repoName, ".git")
	if repoName == "" {
		return "", fmt.Errorf("cannot infer clone directory from URL: %s", url)
	}

	return repoName, nil
}
