package cmd

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/byterings/bgit/internal/config"
)

func TestRestoreManagedHooksPathRestoresPreviousValue(t *testing.T) {
	requireGit(t)
	home := t.TempDir()
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)

	hooksDir, err := getManagedHooksDir()
	if err != nil {
		t.Fatal(err)
	}
	previousHooksPath := filepath.Join(home, "custom-hooks")
	gitConfigSet(t, "core.hooksPath", hooksDir)

	cfg := &config.Config{
		HooksPathBackedUp: true,
		PreviousHooksPath: previousHooksPath,
	}

	if err := restoreManagedHooksPath(cfg, false); err != nil {
		t.Fatalf("restoreManagedHooksPath() error = %v", err)
	}

	if got := gitConfigGet(t, "core.hooksPath"); got != previousHooksPath {
		t.Fatalf("core.hooksPath = %q, want %q", got, previousHooksPath)
	}
}

func TestRestoreManagedHooksPathUnsetsWhenNoPreviousValue(t *testing.T) {
	requireGit(t)
	home := t.TempDir()
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)

	hooksDir, err := getManagedHooksDir()
	if err != nil {
		t.Fatal(err)
	}
	gitConfigSet(t, "core.hooksPath", hooksDir)

	cfg := &config.Config{HooksPathBackedUp: true}

	if err := restoreManagedHooksPath(cfg, false); err != nil {
		t.Fatalf("restoreManagedHooksPath() error = %v", err)
	}

	if got := gitConfigGet(t, "core.hooksPath"); got != "" {
		t.Fatalf("core.hooksPath = %q, want empty", got)
	}
}

func TestRestoreGitIdentityRestoresBackedUpIdentity(t *testing.T) {
	requireGit(t)
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)

	gitConfigSet(t, "user.name", "Work User")
	gitConfigSet(t, "user.email", "work@example.com")

	cfg := &config.Config{
		GitIdentityBackedUp: true,
		PreviousGitName:     "Normal User",
		PreviousGitEmail:    "normal@example.com",
		Users: []config.User{
			{Name: "Work User", Email: "work@example.com"},
		},
	}

	if err := restoreGitIdentity(cfg, false, false); err != nil {
		t.Fatalf("restoreGitIdentity() error = %v", err)
	}

	if got := gitConfigGet(t, "user.name"); got != "Normal User" {
		t.Fatalf("user.name = %q, want Normal User", got)
	}
	if got := gitConfigGet(t, "user.email"); got != "normal@example.com" {
		t.Fatalf("user.email = %q, want normal@example.com", got)
	}
}

func TestRestoreGitIdentityLeavesNonBgitIdentityUnchanged(t *testing.T) {
	requireGit(t)
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)

	gitConfigSet(t, "user.name", "Manual User")
	gitConfigSet(t, "user.email", "manual@example.com")

	cfg := &config.Config{
		GitIdentityBackedUp: true,
		PreviousGitName:     "Normal User",
		PreviousGitEmail:    "normal@example.com",
		Users: []config.User{
			{Name: "Work User", Email: "work@example.com"},
		},
	}

	if err := restoreGitIdentity(cfg, false, false); err != nil {
		t.Fatalf("restoreGitIdentity() error = %v", err)
	}

	if got := gitConfigGet(t, "user.name"); got != "Manual User" {
		t.Fatalf("user.name = %q, want Manual User", got)
	}
	if got := gitConfigGet(t, "user.email"); got != "manual@example.com" {
		t.Fatalf("user.email = %q, want manual@example.com", got)
	}
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func gitConfigSet(t *testing.T, key, value string) {
	t.Helper()
	cmd := exec.Command("git", "config", "--global", key, value)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config set %s failed: %s: %v", key, string(output), err)
	}
}

func gitConfigGet(t *testing.T, key string) string {
	t.Helper()
	cmd := exec.Command("git", "config", "--global", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return ""
		}
		t.Fatalf("git config get %s failed: %v", key, err)
	}
	return strings.TrimSpace(string(output))
}
