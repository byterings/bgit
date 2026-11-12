package main

import (
	"fmt"
	"strings"

	"github.com/byterings/bgit/core/config"
	coresetup "github.com/byterings/bgit/core/setup"
	"github.com/byterings/bgit/internal/git"
)

// loadOnboardingState reports what the first-run wizard still has to do.
// It is read-only and never creates the bgit config.
func loadOnboardingState() (*OnboardingState, error) {
	state := &OnboardingState{
		GitInstalled: git.IsGitInstalled(),
	}
	if state.GitInstalled {
		state.GitVersion = gitVersion()
	}

	exists, err := config.ConfigExists()
	if err != nil {
		return nil, err
	}
	state.Configured = exists

	if exists {
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, err
		}
		state.SetupCompleted = cfg.SetupCompleted
		state.IdentityCount = len(cfg.Users)
		state.ActiveAlias = cfg.ActiveUser
	}

	if state.GitInstalled {
		hook, err := coresetup.InspectHook()
		if err != nil {
			return nil, err
		}
		state.PushProtection = hook.Active()
		state.HooksPath = hook.HooksPath
		state.HooksPathManaged = hook.HooksPathManaged
	}

	state.Needed = !state.GitInstalled || !state.SetupCompleted || state.IdentityCount == 0 || state.ActiveAlias == ""
	return state, nil
}

// runDesktopSetup performs the same one-time setup as `bgit setup`, with this
// binary as the hook's fallback helper.
func runDesktopSetup() (*OnboardingState, error) {
	if !git.IsGitInstalled() {
		return nil, fmt.Errorf("Git is not installed. Install Git, then try again")
	}

	cfg, _, err := coresetup.EnsureConfigInitialized()
	if err != nil {
		return nil, err
	}
	// The user chose to replace a custom hooks folder from the app, which
	// promises to put it back on uninstall, so the replaced path is recorded.
	options := coresetup.Options{HookHelperPath: hookHelperPath(), RememberReplacedHooksPath: true}
	if _, err := coresetup.Apply(cfg, options); err != nil {
		return nil, err
	}
	return loadOnboardingState()
}

// ensureDesktopSetup runs setup when the config exists but setup never
// completed, mirroring the CLI's automatic first-run setup.
func ensureDesktopSetup(cfg *config.Config) error {
	if cfg.SetupCompleted {
		return nil
	}
	if _, err := coresetup.Apply(cfg, coresetup.Options{HookHelperPath: hookHelperPath()}); err != nil {
		return fmt.Errorf("automatic setup failed: %w", err)
	}
	return nil
}

// repairDesktopHook keeps the managed hook usable across app updates, moves,
// and backup imports. Failures are not fatal for the app, so they are ignored.
func repairDesktopHook() {
	exists, err := config.ConfigExists()
	if err != nil || !exists {
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return
	}
	_, _ = coresetup.EnsureHook(cfg, hookHelperPath())
}

func gitVersion() string {
	output, err := git.GlobalCommand("--version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(output)), "git version"))
}
