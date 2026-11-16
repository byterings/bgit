// Package check owns the push safety decisions: which identity a repository is
// expected to use, and whether its Git identity and origin remote match. The
// CLI (`bgit check`) adds prompts and output on top; the desktop app uses the
// non-interactive report and fix.
package check

import (
	"fmt"
	"strings"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
	corerepo "github.com/byterings/bgit/core/repo"
	"github.com/byterings/bgit/internal/git"
)

type Expectation = models.CheckExpectation
type Report = models.RepoCheckReport
type FixResult = models.RepoFixResult

const (
	// SourceActive marks an expectation that fell back to the global active user.
	SourceActive = "active"

	originRemote = "origin"
)

// ResolveExpected returns the identity a repository is expected to use.
// The repository owner (binding, then workspace, then remote host alias) wins;
// without an owner the global active user is expected.
func ResolveExpected(cfg *config.Config, repoRoot, remoteURL string) *Expectation {
	expectation := &Expectation{}

	ownerResult := corerepo.ResolveRepoOwner(cfg, repoRoot, remoteURL)
	if ownerResult != nil && ownerResult.Owner != nil {
		expectation.OwnerAlias = ownerResult.Owner.Alias
		expectation.Alias = ownerResult.Owner.Alias
		expectation.Source = ownerResult.Owner.Source
		expectation.User = ownerResult.Owner.User
		return expectation
	}

	if cfg.ActiveUser == "" {
		return expectation
	}

	expectation.Alias = cfg.ActiveUser
	expectation.Source = SourceActive
	expectation.User = cfg.FindUserByAlias(cfg.ActiveUser)
	return expectation
}

// GitIdentityMatches reports whether the effective Git identity of the
// repository equals the expected user.
func GitIdentityMatches(repoRoot string, expectedUser *config.User) (bool, error) {
	gitName, gitEmail, err := git.GetUserForRepo(repoRoot)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(gitName) == expectedUser.Name &&
		strings.TrimSpace(gitEmail) == expectedUser.Email, nil
}

// VerifyGitIdentity returns an error describing the mismatch when the
// effective Git identity of the repository is not the expected user.
func VerifyGitIdentity(repoRoot string, expectedUser *config.User, expectedAlias string) error {
	gitName, gitEmail, err := git.GetUserForRepo(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to read git config: %w", err)
	}

	if strings.TrimSpace(gitName) != expectedUser.Name || strings.TrimSpace(gitEmail) != expectedUser.Email {
		return fmt.Errorf("git config mismatch for '%s'. Expected %s <%s>, got %s <%s>. Run: bgit sync --fix",
			expectedAlias,
			expectedUser.Name,
			expectedUser.Email,
			strings.TrimSpace(gitName),
			strings.TrimSpace(gitEmail),
		)
	}

	return nil
}

// ExpectedRemoteURL returns the origin URL the expected user must push through.
func ExpectedRemoteURL(remoteURL string, expectedUser *config.User) (string, error) {
	return corerepo.ConvertToBgitURL(remoteURL, expectedUser.GitHubUsername)
}

// Inspect builds a non-interactive safety report for a repository. It never
// changes anything.
func Inspect(cfg *config.Config, repoRoot string) *Report {
	report := &Report{RepoRoot: repoRoot}

	// A missing origin is reported below; owner resolution still works without it.
	remoteURL, _ := git.GetRemoteURL(repoRoot, originRemote)
	report.RemoteURL = remoteURL

	expectation := ResolveExpected(cfg, repoRoot, remoteURL)
	report.ExpectedAlias = expectation.Alias
	report.ExpectedSource = expectation.Source
	report.ExpectedUser = expectation.User

	if expectation.Alias == "" {
		report.Issue = "No identity applies to this repository. Bind it to an identity or set an active identity."
		return report
	}
	if expectation.User == nil {
		report.Issue = fmt.Sprintf("Identity '%s' was not found in the bgit configuration.", expectation.Alias)
		return report
	}

	gitName, gitEmail, err := git.GetUserForRepo(repoRoot)
	if err == nil {
		report.IdentityChecked = true
		report.GitName = strings.TrimSpace(gitName)
		report.GitEmail = strings.TrimSpace(gitEmail)
		report.IdentityMatches = report.GitName == expectation.User.Name && report.GitEmail == expectation.User.Email
	}

	if remoteURL == "" {
		report.RemoteIssue = "No origin remote is configured."
		return report
	}
	expectedURL, err := ExpectedRemoteURL(remoteURL, expectation.User)
	if err != nil {
		report.RemoteIssue = "The origin remote is not a GitHub URL that bgit can manage."
		return report
	}
	report.RemoteChecked = true
	report.ExpectedRemoteURL = expectedURL
	report.RemoteMatches = remoteURL == expectedURL

	return report
}

// Fix aligns a repository with its expected identity and returns a fresh
// report. The Git identity is written to the repository unless the expectation
// comes from the global active user, where the global identity is the one that
// must match.
func Fix(cfg *config.Config, repoRoot string) (*FixResult, error) {
	report := Inspect(cfg, repoRoot)
	if report.Issue != "" {
		return nil, fmt.Errorf("%s", report.Issue)
	}

	result := &FixResult{}
	user := report.ExpectedUser

	if !report.IdentityMatches {
		if report.ExpectedSource == SourceActive {
			if err := git.SetGlobalUser(user.Name, user.Email); err != nil {
				return nil, err
			}
			result.IdentityScope = "global"
		} else {
			if err := git.SetLocalUser(repoRoot, user.Name, user.Email); err != nil {
				return nil, err
			}
			result.IdentityScope = "repository"
		}
		result.IdentityFixed = true
	}

	if report.RemoteChecked && !report.RemoteMatches {
		if err := git.SetRemoteURL(repoRoot, originRemote, report.ExpectedRemoteURL); err != nil {
			return nil, err
		}
		result.RemoteFixed = true
	}

	result.Report = Inspect(cfg, repoRoot)
	return result, nil
}
