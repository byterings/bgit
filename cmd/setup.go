package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/core/config"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/git"
	"github.com/byterings/bgit/internal/platform"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

const (
	managedHooksDirName = "hooks"
	prePushHookName     = "pre-push"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "One-time setup for bgit",
	Long: `Run one-time setup for bgit.

This command initializes bgit configuration, installs global pre-push safety checks,
and prepares SSH configuration management.`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	cfg, newlyInitialized, err := ensureConfigInitialized()
	if err != nil {
		return err
	}

	if err := applySetup(cfg, false); err != nil {
		return err
	}

	if newlyInitialized {
		ui.Success("Initialized bgit configuration")
	}
	ui.Success("Setup complete")
	ui.Info("Pre-push safety checks are now enabled globally")
	ui.Info("Optional shell prompt integration: use `bgit prompt --plain` in your shell prompt")
	ui.Info("Next step: run 'bgit add' to add your first identity")
	return nil
}

func ensureConfigInitialized() (*config.Config, bool, error) {
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

func applySetup(cfg *config.Config, silent bool) error {
	if !git.IsGitInstalled() {
		return fmt.Errorf("git is not installed")
	}

	hooksDir, err := getManagedHooksDir()
	if err != nil {
		return err
	}

	if err := platform.MkdirSecure(hooksDir); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	existingHooksPath, err := getGlobalHooksPath()
	if err != nil {
		return fmt.Errorf("failed to inspect git hooks path: %w", err)
	}

	if existingHooksPath != "" &&
		filepath.Clean(existingHooksPath) != filepath.Clean(hooksDir) &&
		!silent {
		ui.Warning(fmt.Sprintf("Replacing existing global hooks path: %s", existingHooksPath))
	}

	if !cfg.PreviousHooksPathSet && filepath.Clean(existingHooksPath) != filepath.Clean(hooksDir) {
		cfg.PreviousHooksPath = existingHooksPath
		cfg.PreviousHooksPathSet = true
	}

	if err := installManagedPrePushHook(hooksDir); err != nil {
		return err
	}

	if err := setGlobalHooksPath(hooksDir); err != nil {
		return err
	}

	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return fmt.Errorf("failed to update SSH config: %w", err)
	}

	cfg.SetupCompleted = true
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if !silent {
		ui.Success(fmt.Sprintf("Installed pre-push hook at %s", filepath.Join(hooksDir, prePushHookName)))
		ui.Success(fmt.Sprintf("Configured global hooks path: %s", hooksDir))
		ui.Success("Updated SSH managed section")
	}

	return nil
}

func getManagedHooksDir() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, managedHooksDirName), nil
}

func installManagedPrePushHook(hooksDir string) error {
	hookPath := filepath.Join(hooksDir, prePushHookName)
	hookContent := `#!/bin/sh
# ---- BEGIN BGIT MANAGED ----
bgit check --from-hook
status=$?
if [ $status -ne 0 ]; then
  echo "bgit: push blocked by safety checks"
fi
exit $status
# ---- END BGIT MANAGED ----
`

	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to write pre-push hook: %w", err)
	}

	if err := os.Chmod(hookPath, 0755); err != nil {
		return fmt.Errorf("failed to make pre-push hook executable: %w", err)
	}

	return nil
}

func getGlobalHooksPath() (string, error) {
	cmd := exec.Command("git", "config", "--global", "--get", "core.hooksPath")
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func setGlobalHooksPath(path string) error {
	cmd := exec.Command("git", "config", "--global", "core.hooksPath", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set core.hooksPath: %s: %w", string(output), err)
	}
	return nil
}

func clearManagedHooksPath() error {
	hooksDir, err := getManagedHooksDir()
	if err != nil {
		return err
	}

	currentPath, err := getGlobalHooksPath()
	if err != nil {
		return err
	}

	if currentPath == "" || filepath.Clean(currentPath) != filepath.Clean(hooksDir) {
		return nil
	}

	cmd := exec.Command("git", "config", "--global", "--unset", "core.hooksPath")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unset core.hooksPath: %s: %w", string(output), err)
	}
	return nil
}
