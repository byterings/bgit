package main

import (
	"context"

	"github.com/byterings/bgit/core/models"
)

// App is the Wails backend boundary for the desktop foundation.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

// context returns the Wails context, or a background context before startup
// (and in the dev harness, which has no Wails runtime).
func (a *App) context() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	repairDesktopHook()
}

func (a *App) GetStatus() (*DesktopStatus, error) {
	return loadDesktopStatus()
}

func (a *App) ListIdentities() ([]IdentityView, error) {
	status, err := loadDesktopStatus()
	if err != nil {
		return nil, err
	}
	return status.Identities, nil
}

func (a *App) GetActiveIdentity() (*IdentityView, error) {
	status, err := loadDesktopStatus()
	if err != nil {
		return nil, err
	}
	return status.ActiveIdentity, nil
}

func (a *App) AddIdentity(request IdentityRequest) (*IdentityActionResult, error) {
	return addDesktopIdentity(request)
}

func (a *App) UpdateIdentity(request UpdateIdentityRequest) (*IdentityActionResult, error) {
	return updateDesktopIdentity(request)
}

func (a *App) ActivateIdentity(alias string) (*IdentityActionResult, error) {
	return activateDesktopIdentity(alias)
}

func (a *App) DeleteIdentity(request DeleteIdentityRequest) (*IdentityActionResult, error) {
	return deleteDesktopIdentity(request)
}

func (a *App) ChooseRepositoryPath() (string, error) {
	return chooseDesktopRepositoryPath(a.ctx)
}

func (a *App) BindRepository(request RepoBindingRequest) (*RepoBindingActionResult, error) {
	return bindDesktopRepository(request)
}

func (a *App) RemoveRepositoryBinding(path string) (*RepoBindingActionResult, error) {
	return removeDesktopRepositoryBinding(path)
}

func (a *App) ChooseIdentityPath() (string, error) {
	return chooseDesktopIdentityPath(a.ctx)
}

func (a *App) ChooseSSHPrivateKeyPath() (string, error) {
	return chooseDesktopSSHPrivateKeyPath(a.ctx)
}

func (a *App) DetectIdentity(request IdentityDetectionRequest) (*IdentityDetectionResult, error) {
	return detectDesktopIdentity(request)
}

func (a *App) SyncDetectedIdentity(request IdentityDetectionRequest) (*IdentityDetectionActionResult, error) {
	return syncDesktopDetectedIdentity(request)
}

func (a *App) GetDoctorStatus() (*DoctorStatus, error) {
	return loadDoctorStatus()
}

func (a *App) ChooseExportArchivePath() (string, error) {
	return chooseDesktopExportArchivePath(a.ctx)
}

func (a *App) ChooseImportArchivePath() (string, error) {
	return chooseDesktopImportArchivePath(a.ctx)
}

func (a *App) ExportBackup(request BackupExportRequest) (*BackupActionResult, error) {
	return exportDesktopBackup(request)
}

func (a *App) ImportBackup(request BackupImportRequest) (*BackupActionResult, error) {
	return importDesktopBackup(request)
}

func (a *App) GetOnboardingState() (*OnboardingState, error) {
	return loadOnboardingState()
}

func (a *App) RunSetup() (*OnboardingState, error) {
	return runDesktopSetup()
}

func (a *App) GetAppInfo() *AppInfo {
	return loadAppInfo()
}

func (a *App) OpenGitHubKeySettings() error {
	return openDesktopExternalURL(a.ctx, githubNewSSHKeyURL)
}

func (a *App) OpenExternalURL(url string) error {
	return openDesktopExternalURL(a.ctx, url)
}

func (a *App) TestIdentityConnection(alias string) (*ConnectionTestResult, error) {
	return testDesktopIdentityConnection(alias)
}

func (a *App) ListRepositories() (*RepositoriesResult, error) {
	return listDesktopRepositories(a.context())
}

func (a *App) AddRepository(request RepoBindingRequest) (*RepositoryActionResult, error) {
	return addDesktopRepository(a.context(), request)
}

func (a *App) CloneRepository(request CloneRequest) (*RepositoryActionResult, error) {
	return cloneDesktopRepository(a.context(), request)
}

func (a *App) GetCloneDefaults(alias string) (*CloneDefaults, error) {
	return cloneDefaults(alias)
}

func (a *App) ChooseCloneFolder() (string, error) {
	return chooseDesktopCloneFolder(a.ctx)
}

func (a *App) FixRepository(path string) (*RepositoryActionResult, error) {
	return fixDesktopRepository(a.context(), path)
}

func (a *App) FixAllRepositories() (*RepositoriesResult, error) {
	return fixAllDesktopRepositories(a.context())
}

func (a *App) RevealRepository(path string) error {
	return revealDesktopRepository(a.ctx, path)
}

func (a *App) RemoveRepository(path string) (*RepositoryActionResult, error) {
	return removeDesktopRepository(path)
}

func (a *App) GetRepositoryState(path string) (*RepoState, error) {
	return getDesktopRepoState(a.context(), path)
}

func (a *App) GetFileDiff(request FileDiffRequest) (*models.GitDiff, error) {
	return getDesktopFileDiff(a.context(), request)
}

func (a *App) CommitChanges(request CommitChangesRequest) (*GitActionResult, error) {
	return commitDesktopChanges(a.context(), request)
}

func (a *App) DiscardChanges(request DiscardChangesRequest) (*GitActionResult, error) {
	return discardDesktopChanges(a.context(), request)
}

func (a *App) ListBranches(path string) ([]models.GitBranch, error) {
	return listDesktopBranches(a.context(), path)
}

func (a *App) SwitchBranch(request BranchRequest) (*GitActionResult, error) {
	return switchDesktopBranch(a.context(), request)
}

func (a *App) CreateBranch(request BranchRequest) (*GitActionResult, error) {
	return createDesktopBranch(a.context(), request)
}

func (a *App) GetHistory(request HistoryRequest) ([]models.GitCommit, error) {
	return getDesktopHistory(a.context(), request)
}

func (a *App) GetCommitFiles(request CommitFilesRequest) ([]models.GitFileChange, error) {
	return getDesktopCommitFiles(a.context(), request)
}

func (a *App) FetchRepository(path string) (*GitActionResult, error) {
	return fetchDesktopRepository(a.context(), path)
}

func (a *App) PullRepository(path string) (*GitActionResult, error) {
	return pullDesktopRepository(a.context(), path)
}

func (a *App) PushRepository(path string) (*GitActionResult, error) {
	return pushDesktopRepository(a.context(), path)
}

func (a *App) AbortMerge(path string) (*GitActionResult, error) {
	return abortDesktopMerge(a.context(), path)
}

func (a *App) AddWorkspace(request WorkspaceRequest) (*WorkspaceActionResult, error) {
	return addDesktopWorkspace(request)
}

func (a *App) RemoveWorkspace(path string) (*WorkspaceActionResult, error) {
	return removeDesktopWorkspace(path)
}

func (a *App) ChooseWorkspaceFolder() (string, error) {
	return chooseDesktopWorkspaceFolder(a.ctx)
}

func (a *App) UninstallBgit(request UninstallRequest) (*UninstallResult, error) {
	return uninstallDesktop(request)
}
