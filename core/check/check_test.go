package check

import (
	"testing"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/internal/testenv"
)

func testConfig() *config.Config {
	cfg := config.NewConfig()
	cfg.Users = []config.User{
		{Alias: "work", Name: "Work Dev", Email: "dev@work.example", GitHubUsername: "work-gh"},
		{Alias: "personal", Name: "Home Dev", Email: "dev@home.example", GitHubUsername: "home-gh"},
	}
	return cfg
}

func TestResolveExpected(t *testing.T) {
	const repo = "/tmp/bgit-check-test/repo"

	t.Run("no owner and no active user", func(t *testing.T) {
		expectation := ResolveExpected(testConfig(), repo, "")
		if expectation.Alias != "" || expectation.User != nil {
			t.Fatalf("expected empty expectation, got %+v", expectation)
		}
	})

	t.Run("falls back to active user", func(t *testing.T) {
		cfg := testConfig()
		cfg.ActiveUser = "personal"
		expectation := ResolveExpected(cfg, repo, "git@github.com:org/repo.git")
		if expectation.Alias != "personal" || expectation.Source != SourceActive || expectation.OwnerAlias != "" {
			t.Fatalf("unexpected expectation: %+v", expectation)
		}
	})

	t.Run("binding wins over active user", func(t *testing.T) {
		cfg := testConfig()
		cfg.ActiveUser = "personal"
		if err := cfg.AddBinding(repo, "work"); err != nil {
			t.Fatal(err)
		}
		expectation := ResolveExpected(cfg, repo, "")
		if expectation.Alias != "work" || expectation.Source != "binding" || expectation.OwnerAlias != "work" {
			t.Fatalf("unexpected expectation: %+v", expectation)
		}
	})

	t.Run("remote host alias identifies the owner", func(t *testing.T) {
		cfg := testConfig()
		cfg.ActiveUser = "personal"
		expectation := ResolveExpected(cfg, repo, "git@github.com-work-gh:org/repo.git")
		if expectation.Alias != "work" || expectation.Source != "remote" {
			t.Fatalf("unexpected expectation: %+v", expectation)
		}
	})

	t.Run("active user missing from config", func(t *testing.T) {
		cfg := testConfig()
		cfg.ActiveUser = "ghost"
		expectation := ResolveExpected(cfg, repo, "")
		if expectation.Alias != "ghost" || expectation.User != nil {
			t.Fatalf("unexpected expectation: %+v", expectation)
		}
	})
}

func TestInspectAndFixBoundRepository(t *testing.T) {
	home := testenv.Isolate(t)
	repo := testenv.InitRepo(t, home, "project")
	testenv.Git(t, repo, "remote", "add", "origin", "https://github.com/org/project.git")
	testenv.Git(t, home, "config", "--global", "user.name", "Home Dev")
	testenv.Git(t, home, "config", "--global", "user.email", "dev@home.example")

	cfg := testConfig()
	cfg.ActiveUser = "personal"
	if err := cfg.AddBinding(repo, "work"); err != nil {
		t.Fatal(err)
	}

	report := Inspect(cfg, repo)
	if report.OK() {
		t.Fatal("report should not be OK before the fix")
	}
	if report.ExpectedAlias != "work" || report.ExpectedSource != "binding" {
		t.Fatalf("unexpected expectation: %+v", report)
	}
	if !report.IdentityChecked || report.IdentityMatches {
		t.Fatalf("identity should be checked and mismatched: %+v", report)
	}
	if !report.RemoteChecked || report.RemoteMatches {
		t.Fatalf("remote should be checked and mismatched: %+v", report)
	}
	if report.ExpectedRemoteURL != "git@github.com-work-gh:org/project.git" {
		t.Fatalf("ExpectedRemoteURL = %q", report.ExpectedRemoteURL)
	}

	result, err := Fix(cfg, repo)
	if err != nil {
		t.Fatalf("fix failed: %v", err)
	}
	if !result.IdentityFixed || result.IdentityScope != "repository" || !result.RemoteFixed {
		t.Fatalf("unexpected fix result: %+v", result)
	}
	if !result.Report.OK() {
		t.Fatalf("report should be OK after the fix: %+v", result.Report)
	}

	// The bound repository is fixed locally; the global identity is untouched.
	if got := testenv.Git(t, repo, "config", "--local", "user.email"); got != "dev@work.example" {
		t.Fatalf("local user.email = %q", got)
	}
	if got := testenv.Git(t, home, "config", "--global", "user.email"); got != "dev@home.example" {
		t.Fatalf("global user.email changed to %q", got)
	}
	if got := testenv.Git(t, repo, "remote", "get-url", "origin"); got != "git@github.com-work-gh:org/project.git" {
		t.Fatalf("origin = %q", got)
	}

	if err := VerifyGitIdentity(repo, cfg.FindUserByAlias("work"), "work"); err != nil {
		t.Fatalf("VerifyGitIdentity failed after fix: %v", err)
	}

	again, err := Fix(cfg, repo)
	if err != nil {
		t.Fatalf("second fix failed: %v", err)
	}
	if again.IdentityFixed || again.RemoteFixed {
		t.Fatalf("second fix should change nothing: %+v", again)
	}
}

func TestFixActiveSourceWritesGlobalIdentity(t *testing.T) {
	home := testenv.Isolate(t)
	repo := testenv.InitRepo(t, home, "loose")
	testenv.Git(t, repo, "remote", "add", "origin", "git@github.com:org/loose.git")

	cfg := testConfig()
	cfg.ActiveUser = "personal"

	result, err := Fix(cfg, repo)
	if err != nil {
		t.Fatalf("fix failed: %v", err)
	}
	if result.IdentityScope != "global" {
		t.Fatalf("IdentityScope = %q, want global", result.IdentityScope)
	}
	if got := testenv.Git(t, home, "config", "--global", "user.email"); got != "dev@home.example" {
		t.Fatalf("global user.email = %q", got)
	}
	if !result.Report.OK() {
		t.Fatalf("report should be OK: %+v", result.Report)
	}
}

func TestInspectReportsProblemsItCannotFix(t *testing.T) {
	home := testenv.Isolate(t)
	repo := testenv.InitRepo(t, home, "odd")

	t.Run("no identity", func(t *testing.T) {
		report := Inspect(testConfig(), repo)
		if report.Issue == "" || report.OK() {
			t.Fatalf("expected an issue, got %+v", report)
		}
		if _, err := Fix(testConfig(), repo); err == nil {
			t.Fatal("fix should fail without an expected identity")
		}
	})

	cfg := testConfig()
	cfg.ActiveUser = "work"

	t.Run("no origin remote", func(t *testing.T) {
		report := Inspect(cfg, repo)
		if report.RemoteChecked || report.RemoteIssue == "" || report.OK() {
			t.Fatalf("expected a remote issue, got %+v", report)
		}
	})

	t.Run("non GitHub remote", func(t *testing.T) {
		testenv.Git(t, repo, "remote", "add", "origin", "https://gitlab.com/org/odd.git")
		report := Inspect(cfg, repo)
		if report.RemoteChecked || report.RemoteIssue == "" {
			t.Fatalf("expected an unsupported remote issue, got %+v", report)
		}
		// Fixing must leave an unsupported remote alone.
		if _, err := Fix(cfg, repo); err != nil {
			t.Fatalf("fix failed: %v", err)
		}
		if got := testenv.Git(t, repo, "remote", "get-url", "origin"); got != "https://gitlab.com/org/odd.git" {
			t.Fatalf("origin was changed to %q", got)
		}
	})
}
