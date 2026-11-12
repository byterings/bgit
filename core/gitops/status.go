package gitops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/byterings/bgit/core/models"
)

// maxStatusFiles caps the number of changed files returned, so a repository
// with a huge untracked folder cannot freeze the app.
const maxStatusFiles = 2000

type Status = models.GitStatus
type FileChange = models.GitFileChange

// GetStatus returns the branch and working tree state of a repository.
func GetStatus(ctx context.Context, repo string) (*Status, error) {
	output, err := run(ctx, runOptions{dir: repo},
		"status", "--porcelain=v2", "--branch", "-z", "--untracked-files=all")
	if err != nil {
		return nil, err
	}

	status := parseStatus(output)

	gitDir, err := gitDir(ctx, repo)
	if err == nil {
		status.MergeInProgress = fileExists(filepath.Join(gitDir, "MERGE_HEAD"))
		status.RebaseInProgress = dirExists(filepath.Join(gitDir, "rebase-merge")) || dirExists(filepath.Join(gitDir, "rebase-apply"))
	}
	return status, nil
}

// parseStatus reads `git status --porcelain=v2 --branch -z` output.
// Records are NUL-terminated; a rename record is followed by one extra record
// holding the original path.
func parseStatus(output string) *Status {
	status := &Status{}
	records := strings.Split(output, "\x00")

	for i := 0; i < len(records); i++ {
		record := records[i]
		if record == "" {
			continue
		}

		switch record[0] {
		case '#':
			parseStatusHeader(status, record)
			continue
		case '1':
			// 1 XY sub mH mI mW hH hI path
			fields := strings.SplitN(record, " ", 9)
			if len(fields) == 9 {
				addStatusFile(status, changeFromXY(fields[1], fields[8], ""))
			}
		case '2':
			// 2 XY sub mH mI mW hH hI Xscore path, then origPath as the next record
			fields := strings.SplitN(record, " ", 10)
			origPath := ""
			if i+1 < len(records) {
				origPath = records[i+1]
				i++
			}
			if len(fields) == 10 {
				addStatusFile(status, changeFromXY(fields[1], fields[9], origPath))
			}
		case 'u':
			// u XY sub m1 m2 m3 mW h1 h2 h3 path
			fields := strings.SplitN(record, " ", 11)
			if len(fields) == 11 {
				status.Conflicts++
				addStatusFile(status, FileChange{Path: fields[10], Kind: "conflicted", Staged: true, Unstaged: true})
			}
		case '?':
			addStatusFile(status, FileChange{Path: strings.TrimPrefix(record, "? "), Kind: "untracked", Unstaged: true})
		}
	}

	return status
}

func parseStatusHeader(status *Status, record string) {
	fields := strings.Fields(record)
	if len(fields) < 3 {
		return
	}
	switch fields[1] {
	case "branch.oid":
		if fields[2] == "(initial)" {
			status.Unborn = true
		} else if len(fields[2]) >= 7 {
			status.HeadOID = fields[2][:7]
		}
	case "branch.head":
		if fields[2] == "(detached)" {
			status.Detached = true
		} else {
			status.Branch = fields[2]
		}
	case "branch.upstream":
		status.Upstream = fields[2]
		status.HasUpstream = true
	case "branch.ab":
		if len(fields) >= 4 {
			status.Ahead, _ = strconv.Atoi(strings.TrimPrefix(fields[2], "+"))
			status.Behind, _ = strconv.Atoi(strings.TrimPrefix(fields[3], "-"))
		}
	}
}

func addStatusFile(status *Status, change FileChange) {
	if len(status.Files) >= maxStatusFiles {
		status.Truncated = true
		return
	}
	status.Files = append(status.Files, change)
}

// changeFromXY maps the two-letter index/worktree state to one change.
func changeFromXY(xy, path, origPath string) FileChange {
	change := FileChange{Path: path, OrigPath: origPath}
	if len(xy) != 2 {
		change.Kind = "modified"
		return change
	}
	index, worktree := xy[0], xy[1]
	change.Staged = index != '.'
	change.Unstaged = worktree != '.'

	switch {
	case index == 'R' || worktree == 'R':
		change.Kind = "renamed"
	case index == 'C' || worktree == 'C':
		change.Kind = "copied"
	case index == 'D' || worktree == 'D':
		change.Kind = "deleted"
	case index == 'A':
		change.Kind = "added"
	case index == 'T' || worktree == 'T':
		change.Kind = "typechange"
	default:
		change.Kind = "modified"
	}
	return change
}

// gitDir returns the absolute path of the repository's .git directory.
func gitDir(ctx context.Context, repo string) (string, error) {
	output, err := run(ctx, runOptions{dir: repo}, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// hasHead reports whether the repository has at least one commit.
func hasHead(ctx context.Context, repo string) bool {
	_, err := run(ctx, runOptions{dir: repo}, "rev-parse", "--verify", "--quiet", "HEAD")
	return err == nil
}

// repoPath validates a repository-relative path coming from the UI and
// returns it in the form Git expects. Paths may never escape the repository.
func repoPath(path string) (string, error) {
	cleaned := filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	if cleaned == "" || cleaned == "." {
		return "", fmt.Errorf("a file path is required")
	}
	if filepath.IsAbs(path) || strings.HasPrefix(cleaned, "/") || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("path is outside the repository: %s", path)
	}
	return cleaned, nil
}

func repoPaths(paths []string) ([]string, error) {
	cleaned := make([]string, 0, len(paths))
	for _, path := range paths {
		clean, err := repoPath(path)
		if err != nil {
			return nil, err
		}
		cleaned = append(cleaned, clean)
	}
	return cleaned, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
