package uninstall

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/byterings/bgit/core/config"
	coresetup "github.com/byterings/bgit/core/setup"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/testenv"
)

func writeKey(t *testing.T, home, name string) string {
	t.Helper()
	path := testenv.WriteFile(t, filepath.Join(home, ".ssh"), name, "private\n")
	testenv.WriteFile(t, filepath.Join(home, ".ssh"), name+".pub", "public\n")
	return path
}

func TestRunRestoresMachineState(t *testing.T) {
	home := testenv.Isolate(t)
	testenv.Git(t, home, "config", "--global", "user.name", "Before Bgit")
	testenv.Git(t, home, "config", "--global", "user.email", "before@example.com")
	testenv.Git(t, home, "config", "--global", "core.hooksPath", filepath.Join(home, "my-hooks"))
	testenv.WriteFile(t, filepath.Join(home, ".ssh"), "config", "Host example\n  HostName example.com\n")

	generated := writeKey(t, home, "bgit_work-gh")
	own := writeKey(t, home, "own_key")

	cfg, _, err := coresetup.EnsureConfigInitialized()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Users = []config.User{
		{Alias: "work", Name: "Work Dev", Email: "dev@work.example", GitHubUsername: "work-gh", SSHKeyPath: generated},
		{Alias: "home", Name: "Home Dev", Email: "dev@home.example", GitHubUsername: "home-gh", SSHKeyPath: own},
	}
	// What bgit records the first time an identity is activated.
	cfg.ActiveUser = "work"
	cfg.PreviousGitIdentitySet = true
	cfg.PreviousGitName = "Before Bgit"
	cfg.PreviousGitEmail = "before@example.com"
	if _, err := coresetup.Apply(cfg, coresetup.Options{}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	testenv.Git(t, home, "config", "--global", "user.name", "Work Dev")
	testenv.Git(t, home, "config", "--global", "user.email", "dev@work.example")

	bound := testenv.InitRepo(t, home, "bound")
	testenv.Git(t, bound, "remote", "add", "origin", "git@github.com-work-gh:org/bound.git")
	inWorkspace := testenv.InitRepo(t, filepath.Join(home, "ws"), "blog")
	testenv.Git(t, inWorkspace, "remote", "add", "origin", "git@github.com-home-gh:me/blog.git")
	loose := testenv.InitRepo(t, filepath.Join(home, "Projects"), "loose")
	testenv.Git(t, loose, "remote", "add", "origin", "git@github.com-work-gh:org/loose.git")
	plain := testenv.InitRepo(t, filepath.Join(home, "Projects"), "plain")
	testenv.Git(t, plain, "remote", "add", "origin", "https://github.com/org/plain.git")
	if err := cfg.AddBinding(bound, "work"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.AddWorkspace(filepath.Join(home, "ws"), "home"); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	result := Run(Options{RemoveKeys: true, ScanHome: true})

	if result.SSHConfigErr != nil || result.HooksErr != nil || result.ConfigErr != nil {
		t.Fatalf("unexpected errors: %+v", result)
	}
	if len(result.RestoredRepos) != 3 || len(result.FailedRepos) != 0 {
		t.Fatalf("restored=%v failed=%v", result.RestoredRepos, result.FailedRepos)
	}
	for repo, want := range map[string]string{
		bound:       "git@github.com:org/bound.git",
		inWorkspace: "git@github.com:me/blog.git",
		loose:       "git@github.com:org/loose.git",
		plain:       "https://github.com/org/plain.git",
	} {
		if got := testenv.Git(t, repo, "remote", "get-url", "origin"); got != want {
			t.Errorf("%s origin = %q, want %q", filepath.Base(repo), got, want)
		}
	}

	if got := testenv.Git(t, home, "config", "--global", "user.email"); got != "before@example.com" {
		t.Errorf("global user.email = %q", got)
	}
	if got := testenv.Git(t, home, "config", "--global", "core.hooksPath"); got != filepath.Join(home, "my-hooks") {
		t.Errorf("core.hooksPath = %q", got)
	}
	if result.GitIdentity == nil || !result.GitIdentity.HadBackup {
		t.Errorf("git identity result = %+v", result.GitIdentity)
	}

	sshConfig, err := os.ReadFile(filepath.Join(home, ".ssh", "config"))
	if err != nil {
		t.Fatal(err)
	}
	if coressh.HasManagedSection(string(sshConfig)) || string(sshConfig) != "Host example\n  HostName example.com\n" {
		t.Errorf("ssh config after uninstall = %q", sshConfig)
	}

	// Generated keys are deleted; a key the user brought is kept.
	if _, err := os.Stat(generated); !os.IsNotExist(err) {
		t.Error("generated private key should be removed")
	}
	if _, err := os.Stat(generated + ".pub"); !os.IsNotExist(err) {
		t.Error("generated public key should be removed")
	}
	if _, err := os.Stat(own); err != nil {
		t.Errorf("the user's own key must be kept: %v", err)
	}
	if len(result.KeyRemovals) != 2 {
		t.Errorf("key removals = %+v", result.KeyRemovals)
	}

	if _, err := os.Stat(result.ConfigDir); !os.IsNotExist(err) {
		t.Error("config directory should be removed")
	}

	// Running again on a clean machine is harmless.
	again := Run(Options{RemoveKeys: true})
	if again.SSHConfigErr != nil || again.HooksErr != nil || again.ConfigErr != nil || again.GitIdentity != nil {
		t.Fatalf("second run: %+v", again)
	}
}

func TestRunKeepsKeysByDefaultAndReportsLegacyIdentity(t *testing.T) {
	home := testenv.Isolate(t)
	generated := writeKey(t, home, "bgit_work-gh")

	cfg, _, err := coresetup.EnsureConfigInitialized()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Users = []config.User{{Alias: "work", Name: "Work Dev", Email: "dev@work.example", GitHubUsername: "work-gh", SSHKeyPath: generated}}
	cfg.ActiveUser = "work"
	if _, err := coresetup.Apply(cfg, coresetup.Options{}); err != nil {
		t.Fatal(err)
	}
	// An older config never stored the previous identity.
	testenv.Git(t, home, "config", "--global", "user.name", "Work Dev")
	testenv.Git(t, home, "config", "--global", "user.email", "dev@work.example")

	result := Run(Options{})

	if _, err := os.Stat(generated); err != nil {
		t.Errorf("keys must be kept unless removal was requested: %v", err)
	}
	if len(result.KeyRemovals) != 0 {
		t.Errorf("key removals = %+v", result.KeyRemovals)
	}
	identity := result.GitIdentity
	if identity == nil || identity.HadBackup || identity.MatchingUser == nil || identity.MatchingUser.Alias != "work" {
		t.Fatalf("legacy identity result = %+v", identity)
	}
	// The identity is left alone because there is nothing to restore it to.
	if got := testenv.Git(t, home, "config", "--global", "user.email"); got != "dev@work.example" {
		t.Errorf("global user.email = %q", got)
	}
	if got, _ := coresetup.GlobalHooksPath(); got != "" {
		t.Errorf("core.hooksPath should be unset, got %q", got)
	}
}
