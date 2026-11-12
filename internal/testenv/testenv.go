// Package testenv gives tests an isolated HOME, Git config, and helpers for
// creating disposable repositories. It is only imported by tests.
package testenv

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Isolate points HOME and Git's global/system config at a temp directory so a
// test can never read or change the developer's real bgit, Git, or SSH state.
func Isolate(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, ".gitconfig"))
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	// A hooks path or signing setup from the host must not leak into tests.
	t.Setenv("GIT_CONFIG_COUNT", "0")
	t.Setenv("SSH_AUTH_SOCK", "")
	return home
}

// Git runs a git command in dir and fails the test on error.
func Git(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

// InitRepo creates an empty repository on branch main and returns its path.
func InitRepo(t *testing.T, parent, name string) string {
	t.Helper()

	repo := filepath.Join(parent, name)
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}
	Git(t, repo, "init", "-q", "-b", "main")
	// Resolve symlinks (macOS temp dirs) so paths match what Git reports.
	if resolved, err := filepath.EvalSymlinks(repo); err == nil {
		repo = resolved
	}
	return repo
}

// WriteFile writes content to a file inside dir, creating parent directories.
func WriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
	return path
}
