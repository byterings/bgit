package identity

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
)

type ResolutionSource = models.ResolutionSource
type Resolution = models.Resolution

const (
	SourceWorkspace = models.SourceWorkspace
	SourceBinding   = models.SourceBinding
	SourceGlobal    = models.SourceGlobal
)

// ResolveIdentity resolves the effective identity for the given path.
func ResolveIdentity(cfg *config.Config, currentPath string) (*Resolution, error) {
	absPath, err := filepath.Abs(currentPath)
	if err != nil {
		absPath = currentPath
	}

	workspace := cfg.FindWorkspaceByPath(absPath)
	if workspace != nil {
		user := cfg.FindUserByAlias(workspace.User)
		if user != nil {
			return &Resolution{
				User:   user,
				Alias:  workspace.User,
				Source: SourceWorkspace,
				Path:   workspace.Path,
			}, nil
		}
	}

	repoRoot := findGitRoot(absPath)
	if repoRoot != "" {
		binding := cfg.FindBindingByPath(repoRoot)
		if binding != nil {
			user := cfg.FindUserByAlias(binding.User)
			if user != nil {
				return &Resolution{
					User:   user,
					Alias:  binding.User,
					Source: SourceBinding,
					Path:   binding.Path,
				}, nil
			}
		}
	}

	if cfg.ActiveUser != "" {
		user := cfg.FindUserByAlias(cfg.ActiveUser)
		if user != nil {
			return &Resolution{
				User:   user,
				Alias:  cfg.ActiveUser,
				Source: SourceGlobal,
			}, nil
		}
	}

	return nil, nil
}

// GetEffectiveUser returns the effective user for the current directory.
func GetEffectiveUser(cfg *config.Config) (*config.User, error) {
	cwd, err := os.Getwd()
	if err != nil {
		if cfg.ActiveUser != "" {
			return cfg.FindUserByAlias(cfg.ActiveUser), nil
		}
		return nil, err
	}

	resolution, err := ResolveIdentity(cfg, cwd)
	if err != nil {
		return nil, err
	}
	if resolution == nil {
		return nil, nil
	}
	return resolution.User, nil
}

// GetEffectiveResolution returns the full resolution for the current directory.
func GetEffectiveResolution(cfg *config.Config) (*Resolution, error) {
	cwd, err := os.Getwd()
	if err != nil {
		if cfg.ActiveUser != "" {
			user := cfg.FindUserByAlias(cfg.ActiveUser)
			if user != nil {
				return &Resolution{
					User:   user,
					Alias:  cfg.ActiveUser,
					Source: SourceGlobal,
				}, nil
			}
		}
		return nil, err
	}

	return ResolveIdentity(cfg, cwd)
}

// IsInsideWorkspace checks if the path is inside any configured workspace.
func IsInsideWorkspace(cfg *config.Config, path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return cfg.FindWorkspaceByPath(absPath) != nil
}

// IsRepoBound checks if the current directory's repo is bound.
func IsRepoBound(cfg *config.Config, path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	repoRoot := findGitRoot(absPath)
	if repoRoot == "" {
		return false
	}
	return cfg.FindBindingByPath(repoRoot) != nil
}

func findGitRoot(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ""
	}

	current := absPath
	for {
		gitDir := filepath.Join(current, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

// IsInsidePath checks if childPath is inside parentPath.
func IsInsidePath(childPath, parentPath string) bool {
	child, err := filepath.Abs(childPath)
	if err != nil {
		return false
	}
	parent, err := filepath.Abs(parentPath)
	if err != nil {
		return false
	}

	if !strings.HasSuffix(parent, string(filepath.Separator)) {
		parent += string(filepath.Separator)
	}

	return strings.HasPrefix(child+string(filepath.Separator), parent) || child == strings.TrimSuffix(parent, string(filepath.Separator))
}

// FindGitRoot is the exported version of findGitRoot.
func FindGitRoot(path string) string {
	return findGitRoot(path)
}
