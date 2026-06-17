package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/byterings/bgit/core/config"
	coreexport "github.com/byterings/bgit/core/export"
	coreidentity "github.com/byterings/bgit/core/identity"
	corerepo "github.com/byterings/bgit/core/repo"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/git"
	"github.com/byterings/bgit/internal/platform"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const desktopBackupVersion = "desktop"

func loadDesktopStatus() (*DesktopStatus, error) {
	exists, err := config.ConfigExists()
	if err != nil {
		return nil, err
	}

	status := &DesktopStatus{Configured: exists}
	if !exists {
		return status, nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	status.SetupCompleted = cfg.SetupCompleted
	status.ActiveAlias = cfg.ActiveUser
	status.IdentityCount = len(cfg.Users)
	status.WorkspaceCount = len(cfg.Workspaces)
	status.BindingCount = len(cfg.Bindings)

	for _, user := range cfg.Users {
		view := identityViewFromUser(user, cfg.ActiveUser)
		status.Identities = append(status.Identities, view)
		if user.Alias == cfg.ActiveUser {
			active := view
			status.ActiveIdentity = &active
		}
	}

	for _, binding := range cfg.Bindings {
		status.Bindings = append(status.Bindings, repoBindingViewFromBinding(cfg, binding))
	}

	resolution, err := coreidentity.GetEffectiveResolution(cfg)
	if err != nil {
		return nil, err
	}
	if resolution != nil {
		status.EffectiveIdentity = &EffectiveIdentityView{
			Alias:  resolution.Alias,
			Source: string(resolution.Source),
			Path:   resolution.Path,
		}
	}

	return status, nil
}

func identityViewFromUser(user config.User, activeAlias string) IdentityView {
	view := IdentityView{
		Alias:              user.Alias,
		Name:               user.Name,
		Email:              user.Email,
		GitHubUsername:     user.GitHubUsername,
		GitHubAvatarURL:    githubAvatarURL(user.GitHubUsername),
		SSHKeyPath:         user.SSHKeyPath,
		SSHPublicKeyStatus: "not_configured",
		Active:             user.Alias == activeAlias,
	}

	if user.SSHKeyPath == "" {
		view.SSHKeyStatus = "not_configured"
		return view
	}
	if _, err := os.Stat(user.SSHKeyPath); err != nil {
		if os.IsNotExist(err) {
			view.SSHKeyStatus = "missing"
		} else {
			view.SSHKeyStatus = "unknown"
		}
		return view
	}

	view.SSHKeyStatus = "available"
	publicKey, err := coressh.GetPublicKeyContent(user.SSHKeyPath)
	if err != nil {
		view.SSHPublicKeyStatus = "missing"
		return view
	}
	view.SSHPublicKey = strings.TrimSpace(publicKey)
	view.SSHPublicKeyStatus = "available"
	return view
}

func githubAvatarURL(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s.png?size=96", username)
}

func repoBindingViewFromBinding(cfg *config.Config, binding config.Binding) RepoBindingView {
	view := RepoBindingView{
		Path:  binding.Path,
		Alias: binding.User,
	}
	user := cfg.FindUserByAlias(binding.User)
	if user == nil {
		view.MissingIdentity = true
		return view
	}
	view.Name = user.Name
	view.Email = user.Email
	view.GitHubUsername = user.GitHubUsername
	return view
}

func addDesktopIdentity(request IdentityRequest) (*IdentityActionResult, error) {
	cfg, err := loadOrCreateDesktopConfig()
	if err != nil {
		return nil, err
	}

	user, err := userFromIdentityRequest(request)
	if err != nil {
		return nil, err
	}

	if request.GenerateSSHKey {
		privateKey, _, err := coressh.GenerateSSHKeySystem(user.GitHubUsername)
		if err != nil {
			return nil, err
		}
		user.SSHKeyPath = privateKey
	} else if user.SSHKeyPath != "" {
		sshKeyPath, err := validateAndNormalizeSSHKeyPath(user.SSHKeyPath)
		if err != nil {
			return nil, err
		}
		user.SSHKeyPath = sshKeyPath
	}

	if _, err := coreidentity.AddUser(cfg, user); err != nil {
		return nil, err
	}
	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return nil, err
	}

	return actionResult(fmt.Sprintf("Identity '%s' added", user.Alias))
}

func updateDesktopIdentity(request UpdateIdentityRequest) (*IdentityActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	alias := strings.TrimSpace(request.Alias)
	if alias == "" {
		return nil, fmt.Errorf("identity alias is required")
	}

	user := cfg.FindUserByAlias(alias)
	if user == nil {
		return nil, fmt.Errorf("identity '%s' not found", alias)
	}

	name := strings.TrimSpace(request.Name)
	email := strings.TrimSpace(request.Email)
	githubUsername := strings.TrimSpace(request.GitHubUsername)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if githubUsername == "" {
		return nil, fmt.Errorf("GitHub username is required")
	}

	sshKeyPath := strings.TrimSpace(request.SSHKeyPath)
	if sshKeyPath != "" {
		sshKeyPath, err = validateAndNormalizeSSHKeyPath(sshKeyPath)
		if err != nil {
			return nil, err
		}
	}

	for i := range cfg.Users {
		if cfg.Users[i].Alias == alias {
			cfg.Users[i].Name = name
			cfg.Users[i].Email = email
			cfg.Users[i].GitHubUsername = githubUsername
			cfg.Users[i].SSHKeyPath = sshKeyPath
			break
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		return nil, err
	}
	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return nil, err
	}

	return actionResult(fmt.Sprintf("Identity '%s' updated", alias))
}

func activateDesktopIdentity(alias string) (*IdentityActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil, fmt.Errorf("identity alias is required")
	}

	result, err := coreidentity.ActivateUser(cfg, alias, coreidentity.LookupByAlias)
	if err != nil {
		return nil, err
	}
	return actionResult(fmt.Sprintf("Identity '%s' activated", result.User.Alias))
}

func deleteDesktopIdentity(request DeleteIdentityRequest) (*IdentityActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	alias := strings.TrimSpace(request.Alias)
	if alias == "" {
		return nil, fmt.Errorf("identity alias is required")
	}

	user := cfg.FindUserByAlias(alias)
	if user == nil {
		return nil, fmt.Errorf("identity '%s' not found", alias)
	}
	sshKeyPath := user.SSHKeyPath

	result, err := coreidentity.DeleteUser(cfg, alias)
	if err != nil {
		return nil, err
	}

	if request.DeleteKeys && sshKeyPath != "" {
		if err := removeIdentityKeyFiles(sshKeyPath); err != nil {
			return nil, err
		}
	}

	message := fmt.Sprintf("Identity '%s' deleted", result.User.Alias)
	if result.ActiveCleared {
		message += "; active identity cleared"
	}
	return actionResult(message)
}

func chooseDesktopRepositoryPath(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("desktop runtime is not ready")
	}

	return wailsruntime.OpenDirectoryDialog(ctx, wailsruntime.OpenDialogOptions{
		Title: "Choose Git repository",
	})
}

func bindDesktopRepository(request RepoBindingRequest) (*RepoBindingActionResult, error) {
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
		return nil, fmt.Errorf("identity is required")
	}

	result, err := corerepo.BindRepository(cfg, repoRoot, alias, true)
	if err != nil {
		return nil, err
	}
	if result.User == nil {
		return nil, fmt.Errorf("identity '%s' not found", alias)
	}
	if err := git.SetLocalUser(repoRoot, result.User.Name, result.User.Email); err != nil {
		return nil, fmt.Errorf("failed to set repository git identity: %w", err)
	}

	message := fmt.Sprintf("Repository bound to '%s'", alias)
	if result.NoChange {
		message = fmt.Sprintf("Repository already bound to '%s'; Git identity refreshed", alias)
	} else if result.ExistingBinding != nil && result.ExistingBinding.User != alias {
		message = fmt.Sprintf("Repository binding changed from '%s' to '%s'", result.ExistingBinding.User, alias)
	}

	return repoBindingActionResult(message, repoRoot)
}

func removeDesktopRepositoryBinding(path string) (*RepoBindingActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	repoRoot, err := normalizeDesktopRepoPath(path)
	if err != nil {
		return nil, err
	}

	result, err := corerepo.RemoveBinding(cfg, repoRoot)
	if err != nil {
		return nil, err
	}
	if result.Binding == nil {
		return repoBindingActionResult("No binding found for this repository", repoRoot)
	}
	return repoBindingActionResult(fmt.Sprintf("Removed binding for '%s'", result.Binding.User), repoRoot)
}

func normalizeDesktopRepoPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("repository path is required")
	}
	expandedPath, err := platform.ExpandTilde(path)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve repository path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("repository path does not exist: %s", absPath)
		}
		return "", fmt.Errorf("cannot inspect repository path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository path must be a directory: %s", absPath)
	}
	repoRoot := coreidentity.FindGitRoot(absPath)
	if repoRoot == "" {
		return "", fmt.Errorf("not a Git repository: %s", absPath)
	}
	return filepath.Abs(repoRoot)
}

func repoBindingActionResult(message, path string) (*RepoBindingActionResult, error) {
	status, err := loadDesktopStatus()
	if err != nil {
		return nil, err
	}
	return &RepoBindingActionResult{
		Message: message,
		Path:    path,
		Status:  status,
	}, nil
}

func loadOrCreateDesktopConfig() (*config.Config, error) {
	exists, err := config.ConfigExists()
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := config.CreateConfigDir(); err != nil {
			return nil, err
		}
		if err := config.CreateBackupDir(); err != nil {
			return nil, err
		}
		if err := config.SaveConfig(config.NewConfig()); err != nil {
			return nil, err
		}
	}
	return config.LoadConfig()
}

func loadExistingDesktopConfig() (*config.Config, error) {
	exists, err := config.ConfigExists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("bgit is not configured on this machine")
	}
	return config.LoadConfig()
}

func userFromIdentityRequest(request IdentityRequest) (config.User, error) {
	user := config.User{
		Alias:          strings.TrimSpace(request.Alias),
		Name:           strings.TrimSpace(request.Name),
		Email:          strings.TrimSpace(request.Email),
		GitHubUsername: strings.TrimSpace(request.GitHubUsername),
		SSHKeyPath:     strings.TrimSpace(request.SSHKeyPath),
	}

	if user.Alias == "" {
		return user, fmt.Errorf("alias is required")
	}
	if user.Name == "" {
		return user, fmt.Errorf("name is required")
	}
	if user.Email == "" {
		return user, fmt.Errorf("email is required")
	}
	if user.GitHubUsername == "" {
		return user, fmt.Errorf("GitHub username is required")
	}
	if request.GenerateSSHKey && user.SSHKeyPath != "" {
		return user, fmt.Errorf("choose either generated SSH key or existing SSH key path")
	}

	return user, nil
}

func validateAndNormalizeSSHKeyPath(path string) (string, error) {
	expandedPath, err := platform.ExpandTilde(path)
	if err != nil {
		return "", err
	}
	if err := coressh.ValidateSSHKeyPath(expandedPath); err != nil {
		return "", err
	}
	return expandedPath, nil
}

func removeIdentityKeyFiles(privateKeyPath string) error {
	if err := os.Remove(privateKeyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete private key: %w", err)
	}
	if err := os.Remove(privateKeyPath + ".pub"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete public key: %w", err)
	}
	return nil
}

func actionResult(message string) (*IdentityActionResult, error) {
	status, err := loadDesktopStatus()
	if err != nil {
		return nil, err
	}
	return &IdentityActionResult{
		Message: message,
		Status:  status,
	}, nil
}

func chooseDesktopExportArchivePath(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("desktop runtime is not ready")
	}

	backupDir, err := config.GetBackupDir()
	if err != nil {
		return "", err
	}
	if err := platform.MkdirSecure(backupDir); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	defaultName := "bgit-export-" + time.Now().UTC().Format("20060102T150405Z") + coreexport.ArchiveExtension

	return wailsruntime.SaveFileDialog(ctx, wailsruntime.SaveDialogOptions{
		Title:                "Save bgit backup archive",
		DefaultDirectory:     backupDir,
		DefaultFilename:      defaultName,
		CanCreateDirectories: true,
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "bgit backup archives (*.bgit)", Pattern: "*.bgit"},
		},
	})
}

func chooseDesktopImportArchivePath(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("desktop runtime is not ready")
	}

	backupDir, _ := config.GetBackupDir()
	if _, err := os.Stat(backupDir); err != nil {
		backupDir = ""
	}
	return wailsruntime.OpenFileDialog(ctx, wailsruntime.OpenDialogOptions{
		Title:            "Choose bgit backup archive",
		DefaultDirectory: backupDir,
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "bgit backup archives (*.bgit)", Pattern: "*.bgit"},
		},
	})
}

func exportDesktopBackup(request BackupExportRequest) (*BackupActionResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	password := strings.TrimSpace(request.Password)
	if password == "" {
		return nil, fmt.Errorf("export password is required")
	}
	if password != request.ConfirmPassword {
		return nil, fmt.Errorf("export passwords do not match")
	}

	result, err := coreexport.CreateArchive(cfg, desktopBackupVersion, password)
	if err != nil {
		return nil, err
	}

	archivePath := result.Path
	if selectedPath := strings.TrimSpace(request.OutputPath); selectedPath != "" {
		selectedPath, err = normalizeArchivePath(selectedPath)
		if err != nil {
			return nil, err
		}
		if selectedPath != archivePath {
			if err := moveFile(result.Path, selectedPath); err != nil {
				return nil, fmt.Errorf("failed to save export archive: %w", err)
			}
			archivePath = selectedPath
		}
	}

	status, err := loadDesktopStatus()
	if err != nil {
		return nil, err
	}
	doctor, err := loadDoctorStatus()
	if err != nil {
		return nil, err
	}

	return &BackupActionResult{
		Message:     "Encrypted bgit backup created",
		ArchivePath: archivePath,
		Status:      status,
		Doctor:      doctor,
	}, nil
}

func importDesktopBackup(request BackupImportRequest) (*BackupActionResult, error) {
	archivePath := strings.TrimSpace(request.ArchivePath)
	if archivePath == "" {
		return nil, fmt.Errorf("choose a .bgit archive to import")
	}
	archivePath, err := normalizeArchivePath(archivePath)
	if err != nil {
		return nil, err
	}

	password := strings.TrimSpace(request.Password)
	if password == "" {
		return nil, fmt.Errorf("import password is required")
	}

	result, err := coreexport.ImportArchive(archivePath, password)
	if err != nil {
		return nil, err
	}

	status, err := loadDesktopStatus()
	if err != nil {
		return nil, err
	}
	doctor, err := loadDoctorStatus()
	if err != nil {
		return nil, err
	}

	return &BackupActionResult{
		Message:         "Imported bgit backup",
		ArchivePath:     archivePath,
		UsersCount:      result.UsersCount,
		WorkspacesCount: result.WorkspacesCount,
		BindingsCount:   result.BindingsCount,
		ActiveUser:      result.ActiveUser,
		Status:          status,
		Doctor:          doctor,
	}, nil
}

func normalizeArchivePath(path string) (string, error) {
	expandedPath, err := platform.ExpandTilde(path)
	if err != nil {
		return "", err
	}
	if filepath.Ext(expandedPath) != coreexport.ArchiveExtension {
		expandedPath += coreexport.ArchiveExtension
	}
	return filepath.Abs(expandedPath)
}

func moveFile(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	if err := os.Rename(source, destination); err == nil {
		return nil
	}

	if err := copyFile(source, destination); err != nil {
		return err
	}
	return os.Remove(source)
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	return output.Sync()
}

func loadDoctorStatus() (*DoctorStatus, error) {
	status := &DoctorStatus{}

	cfg, exists, configSection := desktopConfigChecks()
	status.Configured = exists
	status.addSection(configSection)

	if !exists || cfg == nil {
		return status, nil
	}

	status.addSection(desktopSSHChecks(cfg))
	status.addSection(desktopSSHAgentChecks())
	status.addSection(desktopGitChecks(cfg))

	return status, nil
}

func desktopConfigChecks() (*config.Config, bool, DoctorSection) {
	section := DoctorSection{Title: "Config"}

	exists, err := config.ConfigExists()
	if err != nil {
		section.Checks = append(section.Checks, doctorFail(fmt.Sprintf("Error checking config: %v", err), ""))
		return nil, false, section
	}
	if !exists {
		section.Checks = append(section.Checks, doctorWarn("Config file not found", "Run: bgit setup"))
		return nil, false, section
	}

	section.Checks = append(section.Checks, doctorPass("Config file exists"))

	cfg, err := config.LoadConfig()
	if err != nil {
		section.Checks = append(section.Checks, doctorFail(fmt.Sprintf("Config file invalid: %v", err), ""))
		return nil, true, section
	}

	section.Checks = append(section.Checks, doctorPass("Config file valid"))
	if len(cfg.Users) == 0 {
		section.Checks = append(section.Checks, doctorWarn("No identities configured", "Add an identity in the desktop app or run: bgit add"))
	} else {
		section.Checks = append(section.Checks, doctorPass(fmt.Sprintf("%d identity(s) configured", len(cfg.Users))))
	}

	if cfg.ActiveUser == "" {
		section.Checks = append(section.Checks, doctorWarn("No active identity set", "Use an identity from the desktop app or run: bgit use <alias>"))
	} else if cfg.FindUserByAlias(cfg.ActiveUser) == nil {
		section.Checks = append(section.Checks, doctorFail(fmt.Sprintf("Active identity '%s' not found in config", cfg.ActiveUser), ""))
	} else {
		section.Checks = append(section.Checks, doctorPass(fmt.Sprintf("Active identity: %s", cfg.ActiveUser)))
	}

	return cfg, true, section
}

func desktopSSHChecks(cfg *config.Config) DoctorSection {
	section := DoctorSection{Title: "SSH Setup"}

	sshDir, err := platform.GetSSHDir()
	if err != nil {
		section.Checks = append(section.Checks, doctorFail(fmt.Sprintf("Cannot determine SSH directory: %v", err), ""))
		return section
	}

	info, err := os.Stat(sshDir)
	if os.IsNotExist(err) {
		section.Checks = append(section.Checks, doctorWarn("SSH directory does not exist", fmt.Sprintf("Run: mkdir -p %s && chmod 700 %s", sshDir, sshDir)))
	} else if err != nil {
		section.Checks = append(section.Checks, doctorFail(fmt.Sprintf("Cannot inspect SSH directory: %v", err), ""))
	} else if runtime.GOOS != "windows" && info.Mode().Perm() != 0700 {
		section.Checks = append(section.Checks, doctorWarn(fmt.Sprintf("SSH directory has permissions %o; expected 700", info.Mode().Perm()), fmt.Sprintf("chmod 700 %s", sshDir)))
	} else {
		section.Checks = append(section.Checks, doctorPass("SSH directory permissions OK"))
	}

	for _, user := range cfg.Users {
		section.Checks = append(section.Checks, desktopSSHKeyCheck(user))
	}

	sshConfigPath, err := coressh.GetSSHConfigPath()
	if err != nil {
		section.Checks = append(section.Checks, doctorFail(fmt.Sprintf("Cannot determine SSH config path: %v", err), ""))
		return section
	}

	content, err := os.ReadFile(sshConfigPath)
	if os.IsNotExist(err) {
		section.Checks = append(section.Checks, doctorWarn("SSH config file not found", "Run: bgit sync --fix"))
	} else if err != nil {
		section.Checks = append(section.Checks, doctorFail(fmt.Sprintf("Cannot read SSH config: %v", err), ""))
	} else if coressh.HasManagedSection(string(content)) {
		section.Checks = append(section.Checks, doctorPass("SSH config has bgit managed entries"))
	} else {
		section.Checks = append(section.Checks, doctorWarn("SSH config missing bgit managed entries", "Run: bgit sync --fix"))
	}

	return section
}

func desktopSSHKeyCheck(user config.User) DoctorCheckView {
	if user.SSHKeyPath == "" {
		return doctorWarn(fmt.Sprintf("No SSH key path for '%s'", user.Alias), fmt.Sprintf("Update '%s' in the desktop app or run: bgit update %s --ssh-key <path>", user.Alias, user.Alias))
	}

	info, err := os.Stat(user.SSHKeyPath)
	if os.IsNotExist(err) {
		return doctorWarn(fmt.Sprintf("SSH key missing for '%s': %s", user.Alias, user.SSHKeyPath), fmt.Sprintf("Update '%s' with a valid SSH key", user.Alias))
	}
	if err != nil {
		return doctorFail(fmt.Sprintf("Cannot inspect SSH key for '%s': %v", user.Alias, err), "")
	}
	if info.IsDir() {
		return doctorFail(fmt.Sprintf("SSH key path for '%s' is a directory", user.Alias), "")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		return doctorWarn(fmt.Sprintf("SSH key '%s' has permissions %o; expected 600", user.Alias, info.Mode().Perm()), fmt.Sprintf("chmod 600 %s", user.SSHKeyPath))
	}

	return doctorPass(fmt.Sprintf("SSH key '%s' exists", user.Alias))
}

func desktopSSHAgentChecks() DoctorSection {
	section := DoctorSection{Title: "SSH Agent"}

	if !coressh.IsSSHAgentRunning() {
		section.Checks = append(section.Checks, doctorWarn("SSH agent not running", "Run: eval $(ssh-agent)"))
		return section
	}

	section.Checks = append(section.Checks, doctorPass("SSH agent running"))
	output, err := coressh.ListAgentKeys()
	if err != nil {
		if output != nil && strings.Contains(output.Output, "no identities") {
			section.Checks = append(section.Checks, doctorWarn("No keys loaded in SSH agent", "Run: ssh-add ~/.ssh/bgit_*"))
		} else {
			section.Checks = append(section.Checks, doctorWarn("Could not list SSH agent keys", "Run: ssh-add -l"))
		}
		return section
	}

	trimmed := strings.TrimSpace(output.Output)
	if trimmed == "" {
		section.Checks = append(section.Checks, doctorWarn("No keys loaded in SSH agent", "Run: ssh-add ~/.ssh/bgit_*"))
		return section
	}
	section.Checks = append(section.Checks, doctorPass(fmt.Sprintf("%d key(s) loaded in SSH agent", len(strings.Split(trimmed, "\n")))))
	return section
}

func desktopGitChecks(cfg *config.Config) DoctorSection {
	section := DoctorSection{Title: "Git Config"}

	if !git.IsGitInstalled() {
		section.Checks = append(section.Checks, doctorFail("Git is not installed", "Install Git and run bgit setup"))
		return section
	}
	section.Checks = append(section.Checks, doctorPass("Git is installed"))

	if cfg.ActiveUser == "" {
		section.Checks = append(section.Checks, doctorWarn("No active identity to compare with global Git config", "Use an identity from the desktop app"))
		return section
	}

	user := cfg.FindUserByAlias(cfg.ActiveUser)
	if user == nil {
		section.Checks = append(section.Checks, doctorFail(fmt.Sprintf("Active identity '%s' not found", cfg.ActiveUser), ""))
		return section
	}

	name, email, err := git.GetGlobalUser()
	if err != nil {
		section.Checks = append(section.Checks, doctorFail(fmt.Sprintf("Could not read global Git identity: %v", err), ""))
		return section
	}
	if name == user.Name {
		section.Checks = append(section.Checks, doctorPass(fmt.Sprintf("user.name = %s", name)))
	} else {
		section.Checks = append(section.Checks, doctorWarn(fmt.Sprintf("user.name mismatch: '%s' expected '%s'", name, user.Name), "Run: bgit sync --fix"))
	}
	if email == user.Email {
		section.Checks = append(section.Checks, doctorPass(fmt.Sprintf("user.email = %s", email)))
	} else {
		section.Checks = append(section.Checks, doctorWarn(fmt.Sprintf("user.email mismatch: '%s' expected '%s'", email, user.Email), "Run: bgit sync --fix"))
	}

	return section
}

func (status *DoctorStatus) addSection(section DoctorSection) {
	for _, check := range section.Checks {
		switch check.Level {
		case "fail":
			status.Errors++
		case "warn":
			status.Warnings++
		}
	}
	status.Sections = append(status.Sections, section)
}

func doctorPass(message string) DoctorCheckView {
	return DoctorCheckView{Level: "pass", Message: message}
}

func doctorWarn(message, fix string) DoctorCheckView {
	return DoctorCheckView{Level: "warn", Message: message, Fix: fix}
}

func doctorFail(message, fix string) DoctorCheckView {
	return DoctorCheckView{Level: "fail", Message: message, Fix: fix}
}
