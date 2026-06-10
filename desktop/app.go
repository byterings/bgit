package main

import "context"

// App is the Wails backend boundary for the desktop foundation.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
