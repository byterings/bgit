package gitops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/core/models"
)

type CommitRequest = models.GitCommitRequest
type CommitResult = models.GitCommitResult

// Commit records changes with an explicit author identity.
//
// With no paths every change is committed. With paths only those files are
// committed; anything else, including changes staged outside the app, is left
// exactly as it was.
func Commit(ctx context.Context, repo string, request CommitRequest) (*CommitResult, error) {
	summary := strings.TrimSpace(request.Summary)
	if summary == "" {
		return nil, fmt.Errorf("a commit summary is required")
	}
	if strings.TrimSpace(request.AuthorName) == "" || strings.TrimSpace(request.AuthorEmail) == "" {
		return nil, fmt.Errorf("a commit identity is required")
	}
	paths, err := repoPaths(request.Paths)
	if err != nil {
		return nil, err
	}

	status, err := GetStatus(ctx, repo)
	if err != nil {
		return nil, err
	}
	if status.UnresolvedConflicts > 0 {
		return nil, &Error{Kind: KindConflict, Message: "Some files still contain conflict markers. Resolve them in your editor before committing."}
	}
	if len(status.Files) == 0 && !status.MergeInProgress {
		return nil, &Error{Kind: KindUnknown, Message: "There is nothing to commit."}
	}

	message := summary
	if description := strings.TrimSpace(request.Description); description != "" {
		message += "\n\n" + description
	}

	// The identity is passed on the command line so the commit is correct even
	// if the repository's Git config drifted from its bgit identity.
	identity := []string{
		"-c", "user.name=" + request.AuthorName,
		"-c", "user.email=" + request.AuthorEmail,
	}

	// A merge cannot be committed partially: Git requires the whole result.
	commitAll := len(paths) == 0 || status.MergeInProgress
	if commitAll {
		if _, err := run(ctx, runOptions{dir: repo}, "add", "-A"); err != nil {
			return nil, err
		}
		args := append(identity, "commit", "--quiet", "--file=-")
		if _, err := run(ctx, runOptions{dir: repo, stdin: message}, args...); err != nil {
			return nil, err
		}
	} else {
		// `commit --only` needs every path to be known to Git, so new files are
		// added first. Tracked files are taken from the working tree as they are.
		if untracked := untrackedAmong(status, paths); len(untracked) > 0 {
			args := append([]string{"add", "--"}, untracked...)
			if _, err := run(ctx, runOptions{dir: repo}, args...); err != nil {
				return nil, err
			}
		}
		args := append(identity, "commit", "--quiet", "--file=-", "--only", "--")
		args = append(args, paths...)
		if _, err := run(ctx, runOptions{dir: repo, stdin: message}, args...); err != nil {
			return nil, err
		}
	}

	oid, err := run(ctx, runOptions{dir: repo}, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	oid = strings.TrimSpace(oid)
	result := &CommitResult{OID: oid, Summary: summary}
	if len(oid) >= 7 {
		result.ShortOID = oid[:7]
	}
	return result, nil
}

func untrackedAmong(status *Status, paths []string) []string {
	untracked := map[string]bool{}
	for _, file := range status.Files {
		if file.Kind == "untracked" {
			untracked[file.Path] = true
		}
	}
	selected := []string{}
	for _, path := range paths {
		if untracked[path] {
			selected = append(selected, path)
		}
	}
	return selected
}

// Discard throws away the uncommitted changes to the given files. Changed
// files go back to the last commit; new files are deleted. This cannot be
// undone, so callers must confirm with the user first.
func Discard(ctx context.Context, repo string, paths []string) error {
	cleaned, err := repoPaths(paths)
	if err != nil {
		return err
	}
	if len(cleaned) == 0 {
		return fmt.Errorf("choose at least one file to discard")
	}

	status, err := GetStatus(ctx, repo)
	if err != nil {
		return err
	}
	changes := map[string]FileChange{}
	for _, file := range status.Files {
		changes[file.Path] = file
	}
	headExists := hasHead(ctx, repo)

	for _, path := range cleaned {
		change, changed := changes[path]
		if !changed {
			continue
		}

		// A renamed file is restored under its original name.
		if change.OrigPath != "" && headExists {
			if _, err := run(ctx, runOptions{dir: repo}, "reset", "--quiet", "HEAD", "--", path, change.OrigPath); err != nil {
				return err
			}
			if _, err := run(ctx, runOptions{dir: repo}, "checkout", "HEAD", "--", change.OrigPath); err != nil {
				return err
			}
			if err := removeWorktreeFile(repo, path); err != nil {
				return err
			}
			continue
		}

		inHead := headExists && existsInHead(ctx, repo, path)
		if inHead {
			if _, err := run(ctx, runOptions{dir: repo}, "checkout", "HEAD", "--", path); err != nil {
				return err
			}
			continue
		}

		// The file does not exist in the last commit: unstage it if needed,
		// then delete it.
		if change.Staged {
			if _, err := run(ctx, runOptions{dir: repo}, "rm", "--quiet", "--cached", "--force", "--", path); err != nil {
				return err
			}
		}
		if err := removeWorktreeFile(repo, path); err != nil {
			return err
		}
	}
	return nil
}

func existsInHead(ctx context.Context, repo, path string) bool {
	_, err := run(ctx, runOptions{dir: repo}, "cat-file", "-e", "HEAD:"+path)
	return err == nil
}

func removeWorktreeFile(repo, path string) error {
	full := filepath.Join(repo, filepath.FromSlash(path))
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not delete %s: %w", path, err)
	}
	return nil
}

// AbortMerge cancels a merge that stopped on conflicts and restores the state
// from before the merge.
func AbortMerge(ctx context.Context, repo string) error {
	_, err := run(ctx, runOptions{dir: repo}, "merge", "--abort")
	return err
}
