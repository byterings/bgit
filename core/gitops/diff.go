package gitops

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/byterings/bgit/core/models"
)

const (
	// maxDiffLines caps the lines returned for one file.
	maxDiffLines = 3000
	// maxUntrackedDiffBytes caps the size of a new file that is rendered.
	maxUntrackedDiffBytes = 1 << 20
)

type Diff = models.GitDiff
type DiffLine = models.GitDiffLine

// DiffFile returns the uncommitted change to one file: the working tree
// compared with the last commit, staged or not.
func DiffFile(ctx context.Context, repo, path string) (*Diff, error) {
	clean, err := repoPath(path)
	if err != nil {
		return nil, err
	}

	base := "HEAD"
	if !hasHead(ctx, repo) {
		base = emptyTree
	}

	output, err := run(ctx, runOptions{dir: repo},
		"diff", "--no-color", "--no-ext-diff", "-M", base, "--", clean)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(output) != "" {
		return parseDiff(clean, output), nil
	}

	// No diff against the last commit: the file is untracked (or unchanged).
	return diffUntracked(ctx, repo, clean)
}

// diffUntracked renders a file Git does not track yet as all-added lines.
func diffUntracked(ctx context.Context, repo, clean string) (*Diff, error) {
	full := filepath.Join(repo, filepath.FromSlash(clean))
	info, err := os.Lstat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return &Diff{Path: clean}, nil
		}
		return nil, err
	}
	if info.IsDir() || !info.Mode().IsRegular() {
		return &Diff{Path: clean, Lines: []DiffLine{{Kind: "note", Text: "This entry is not a regular file."}}}, nil
	}
	if info.Size() > maxUntrackedDiffBytes {
		return &Diff{Path: clean, Truncated: true, Lines: []DiffLine{{Kind: "note", Text: "This new file is too large to preview."}}}, nil
	}

	// --no-index exits with 1 when the files differ, which is the normal case.
	output, err := run(ctx, runOptions{dir: repo, okExitCodes: []int{1}},
		"diff", "--no-color", "--no-ext-diff", "--no-index", "--", os.DevNull, clean)
	if err != nil {
		return nil, err
	}
	return parseDiff(clean, output), nil
}

// DiffCommitFile returns the change a commit made to one file, compared with
// the commit's first parent.
func DiffCommitFile(ctx context.Context, repo, oid, path string) (*Diff, error) {
	clean, err := repoPath(path)
	if err != nil {
		return nil, err
	}
	if err := validateRevision(oid); err != nil {
		return nil, err
	}

	output, err := run(ctx, runOptions{dir: repo},
		"diff", "--no-color", "--no-ext-diff", "-M", commitParent(ctx, repo, oid), oid, "--", clean)
	if err != nil {
		return nil, err
	}
	return parseDiff(clean, output), nil
}

// CommitFiles lists the files a commit changed, compared with its first parent.
func CommitFiles(ctx context.Context, repo, oid string) ([]FileChange, error) {
	if err := validateRevision(oid); err != nil {
		return nil, err
	}

	output, err := run(ctx, runOptions{dir: repo},
		"diff", "--name-status", "-z", "-M", commitParent(ctx, repo, oid), oid)
	if err != nil {
		return nil, err
	}
	return parseNameStatus(output), nil
}

// commitParent returns the first parent of a commit, or the empty tree for a
// root commit.
func commitParent(ctx context.Context, repo, oid string) string {
	output, err := run(ctx, runOptions{dir: repo}, "rev-parse", "--verify", "--quiet", oid+"^")
	if err != nil {
		return emptyTree
	}
	return strings.TrimSpace(output)
}

// parseNameStatus reads `git diff --name-status -z` output:
// "M\0path\0" or "R100\0old\0new\0".
func parseNameStatus(output string) []FileChange {
	fields := strings.Split(output, "\x00")
	changes := []FileChange{}

	for i := 0; i < len(fields); i++ {
		code := fields[i]
		if code == "" || i+1 >= len(fields) {
			continue
		}
		change := FileChange{Staged: true}
		switch code[0] {
		case 'R', 'C':
			if i+2 >= len(fields) {
				return changes
			}
			change.OrigPath = fields[i+1]
			change.Path = fields[i+2]
			change.Kind = "renamed"
			if code[0] == 'C' {
				change.Kind = "copied"
			}
			i += 2
		default:
			change.Path = fields[i+1]
			switch code[0] {
			case 'A':
				change.Kind = "added"
			case 'D':
				change.Kind = "deleted"
			case 'T':
				change.Kind = "typechange"
			default:
				change.Kind = "modified"
			}
			i++
		}
		changes = append(changes, change)
	}
	return changes
}

// parseDiff turns unified diff text into numbered lines.
func parseDiff(path, output string) *Diff {
	diff := &Diff{Path: path}
	oldLine, newLine := 0, 0
	inHunk := false

	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		if strings.HasPrefix(line, "Binary files ") || strings.HasPrefix(line, "GIT binary patch") {
			diff.Binary = true
			diff.Lines = []DiffLine{{Kind: "note", Text: "Binary file. The change cannot be shown as text."}}
			return diff
		}

		if strings.HasPrefix(line, "@@") {
			inHunk = true
			oldLine, newLine = parseHunkHeader(line)
			diff.Lines = append(diff.Lines, DiffLine{Kind: "hunk", Text: line})
			continue
		}
		if !inHunk {
			// File header lines (diff --git, index, ---, +++) carry no content.
			continue
		}
		if len(diff.Lines) >= maxDiffLines {
			diff.Truncated = true
			break
		}

		switch {
		case strings.HasPrefix(line, "+"):
			diff.Additions++
			diff.Lines = append(diff.Lines, DiffLine{Kind: "add", Text: line[1:], NewLine: newLine})
			newLine++
		case strings.HasPrefix(line, "-"):
			diff.Deletions++
			diff.Lines = append(diff.Lines, DiffLine{Kind: "del", Text: line[1:], OldLine: oldLine})
			oldLine++
		case strings.HasPrefix(line, `\`):
			// "\ No newline at end of file"
			diff.Lines = append(diff.Lines, DiffLine{Kind: "note", Text: strings.TrimPrefix(line, `\ `)})
		default:
			diff.Lines = append(diff.Lines, DiffLine{Kind: "context", Text: strings.TrimPrefix(line, " "), OldLine: oldLine, NewLine: newLine})
			oldLine++
			newLine++
		}
	}
	return diff
}

// parseHunkHeader reads the starting line numbers from "@@ -a,b +c,d @@".
func parseHunkHeader(line string) (oldStart, newStart int) {
	for _, field := range strings.Fields(line) {
		if len(field) < 2 {
			continue
		}
		start, _, _ := strings.Cut(field[1:], ",")
		number, err := strconv.Atoi(start)
		if err != nil {
			continue
		}
		switch field[0] {
		case '-':
			oldStart = number
		case '+':
			newStart = number
		}
	}
	return oldStart, newStart
}
