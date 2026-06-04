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
