package gitops

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/byterings/bgit/core/models"
)

type Branch = models.GitBranch

// validateRevision accepts only commit ids, so a value from the UI can never
// be read by Git as an option or an arbitrary expression.
func validateRevision(oid string) error {
	if len(oid) < 7 || len(oid) > 64 {
		return fmt.Errorf("invalid commit id")
	}
	for _, char := range oid {
		isHex := (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')
		if !isHex {
			return fmt.Errorf("invalid commit id")
		}
	}
	return nil
}

// validateBranchName checks a branch name with Git's own rules.
func validateBranchName(ctx context.Context, repo, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("a branch name is required")
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("a branch name cannot start with '-'")
	}
	if _, err := run(ctx, runOptions{dir: repo}, "check-ref-format", "--branch", name); err != nil {
		return fmt.Errorf("'%s' is not a valid branch name", name)
	}
	return nil
}

// ListBranches returns local branches followed by remote branches, each
// sorted with the most recently committed first.
func ListBranches(ctx context.Context, repo string) ([]Branch, error) {
	const format = "%(refname)%00%(refname:short)%00%(objectname:short)%00%(upstream:short)%00%(HEAD)%00%(committerdate:unix)"
	output, err := run(ctx, runOptions{dir: repo},
		"for-each-ref", "--sort=-committerdate", "--format="+format, "refs/heads", "refs/remotes")
	if err != nil {
		return nil, err
	}

	local, remote := []Branch{}, []Branch{}
	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		fields := strings.Split(line, "\x00")
		if len(fields) != 6 {
			continue
		}
		branch := Branch{
			Name:     fields[1],
			OID:      fields[2],
			Upstream: fields[3],
			Current:  fields[4] == "*",
		}
		branch.CommitTime, _ = strconv.ParseInt(fields[5], 10, 64)

		if strings.HasPrefix(fields[0], "refs/remotes/") {
			// "origin/HEAD" is a pointer to the default branch, not a branch.
			if strings.HasSuffix(fields[0], "/HEAD") {
				continue
			}
			branch.Remote = true
			remote = append(remote, branch)
			continue
		}
		local = append(local, branch)
	}
	return append(local, remote...), nil
}

// SwitchBranch checks out a local branch, or creates a local tracking branch
// when given a remote branch such as "origin/feature".
func SwitchBranch(ctx context.Context, repo, name string) error {
	if err := validateBranchName(ctx, repo, name); err != nil {
		return err
	}

	branches, err := ListBranches(ctx, repo)
	if err != nil {
		return err
	}
	for _, branch := range branches {
		if branch.Name != name {
			continue
		}
		if !branch.Remote {
			_, err := run(ctx, runOptions{dir: repo}, "checkout", name, "--")
			return err
		}

		// Remote branch: reuse the local branch of the same name if it exists.
		_, localName, _ := strings.Cut(name, "/")
		for _, candidate := range branches {
			if !candidate.Remote && candidate.Name == localName {
				_, err := run(ctx, runOptions{dir: repo}, "checkout", localName, "--")
				return err
			}
		}
		_, err := run(ctx, runOptions{dir: repo}, "checkout", "--track", name, "--")
		return err
	}
	return &Error{Kind: KindNotFound, Message: fmt.Sprintf("Branch '%s' does not exist.", name)}
}

// CreateBranch creates a branch from the current commit and switches to it.
// Uncommitted changes are carried over to the new branch.
func CreateBranch(ctx context.Context, repo, name string) error {
	name = strings.TrimSpace(name)
	if err := validateBranchName(ctx, repo, name); err != nil {
		return err
	}

	branches, err := ListBranches(ctx, repo)
	if err != nil {
		return err
	}
	for _, branch := range branches {
		if !branch.Remote && branch.Name == name {
			return fmt.Errorf("a branch named '%s' already exists", name)
		}
	}

	_, err = run(ctx, runOptions{dir: repo}, "checkout", "-b", name)
	return err
}
