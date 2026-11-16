package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/core/config"
	corerepo "github.com/byterings/bgit/core/repo"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// workspaceViews lists configured workspaces with their identity and how many
// repositories they contain.
func workspaceViews(cfg *config.Config) []WorkspaceView {
	views := []WorkspaceView{}
	for _, workspace := range cfg.GetWorkspaces() {
		view := WorkspaceView{Path: workspace.Path, Alias: workspace.User}
		if user := cfg.FindUserByAlias(workspace.User); user != nil {
			view.IdentityName = user.Name
			view.IdentityEmail = user.Email
		} else {
			view.MissingIdentity = true
		}
		if info, err := os.Stat(workspace.Path); err == nil && info.IsDir() {
			view.Exists = true
			view.RepositoryCount = len(findRepositories(workspace.Path, workspaceScanDepth))
		}
		views = append(views, view)
	}
	return views
}

// addDesktopWorkspace registers a folder so every repository inside it uses
// the given identity.
func addDesktopWorkspace(request WorkspaceRequest) (*WorkspaceActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	alias := strings.TrimSpace(request.Alias)
	if alias == "" {
		return nil, fmt.Errorf("choose an identity for this workspace")
	}
	if cfg.FindUserByAlias(alias) == nil {
		return nil, fmt.Errorf("identity '%s' not found", alias)
	}

	folder, err := normalizeDesktopExistingDirPath(request.Path)
	if err != nil {
		return nil, err
	}

	// Nested workspaces would make it unclear which identity wins.
	for _, existing := range cfg.GetWorkspaces() {
		switch {
		case existing.Path == folder:
			return nil, fmt.Errorf("this folder is already a workspace for '%s'. Remove it first to change the identity", existing.User)
		case isPathWithin(folder, existing.Path):
			return nil, fmt.Errorf("this folder is inside the workspace %s", existing.Path)
		case isPathWithin(existing.Path, folder):
			return nil, fmt.Errorf("this folder contains the workspace %s", existing.Path)
		}
	}

	if _, err := corerepo.RegisterWorkspace(cfg, folder, alias); err != nil {
		return nil, err
	}
	return workspaceActionResult(fmt.Sprintf("Repositories in '%s' now use identity '%s'", filepath.Base(folder), alias))
}

// removeDesktopWorkspace removes a workspace by path. The folder and the
// repositories in it are not touched.
func removeDesktopWorkspace(path string) (*WorkspaceActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	// The stored path is used as-is: the folder may no longer exist.
	path = strings.TrimSpace(path)
	if !cfg.RemoveWorkspaceByPath(path) {
		return nil, fmt.Errorf("no workspace is configured at %s", path)
	}
	if err := config.SaveConfig(cfg); err != nil {
		return nil, err
	}
	return workspaceActionResult("Workspace removed. The folder and its repositories were not changed")
}

func workspaceActionResult(message string) (*WorkspaceActionResult, error) {
	status, err := loadDesktopStatus()
	if err != nil {
		return nil, err
	}
	return &WorkspaceActionResult{Message: message, Status: status}, nil
}

func chooseDesktopWorkspaceFolder(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("desktop runtime is not ready")
	}
	return wailsruntime.OpenDirectoryDialog(ctx, wailsruntime.OpenDialogOptions{
		Title:                "Choose a workspace folder",
		CanCreateDirectories: true,
	})
}

// isPathWithin reports whether child is strictly inside parent.
func isPathWithin(child, parent string) bool {
	relative, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}
