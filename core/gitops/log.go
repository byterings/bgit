package gitops

import (
	"context"
	"strconv"
	"strings"

	"github.com/byterings/bgit/core/models"
)

type CommitInfo = models.GitCommit

const maxLogLimit = 200

// Log returns commits reachable from HEAD, newest first.
func Log(ctx context.Context, repo string, limit, skip int) ([]CommitInfo, error) {
	if !hasHead(ctx, repo) {
		return []CommitInfo{}, nil
	}
	if limit <= 0 || limit > maxLogLimit {
		limit = maxLogLimit
	}
	if skip < 0 {
		skip = 0
	}

	// Fields are separated by NUL and commits by the record separator, so
	// commit messages can contain any text.
	const format = "%H%x00%h%x00%an%x00%ae%x00%at%x00%s%x00%b%x1e"
	output, err := run(ctx, runOptions{dir: repo},
		"log", "-n", strconv.Itoa(limit), "--skip="+strconv.Itoa(skip), "--format="+format)
	if err != nil {
		return nil, err
	}

	unpushed := unpushedCommits(ctx, repo)
	commits := []CommitInfo{}
	for _, record := range strings.Split(output, "\x1e") {
		record = strings.TrimLeft(record, "\n")
		if record == "" {
			continue
		}
		fields := strings.Split(record, "\x00")
		if len(fields) != 7 {
			continue
		}
		commit := CommitInfo{
			OID:         fields[0],
			ShortOID:    fields[1],
			AuthorName:  fields[2],
			AuthorEmail: fields[3],
			Summary:     fields[5],
			Body:        strings.TrimSpace(fields[6]),
			Unpushed:    unpushed[fields[0]],
		}
		commit.Time, _ = strconv.ParseInt(fields[4], 10, 64)
		commits = append(commits, commit)
	}
	return commits, nil
}

// unpushedCommits returns the commits on HEAD that the upstream branch does
// not have. Without an upstream every commit counts as unpushed.
func unpushedCommits(ctx context.Context, repo string) map[string]bool {
	revisions := "@{upstream}..HEAD"
	if _, err := run(ctx, runOptions{dir: repo}, "rev-parse", "--verify", "--quiet", "@{upstream}"); err != nil {
		revisions = "HEAD"
	}

	output, err := run(ctx, runOptions{dir: repo}, "rev-list", "-n", "1000", revisions)
	if err != nil {
		return map[string]bool{}
	}
	unpushed := map[string]bool{}
	for _, oid := range strings.Fields(output) {
		unpushed[oid] = true
	}
	return unpushed
}
