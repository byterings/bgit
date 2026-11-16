// Package uninstall undoes everything bgit changed on a machine: repository
// remotes, the managed SSH config section, the global hooks path, the global
// Git identity, optionally generated SSH keys, and the bgit config directory.
//
// The CLI (`bgit uninstall`) prints progress around these steps; the desktop
// app runs them through Run and shows the returned summary.
package uninstall

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/core/config"
	corerepo "github.com/byterings/bgit/core/repo"
	coresetup "github.com/byterings/bgit/core/setup"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/git"
)

// RestoreConfiguredRepos restores standard GitHub remotes for every bound
// repository and every repository inside a workspace.
func RestoreConfiguredRepos(cfg *config.Config) (fixed []string, failed []string) {
	seen := make(map[string]bool)

	for _, binding := range cfg.GetBindings() {
		restoreRepoRemote(binding.Path, seen, &fixed, &failed)
	}

	for _, workspace := range cfg.GetWorkspaces() {
		workspaceFixed, workspaceFailed := ScanAndRestoreRepos(workspace.Path)
		fixed = AppendUnique(fixed, workspaceFixed...)
		failed = AppendUnique(failed, workspaceFailed...)
	}

	return fixed, failed
}

// ScanAndRestoreRepos walks startPath (and common project folders inside it)
// and restores standard GitHub remotes for repositories that still use a bgit
// host alias.
func ScanAndRestoreRepos(startPath string) (fixed []string, failed []string) {
	scanDirs := []string{startPath}

	commonDirs := []string{"Documents", "Projects", "repos", "src", "code", "work", "dev", "git"}
	for _, dir := range commonDirs {
		fullPath := filepath.Join(startPath, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			scanDirs = append(scanDirs, fullPath)
		}
	}

	visited := make(map[string]bool)

	for _, scanDir := range scanDirs {
		filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
				return filepath.SkipDir
			}

			skipDirs := []string{"node_modules", "vendor", ".cache", ".local", "snap", ".npm", ".cargo"}
			for _, skip := range skipDirs {
				if info.Name() == skip {
					return filepath.SkipDir
				}
			}

			if info.IsDir() && info.Name() == ".git" {
				repoPath := filepath.Dir(path)

				if visited[repoPath] {
					return filepath.SkipDir
				}

				restoreRepoRemote(repoPath, visited, &fixed, &failed)

				return filepath.SkipDir // Don't descend into .git
			}

			return nil
		})
	}

	return fixed, failed
}

func restoreRepoRemote(repoPath string, visited map[string]bool, fixed, failed *[]string) {
	if visited[repoPath] {
		return
	}
	visited[repoPath] = true

	url, err := git.GetRemoteURL(repoPath, "origin")
	if err != nil || url == "" || !strings.Contains(url, "github.com-") {
		return
	}

	newURL, err := corerepo.ConvertToStandardURL(url)
	if err != nil {
		*failed = AppendUnique(*failed, repoPath)
		return
	}

	if err := git.SetRemoteURL(repoPath, "origin", newURL); err != nil {
		*failed = AppendUnique(*failed, repoPath)
	} else {
		*fixed = AppendUnique(*fixed, repoPath)
	}
}

// GitIdentityResult describes what restoring the global Git identity did.
type GitIdentityResult struct {
	// HadBackup is true when the pre-bgit identity was stored and restoring
	// it was attempted.
	HadBackup bool
	NameErr   error
	EmailErr  error
	// InspectErr is set when there was no backup and the current identity
	// could not be read.
	InspectErr error
	// MatchingUser is the bgit user whose identity is still the global Git
	// identity, when there was no backup to restore.
	MatchingUser *config.User
}

// RestoreGitIdentity puts back the global Git identity from before bgit was
// first used. Older configs did not store it; then the current identity is
// left alone and reported if it still belongs to a bgit user.
func RestoreGitIdentity(cfg *config.Config) GitIdentityResult {
	result := GitIdentityResult{}

	if cfg.PreviousGitIdentitySet {
		result.HadBackup = true
		result.NameErr = restoreGitConfigValue("user.name", cfg.PreviousGitName)
		result.EmailErr = restoreGitConfigValue("user.email", cfg.PreviousGitEmail)
		return result
	}

	name, email, err := globalGitIdentity()
	if err != nil {
		result.InspectErr = err
		return result
	}

	for i, user := range cfg.Users {
		if name == user.Name && email == user.Email {
			result.MatchingUser = &cfg.Users[i]
			return result
		}
	}
	return result
}

func restoreGitConfigValue(key, value string) error {
	if value == "" {
		return unsetGlobalConfig(key)
	}

	output, err := git.GlobalCommand("config", "--global", key, value).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(output), err)
	}
	return nil
}

func unsetGlobalConfig(key string) error {
	output, err := git.GlobalCommand("config", "--global", "--unset", key).CombinedOutput()
	if err != nil {
		if exitCode(err) == 5 {
			return nil
		}
		return fmt.Errorf("%s: %w", string(output), err)
	}
	return nil
}

func globalGitIdentity() (name, email string, err error) {
	name, err = globalConfigValue("user.name")
	if err != nil {
		return "", "", err
	}

	email, err = globalConfigValue("user.email")
	if err != nil {
		return "", "", err
	}

	return name, email, nil
}

func globalConfigValue(key string) (string, error) {
	output, err := git.GlobalCommand("config", "--global", "--get", key).Output()
	if err != nil {
		if exitCode(err) == 1 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// KeyRemoval is the outcome of deleting one SSH key file.
type KeyRemoval struct {
	Path string
	// Removed is true when the file existed and was deleted.
	Removed bool
	// Err is set when the file exists but could not be deleted.
	Err error
}

// RemoveGeneratedKeys deletes the SSH key files bgit generated for the
// configured users. Keys the user brought themselves are never touched.
func RemoveGeneratedKeys(cfg *config.Config) []KeyRemoval {
	removals := []KeyRemoval{}

	for _, user := range cfg.Users {
		if user.SSHKeyPath == "" || !IsGeneratedKey(user.SSHKeyPath) {
			continue
		}

		for _, path := range []string{user.SSHKeyPath, user.SSHKeyPath + ".pub"} {
			if err := os.Remove(path); err != nil {
				if !os.IsNotExist(err) {
					removals = append(removals, KeyRemoval{Path: path, Err: err})
				}
				continue
			}
			removals = append(removals, KeyRemoval{Path: path, Removed: true})
		}
	}

	return removals
}

// IsGeneratedKey reports whether a key path follows bgit's generated-key naming.
func IsGeneratedKey(path string) bool {
	base := filepath.Base(path)
	return strings.HasPrefix(base, "bgit_") || strings.HasPrefix(base, "brgit_")
}

// AppendUnique appends the non-empty values that are not in items yet.
func AppendUnique(items []string, more ...string) []string {
	seen := make(map[string]bool, len(items)+len(more))
	for _, item := range items {
		seen[item] = true
	}
	for _, item := range more {
		if item == "" || seen[item] {
			continue
		}
		items = append(items, item)
		seen[item] = true
	}
	return items
}

// Options controls a non-interactive uninstall.
type Options struct {
	// RemoveKeys also deletes SSH keys bgit generated.
	RemoveKeys bool
	// ScanHome also searches the home directory for repositories that still
	// use a bgit remote but are not bound or inside a workspace.
	ScanHome bool
}

// Result summarises a non-interactive uninstall. Each step is attempted even
// when an earlier one fails, so a partial failure never leaves bgit half
// installed without saying what is left.
type Result struct {
	RestoredRepos []string
	FailedRepos   []string

	SSHConfigErr error
	HooksErr     error

	// GitIdentity is nil when no config could be loaded.
	GitIdentity *GitIdentityResult

	KeyRemovals []KeyRemoval

	ConfigDir string
	ConfigErr error
}

// Run removes bgit's changes from this machine without prompting.
func Run(opts Options) *Result {
	result := &Result{}

	// Without a readable config the machine-wide steps still run; only the
	// steps that need the list of users and repositories are skipped.
	var cfg *config.Config
	if loaded, err := config.LoadConfig(); err == nil {
		cfg = loaded
	}

	if cfg != nil {
		result.RestoredRepos, result.FailedRepos = RestoreConfiguredRepos(cfg)
	}
	if opts.ScanHome {
		if home, err := os.UserHomeDir(); err == nil {
			fixed, failed := ScanAndRestoreRepos(home)
			result.RestoredRepos = AppendUnique(result.RestoredRepos, fixed...)
			result.FailedRepos = AppendUnique(result.FailedRepos, failed...)
		}
	}

	result.SSHConfigErr = coressh.RemoveManagedSectionFromFile()
	result.HooksErr = coresetup.RestoreGlobalHooksPath(cfg)

	if cfg != nil {
		identity := RestoreGitIdentity(cfg)
		result.GitIdentity = &identity
		if opts.RemoveKeys {
			result.KeyRemovals = RemoveGeneratedKeys(cfg)
		}
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		result.ConfigErr = err
		return result
	}
	result.ConfigDir = configDir
	result.ConfigErr = os.RemoveAll(configDir)
	return result
}
