// Package setup owns one-time bgit setup: config initialization, the managed
// global pre-push hook, and the global hooks path. It is shared by the CLI
// (`bgit setup`) and the desktop app so both install the same state.
package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/git"
	"github.com/byterings/bgit/internal/platform"
)

const (
	// ManagedHooksDirName is the hooks directory name inside the bgit config dir.
	ManagedHooksDirName = "hooks"
	// PrePushHookName is the managed hook file name.
	PrePushHookName = "pre-push"

	hookManagedStart = "# ---- BEGIN BGIT MANAGED ----"
	hookManagedEnd   = "# ---- END BGIT MANAGED ----"
)

type Result = models.SetupResult

// Options controls optional setup behavior.
type Options struct {
	// HookHelperPath is an absolute path to a bgit-compatible binary that the
	// managed hook falls back to when `bgit` is not on PATH. The desktop app
	// passes its own executable here. Empty writes the classic CLI hook.
	HookHelperPath string
}

// defaultHookHelper is used when Options.HookHelperPath is empty. The desktop
// binary sets it at startup so every setup path it reaches, including the
// embedded CLI commands, keeps the hook working without `bgit` on PATH.
var defaultHookHelper string

// SetDefaultHookHelper sets the helper binary written into the managed hook
// when a caller does not pass one explicitly.
func SetDefaultHookHelper(path string) {
	defaultHookHelper = strings.TrimSpace(path)
}

// EnsureConfigInitialized creates the bgit config and backup directories and an
// empty config when none exists, then loads the config.
func EnsureConfigInitialized() (*config.Config, bool, error) {
	exists, err := config.ConfigExists()
	if err != nil {
		return nil, false, err
	}

	newlyInitialized := false
	if !exists {
		if err := config.CreateConfigDir(); err != nil {
			return nil, false, err
		}
		if err := config.CreateBackupDir(); err != nil {
			return nil, false, err
		}
		cfg := config.NewConfig()
		if err := config.SaveConfig(cfg); err != nil {
			return nil, false, err
		}
		newlyInitialized = true
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, false, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, newlyInitialized, nil
}

// Apply installs the managed pre-push hook, points the global hooks path at it,
// refreshes the managed SSH config section, and marks setup as completed.
//
// The returned result is non-nil once the existing hooks path has been
// inspected, even when a later step fails, so callers can still report a
// replaced hooks path.
func Apply(cfg *config.Config, opts Options) (*Result, error) {
	if !git.IsGitInstalled() {
		return nil, fmt.Errorf("git is not installed")
	}

	hooksDir, err := ManagedHooksDir()
	if err != nil {
		return nil, err
	}

	if err := platform.MkdirSecure(hooksDir); err != nil {
		return nil, fmt.Errorf("failed to create hooks directory: %w", err)
	}

	existingHooksPath, err := GlobalHooksPath()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect git hooks path: %w", err)
	}

	result := &Result{
		HooksDir: hooksDir,
		HookPath: filepath.Join(hooksDir, PrePushHookName),
	}
	replacesForeignPath := filepath.Clean(existingHooksPath) != filepath.Clean(hooksDir)
	if existingHooksPath != "" && replacesForeignPath {
		result.ReplacedHooksPath = existingHooksPath
	}

	// Remember the pre-bgit hooks path once so uninstall can restore it.
	if !cfg.PreviousHooksPathSet && replacesForeignPath {
		cfg.PreviousHooksPath = existingHooksPath
		cfg.PreviousHooksPathSet = true
	}

	helperPath := opts.HookHelperPath
	if strings.TrimSpace(helperPath) == "" {
		helperPath = defaultHookHelper
	}
	if err := InstallManagedPrePushHook(hooksDir, helperPath); err != nil {
		return result, err
	}

	if err := SetGlobalHooksPath(hooksDir); err != nil {
		return result, err
	}

	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return result, fmt.Errorf("failed to update SSH config: %w", err)
	}

	cfg.SetupCompleted = true
	if err := config.SaveConfig(cfg); err != nil {
		return result, fmt.Errorf("failed to save config: %w", err)
	}

	return result, nil
}

// ManagedHooksDir returns the bgit-managed global hooks directory.
func ManagedHooksDir() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ManagedHooksDirName), nil
}

// HookScript returns the managed pre-push hook content.
//
// With an empty helperPath this is the classic CLI hook, which requires `bgit`
// on PATH. With a helperPath the hook prefers `bgit` on PATH, falls back to the
// helper binary, and lets the push through with a notice when neither exists,
// so removing the desktop app can never block every push on the machine.
func HookScript(helperPath string) string {
	if strings.TrimSpace(helperPath) == "" {
		return `#!/bin/sh
` + hookManagedStart + `
bgit check --from-hook
status=$?
if [ $status -ne 0 ]; then
  echo "bgit: push blocked by safety checks"
fi
exit $status
` + hookManagedEnd + `
`
	}

	return `#!/bin/sh
` + hookManagedStart + `
BGIT_HELPER=` + shellQuote(filepath.ToSlash(helperPath)) + `
if command -v bgit >/dev/null 2>&1; then
  bgit check --from-hook
elif [ -x "$BGIT_HELPER" ]; then
  "$BGIT_HELPER" check --from-hook
else
  echo "bgit: safety check skipped because bgit was not found" >&2
  exit 0
fi
status=$?
if [ $status -ne 0 ]; then
  echo "bgit: push blocked by safety checks"
fi
exit $status
` + hookManagedEnd + `
`
}

// InstallManagedPrePushHook writes the managed pre-push hook into hooksDir.
func InstallManagedPrePushHook(hooksDir, helperPath string) error {
	hookPath := filepath.Join(hooksDir, PrePushHookName)

	if err := os.WriteFile(hookPath, []byte(HookScript(helperPath)), 0755); err != nil {
		return fmt.Errorf("failed to write pre-push hook: %w", err)
	}

	if err := os.Chmod(hookPath, 0755); err != nil {
		return fmt.Errorf("failed to make pre-push hook executable: %w", err)
	}

	return nil
}

// InstalledHookHelperPath returns the helper path baked into the installed
// managed hook, or "" when the hook is missing or is the classic CLI hook.
func InstalledHookHelperPath() (string, error) {
	hooksDir, err := ManagedHooksDir()
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(filepath.Join(hooksDir, PrePushHookName))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	for _, line := range strings.Split(string(content), "\n") {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), "BGIT_HELPER="); ok {
			return shellUnquote(value), nil
		}
	}
	return "", nil
}

// IsManagedHooksPath reports whether path is a bgit-managed hooks directory.
func IsManagedHooksPath(path string) bool {
	cleaned := filepath.ToSlash(filepath.Clean(path))
	if strings.Contains(cleaned, "/.bgit/hooks") || strings.HasSuffix(cleaned, ".bgit/hooks") {
		return true
	}

	content, err := os.ReadFile(filepath.Join(path, PrePushHookName))
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "BEGIN BGIT MANAGED") ||
		coressh.HasManagedSection(string(content))
}

// GlobalHooksPath returns the global core.hooksPath, or "" when it is unset.
func GlobalHooksPath() (string, error) {
	cmd := git.GlobalCommand("config", "--global", "--get", "core.hooksPath")
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// SetGlobalHooksPath sets the global core.hooksPath.
func SetGlobalHooksPath(path string) error {
	cmd := git.GlobalCommand("config", "--global", "core.hooksPath", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set core.hooksPath: %s: %w", string(output), err)
	}
	return nil
}

// UnsetGlobalHooksPath removes the global core.hooksPath. A missing key is not
// an error.
func UnsetGlobalHooksPath() error {
	cmd := git.GlobalCommand("config", "--global", "--unset", "core.hooksPath")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 5 {
			return nil
		}
		return fmt.Errorf("%s: %w", string(output), err)
	}
	return nil
}

// ClearManagedHooksPath unsets the global hooks path only when it still points
// at the bgit-managed hooks directory.
func ClearManagedHooksPath() error {
	hooksDir, err := ManagedHooksDir()
	if err != nil {
		return err
	}

	currentPath, err := GlobalHooksPath()
	if err != nil {
		return err
	}

	if currentPath == "" || filepath.Clean(currentPath) != filepath.Clean(hooksDir) {
		return nil
	}

	if err := UnsetGlobalHooksPath(); err != nil {
		return fmt.Errorf("failed to unset core.hooksPath: %w", err)
	}
	return nil
}

// RestoreGlobalHooksPath puts back the hooks path that was configured before
// bgit setup, when the current one is still managed by bgit.
func RestoreGlobalHooksPath(cfg *config.Config) error {
	currentPath, err := GlobalHooksPath()
	if err != nil {
		return err
	}

	if currentPath == "" || !IsManagedHooksPath(currentPath) {
		return nil
	}

	if cfg != nil && cfg.PreviousHooksPathSet && cfg.PreviousHooksPath != "" {
		return SetGlobalHooksPath(cfg.PreviousHooksPath)
	}

	return UnsetGlobalHooksPath()
}

// shellQuote wraps a value in single quotes for POSIX sh.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func shellUnquote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		value = value[1 : len(value)-1]
	}
	return strings.ReplaceAll(value, `'\''`, "'")
}
