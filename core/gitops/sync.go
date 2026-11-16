package gitops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/core/models"
)

type SyncResult = models.GitSyncResult

const defaultRemote = "origin"

// Fetch downloads new commits and branches from the remote without changing
// the working tree.
func Fetch(ctx context.Context, repo string) (*SyncResult, error) {
	if err := requireRemote(ctx, repo); err != nil {
		return nil, err
	}
	output, err := runCombined(ctx, runOptions{dir: repo, network: true}, "fetch", "--prune", defaultRemote)
	if err != nil {
		return nil, err
	}
	return &SyncResult{Output: output}, nil
}

// Pull brings the remote branch's commits into the current branch. When both
// sides changed it merges, unless the user configured Git to rebase on pull.
//
// A merge creates a commit, so the identity to record on it is passed in.
// Empty values fall back to whatever Git config is in effect.
func Pull(ctx context.Context, repo, authorName, authorEmail string) (*SyncResult, error) {
	if err := requireRemote(ctx, repo); err != nil {
		return nil, err
	}
	status, err := GetStatus(ctx, repo)
	if err != nil {
		return nil, err
	}
	if status.Detached {
		return nil, &Error{Kind: KindUnknown, Message: "No branch is checked out. Switch to a branch before pulling."}
	}
	if !status.HasUpstream {
		return nil, &Error{Kind: KindNoUpstream, Message: "This branch is not on the remote yet, so there is nothing to pull."}
	}

	args := []string{}
	if strings.TrimSpace(authorName) != "" && strings.TrimSpace(authorEmail) != "" {
		args = append(args, "-c", "user.name="+authorName, "-c", "user.email="+authorEmail)
	}
	args = append(args, "pull", "--no-edit")
	if !pullStrategyConfigured(ctx, repo) {
		// Git refuses to pull diverged branches until a strategy is chosen.
		// Merge is the default that never rewrites the user's commits.
		args = append(args, "--no-rebase")
	}
	output, err := runCombined(ctx, runOptions{dir: repo, network: true}, args...)
	if err != nil {
		return nil, err
	}
	return &SyncResult{Output: output}, nil
}

// Push uploads the current branch. A branch that is not on the remote yet is
// published and linked to it.
func Push(ctx context.Context, repo string) (*SyncResult, error) {
	if err := requireRemote(ctx, repo); err != nil {
		return nil, err
	}
	status, err := GetStatus(ctx, repo)
	if err != nil {
		return nil, err
	}
	if status.Detached || status.Branch == "" {
		return nil, &Error{Kind: KindUnknown, Message: "No branch is checked out. Switch to a branch before pushing."}
	}
	if status.Unborn {
		return nil, &Error{Kind: KindUnknown, Message: "There are no commits to push yet."}
	}

	result := &SyncResult{}
	args := []string{"push"}
	if !status.HasUpstream {
		args = append(args, "--set-upstream", defaultRemote, status.Branch)
		result.SetUpstream = true
	}
	output, err := runCombined(ctx, runOptions{dir: repo, network: true}, args...)
	if err != nil {
		return nil, err
	}
	result.Output = output
	return result, nil
}

// Clone clones a repository into destination, which must not exist yet or
// must be an empty directory.
func Clone(ctx context.Context, url, destination string) error {
	url = strings.TrimSpace(url)
	if url == "" || strings.HasPrefix(url, "-") {
		return fmt.Errorf("a repository address is required")
	}
	if strings.TrimSpace(destination) == "" {
		return fmt.Errorf("a destination folder is required")
	}

	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("could not create %s: %w", parent, err)
	}

	_, err := runCombined(ctx, runOptions{dir: parent, network: true, timeout: cloneTimeout},
		"clone", "--", url, destination)
	return err
}

// requireRemote makes sure the repository has an origin remote.
func requireRemote(ctx context.Context, repo string) error {
	if _, err := run(ctx, runOptions{dir: repo}, "remote", "get-url", defaultRemote); err != nil {
		return &Error{Kind: KindNotFound, Message: "This repository has no remote named origin, so it cannot be synced."}
	}
	return nil
}

func pullStrategyConfigured(ctx context.Context, repo string) bool {
	for _, key := range []string{"pull.rebase", "pull.ff"} {
		if output, err := run(ctx, runOptions{dir: repo}, "config", "--get", key); err == nil && strings.TrimSpace(output) != "" {
			return true
		}
	}
	return false
}
