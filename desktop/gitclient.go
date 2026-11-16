package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	corecheck "github.com/byterings/bgit/core/check"
	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/gitops"
	"github.com/byterings/bgit/core/models"
)

// repoSession is a managed repository together with the identity it must use.
type repoSession struct {
	cfg   *config.Config
	entry repoEntry
	user  *config.User
	alias string
}

// openRepoSession validates the repository path and resolves its identity.
func openRepoSession(path string) (*repoSession, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}
	entry, err := resolveKnownRepository(cfg, path)
	if err != nil {
		return nil, err
	}
	report := corecheck.Inspect(cfg, entry.path)
	return &repoSession{cfg: cfg, entry: entry, user: report.ExpectedUser, alias: report.ExpectedAlias}, nil
}

func (s *repoSession) state(ctx context.Context) (*RepoState, error) {
	status, err := gitops.GetStatus(ctx, s.entry.path)
	if err != nil {
		return nil, err
	}
	if status.Files == nil {
		status.Files = []models.GitFileChange{}
	}
	view := buildRepositoryView(ctx, s.cfg, s.entry)
	return &RepoState{Repository: &view, Status: status}, nil
}

// requireIdentity returns the identity for commits, or an explanation of why
// the repository has none.
func (s *repoSession) requireIdentity() (*config.User, error) {
	if s.user == nil {
		return nil, fmt.Errorf("no identity applies to this repository. Bind it to an identity first")
	}
	return s.user, nil
}

// gitFailure reports a failed Git operation as a result the UI can act on
// (for example offering "Pull" after a rejected push) instead of a bare error.
func (s *repoSession) gitFailure(ctx context.Context, err error) (*GitActionResult, error) {
	var gitErr *gitops.Error
	if !errors.As(err, &gitErr) {
		return nil, err
	}
	result := &GitActionResult{Kind: gitErr.Kind, Message: gitErr.Message, Detail: gitErr.Output}
	// The state is still useful after a failure: a conflicted pull leaves a
	// merge in progress that the UI must show.
	if state, stateErr := s.state(ctx); stateErr == nil {
		result.State = state
	}
	return result, nil
}

func (s *repoSession) gitSuccess(ctx context.Context, message, detail string) (*GitActionResult, error) {
	state, err := s.state(ctx)
	if err != nil {
		return nil, err
	}
	return &GitActionResult{OK: true, Message: message, Detail: detail, State: state}, nil
}

func getDesktopRepoState(ctx context.Context, path string) (*RepoState, error) {
	session, err := openRepoSession(path)
	if err != nil {
		return nil, err
	}
	return session.state(ctx)
}

func getDesktopFileDiff(ctx context.Context, request FileDiffRequest) (*models.GitDiff, error) {
	session, err := openRepoSession(request.Path)
	if err != nil {
		return nil, err
	}
	var diff *models.GitDiff
	if strings.TrimSpace(request.Commit) != "" {
		diff, err = gitops.DiffCommitFile(ctx, session.entry.path, request.Commit, request.File)
	} else {
		diff, err = gitops.DiffFile(ctx, session.entry.path, request.File)
	}
	if err != nil {
		return nil, err
	}
	if diff.Lines == nil {
		diff.Lines = []models.GitDiffLine{}
	}
	return diff, nil
}

func commitDesktopChanges(ctx context.Context, request CommitChangesRequest) (*GitActionResult, error) {
	session, err := openRepoSession(request.Path)
	if err != nil {
		return nil, err
	}
	user, err := session.requireIdentity()
	if err != nil {
		return nil, err
	}

	commit, err := gitops.Commit(ctx, session.entry.path, gitops.CommitRequest{
		Summary:     request.Summary,
		Description: request.Description,
		Paths:       request.Files,
		AuthorName:  user.Name,
		AuthorEmail: user.Email,
	})
	if err != nil {
		return session.gitFailure(ctx, err)
	}
	return session.gitSuccess(ctx, fmt.Sprintf("Committed %s as %s", commit.ShortOID, session.alias), "")
}

func discardDesktopChanges(ctx context.Context, request DiscardChangesRequest) (*GitActionResult, error) {
	session, err := openRepoSession(request.Path)
	if err != nil {
		return nil, err
	}
	if err := gitops.Discard(ctx, session.entry.path, request.Files); err != nil {
		return session.gitFailure(ctx, err)
	}
	return session.gitSuccess(ctx, "Discarded changes to "+countOf(len(request.Files), "file", "files"), "")
}

func listDesktopBranches(ctx context.Context, path string) ([]models.GitBranch, error) {
	session, err := openRepoSession(path)
	if err != nil {
		return nil, err
	}
	branches, err := gitops.ListBranches(ctx, session.entry.path)
	if err != nil {
		return nil, err
	}
	if branches == nil {
		branches = []models.GitBranch{}
	}
	return branches, nil
}

func switchDesktopBranch(ctx context.Context, request BranchRequest) (*GitActionResult, error) {
	session, err := openRepoSession(request.Path)
	if err != nil {
		return nil, err
	}
	if err := gitops.SwitchBranch(ctx, session.entry.path, request.Name); err != nil {
		return session.gitFailure(ctx, err)
	}
	return session.gitSuccess(ctx, fmt.Sprintf("Switched to %s", request.Name), "")
}

func createDesktopBranch(ctx context.Context, request BranchRequest) (*GitActionResult, error) {
	session, err := openRepoSession(request.Path)
	if err != nil {
		return nil, err
	}
	if err := gitops.CreateBranch(ctx, session.entry.path, request.Name); err != nil {
		return session.gitFailure(ctx, err)
	}
	return session.gitSuccess(ctx, fmt.Sprintf("Created and switched to %s", strings.TrimSpace(request.Name)), "")
}

func getDesktopHistory(ctx context.Context, request HistoryRequest) ([]models.GitCommit, error) {
	session, err := openRepoSession(request.Path)
	if err != nil {
		return nil, err
	}
	return gitops.Log(ctx, session.entry.path, request.Limit, request.Skip)
}

func getDesktopCommitFiles(ctx context.Context, request CommitFilesRequest) ([]models.GitFileChange, error) {
	session, err := openRepoSession(request.Path)
	if err != nil {
		return nil, err
	}
	return gitops.CommitFiles(ctx, session.entry.path, request.Commit)
}

func fetchDesktopRepository(ctx context.Context, path string) (*GitActionResult, error) {
	session, err := openRepoSession(path)
	if err != nil {
		return nil, err
	}
	loadKeyIntoAgent(session.user)
	result, err := gitops.Fetch(ctx, session.entry.path)
	if err != nil {
		return session.gitFailure(ctx, err)
	}
	return session.gitSuccess(ctx, "Fetched from the remote", result.Output)
}

func pullDesktopRepository(ctx context.Context, path string) (*GitActionResult, error) {
	session, err := openRepoSession(path)
	if err != nil {
		return nil, err
	}
	loadKeyIntoAgent(session.user)

	// A pull may create a merge commit; it must carry the repository identity.
	name, email := "", ""
	if session.user != nil {
		name, email = session.user.Name, session.user.Email
	}
	result, err := gitops.Pull(ctx, session.entry.path, name, email)
	if err != nil {
		return session.gitFailure(ctx, err)
	}
	return session.gitSuccess(ctx, "Pulled the latest commits", result.Output)
}

// pushDesktopRepository pushes the current branch. The repository is checked
// against its identity first, so a mismatch is explained here with a fix
// instead of surfacing as a blocked push from the hook.
func pushDesktopRepository(ctx context.Context, path string) (*GitActionResult, error) {
	session, err := openRepoSession(path)
	if err != nil {
		return nil, err
	}

	report := corecheck.Inspect(session.cfg, session.entry.path)
	fixable := report.Issue == "" && ((report.IdentityChecked && !report.IdentityMatches) || (report.RemoteChecked && !report.RemoteMatches))
	if fixable {
		view := RepositoryView{Problems: []string{}, Notes: []string{}}
		applyCheckReport(&view, report)
		state, stateErr := session.state(ctx)
		if stateErr != nil {
			return nil, stateErr
		}
		return &GitActionResult{
			Kind:    gitops.KindSafety,
			Message: fmt.Sprintf("'%s' does not match identity '%s'. Fix the repository, then push.", filepath.Base(session.entry.path), report.ExpectedAlias),
			Detail:  strings.Join(view.Problems, "\n"),
			State:   state,
		}, nil
	}

	loadKeyIntoAgent(session.user)
	result, err := gitops.Push(ctx, session.entry.path)
	if err != nil {
		return session.gitFailure(ctx, err)
	}
	message := "Pushed to the remote"
	if result.SetUpstream {
		message = "Published the branch to the remote"
	}
	return session.gitSuccess(ctx, message, result.Output)
}

func abortDesktopMerge(ctx context.Context, path string) (*GitActionResult, error) {
	session, err := openRepoSession(path)
	if err != nil {
		return nil, err
	}
	if err := gitops.AbortMerge(ctx, session.entry.path); err != nil {
		return session.gitFailure(ctx, err)
	}
	return session.gitSuccess(ctx, "Merge cancelled", "")
}
