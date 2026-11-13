package gitops

import (
	"context"
	"fmt"
	"strings"

	"github.com/byterings/bgit/core/models"
)

type Remote = models.GitRemote

// ListRemotes returns the repository's remotes with their fetch and push
// addresses, in the order Git reports them.
func ListRemotes(ctx context.Context, repo string) ([]Remote, error) {
	output, err := run(ctx, runOptions{dir: repo}, "remote", "-v")
	if err != nil {
		return nil, err
	}

	// Each remote is listed twice: "<name>\t<url> (fetch)" and "... (push)".
	remotes := []Remote{}
	index := map[string]int{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		name, rest, found := strings.Cut(line, "\t")
		if !found {
			continue
		}
		url, kind, found := cutLast(rest, " (")
		if !found {
			continue
		}
		position, known := index[name]
		if !known {
			position = len(remotes)
			index[name] = position
			remotes = append(remotes, Remote{Name: name})
		}
		switch strings.TrimSuffix(kind, ")") {
		case "fetch":
			remotes[position].FetchURL = url
		case "push":
			remotes[position].PushURL = url
		}
	}
	return remotes, nil
}

func cutLast(value, separator string) (before, after string, found bool) {
	position := strings.LastIndex(value, separator)
	if position < 0 {
		return value, "", false
	}
	return value[:position], value[position+len(separator):], true
}

// DeleteBranch deletes a local branch. Without force Git refuses to delete a
// branch whose commits exist nowhere else; that comes back as KindUnmerged so
// the caller can ask the user before forcing.
func DeleteBranch(ctx context.Context, repo, name string, force bool) error {
	name = strings.TrimSpace(name)
	if err := validateBranchName(ctx, repo, name); err != nil {
		return err
	}

	status, err := GetStatus(ctx, repo)
	if err != nil {
		return err
	}
	if !status.Detached && status.Branch == name {
		return &Error{Kind: KindUnknown, Message: fmt.Sprintf("'%s' is the current branch. Switch to another branch before deleting it.", name)}
	}

	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err = run(ctx, runOptions{dir: repo}, "branch", flag, "--", name)
	return err
}
