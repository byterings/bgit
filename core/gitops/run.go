// Package gitops wraps the Git command line for the everyday operations the
// desktop app offers: status, diff, commit, branches, history, and sync.
//
// Every command runs without a terminal: nothing may prompt, because a
// windowed app has nobody to answer. Failures come back as *Error with a
// category and a plain-language message.
package gitops

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/byterings/bgit/internal/platform"
)

const (
	localTimeout   = 60 * time.Second
	networkTimeout = 10 * time.Minute
	cloneTimeout   = 30 * time.Minute

	// emptyTree is the id of Git's empty tree, used to diff against "nothing".
	emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
)

// Error categories.
const (
	KindAuth       = "auth"
	KindNetwork    = "network"
	KindRejected   = "rejected"
	KindSafety     = "safety"
	KindConflict   = "conflict"
	KindDirty      = "dirty"
	KindNotFound   = "not_found"
	KindNoUpstream = "no_upstream"
	KindIdentity   = "identity"
	KindTimeout    = "timeout"
	KindUnknown    = "unknown"
)

// Error is a failed Git command.
type Error struct {
	Kind string
	// Message is a plain-language explanation for the user.
	Message string
	// Output is Git's raw output, for details views.
	Output string
}

func (e *Error) Error() string {
	return e.Message
}

// KindOf returns the category of a gitops error, or KindUnknown.
func KindOf(err error) string {
	var gitErr *Error
	if errors.As(err, &gitErr) {
		return gitErr.Kind
	}
	return KindUnknown
}

type runOptions struct {
	dir     string
	network bool
	timeout time.Duration
	stdin   string
	// okExitCodes lists non-zero exit codes that are not failures
	// (for example `git diff --no-index` returns 1 when files differ).
	okExitCodes []int
	// combine also returns standard error in the output on success.
	combine bool
}

// baseEnv makes Git non-interactive and its messages stable to classify.
func baseEnv() []string {
	return append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_OPTIONAL_LOCKS=0",
		"GIT_EDITOR=true",
		"GIT_MERGE_AUTOEDIT=no",
		"LC_ALL=C",
		"LANGUAGE=C",
	)
}

// sshCommand returns the SSH command for network operations. It keeps any SSH
// command the user configured and adds options that stop SSH from prompting:
// a passphrase or host-key question would hang forever without a terminal.
// New host keys are accepted on first contact; changed keys still fail.
func sshCommand(ctx context.Context, dir string) string {
	base := strings.TrimSpace(os.Getenv("GIT_SSH_COMMAND"))
	if base == "" {
		cmd := platform.PrepareCommand(exec.CommandContext(ctx, "git", "config", "--get", "core.sshCommand"))
		cmd.Dir = dir
		if output, err := cmd.Output(); err == nil {
			base = strings.TrimSpace(string(output))
		}
	}
	if base == "" {
		base = "ssh"
	}
	return base + " -o BatchMode=yes -o StrictHostKeyChecking=accept-new"
}

// runCombined executes git and returns standard output and standard error
// together, trimmed. Git writes its sync summaries to standard error.
func runCombined(ctx context.Context, opts runOptions, args ...string) (string, error) {
	opts.combine = true
	output, err := run(ctx, opts, args...)
	return strings.TrimSpace(output), err
}

// run executes git and returns its standard output.
func run(ctx context.Context, opts runOptions, args ...string) (string, error) {
	timeout := opts.timeout
	if timeout == 0 {
		timeout = localTimeout
		if opts.network {
			timeout = networkTimeout
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := platform.PrepareCommand(exec.CommandContext(ctx, "git", args...))
	cmd.Dir = opts.dir
	cmd.Env = baseEnv()
	if opts.network {
		cmd.Env = append(cmd.Env, "GIT_SSH_COMMAND="+sshCommand(ctx, opts.dir))
	}
	if opts.stdin != "" {
		cmd.Stdin = strings.NewReader(opts.stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		if opts.combine {
			return stdout.String() + stderr.String(), nil
		}
		return stdout.String(), nil
	}

	if ctx.Err() == context.DeadlineExceeded {
		return "", &Error{Kind: KindTimeout, Message: "Git took too long and was stopped. Check the network connection and try again.", Output: stderr.String()}
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		for _, code := range opts.okExitCodes {
			if exitErr.ExitCode() == code {
				return stdout.String(), nil
			}
		}
		return "", classify(stdout.String(), stderr.String())
	}

	if errors.Is(err, exec.ErrNotFound) {
		return "", &Error{Kind: KindNotFound, Message: "Git is not installed or could not be started."}
	}
	return "", &Error{Kind: KindUnknown, Message: fmt.Sprintf("Git could not be started: %v", err)}
}

// classify turns Git's output into a categorised, readable error.
func classify(stdout, stderr string) *Error {
	output := strings.TrimSpace(stderr)
	if output == "" {
		output = strings.TrimSpace(stdout)
	}
	lower := strings.ToLower(output + "\n" + stdout)

	gitErr := &Error{Kind: KindUnknown, Output: output}
	switch {
	case strings.Contains(lower, "bgit: push blocked by safety checks"):
		gitErr.Kind = KindSafety
		gitErr.Message = "bgit blocked this push because the repository does not match its identity. Fix the repository, then push again."
	case strings.Contains(lower, "permission denied (publickey)"),
		strings.Contains(lower, "authentication failed"),
		strings.Contains(lower, "could not read username"),
		strings.Contains(lower, "terminal prompts disabled"):
		gitErr.Kind = KindAuth
		gitErr.Message = "GitHub did not accept this identity. Check that its SSH key is added to the GitHub account and that the account can access this repository."
	case strings.Contains(lower, "repository not found"):
		gitErr.Kind = KindAuth
		gitErr.Message = "GitHub could not find this repository for the current identity. Check the address, and that this identity's GitHub account can access it."
	case strings.Contains(lower, "could not resolve host"),
		strings.Contains(lower, "could not resolve hostname"),
		strings.Contains(lower, "network is unreachable"),
		strings.Contains(lower, "connection timed out"),
		strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "no route to host"):
		gitErr.Kind = KindNetwork
		gitErr.Message = "Could not reach the remote. Check the internet connection and try again."
	case strings.Contains(lower, "host key verification failed"):
		gitErr.Kind = KindAuth
		gitErr.Message = "The saved host key for the remote does not match. Check the known_hosts file in your .ssh folder."
	case strings.Contains(lower, "[rejected]"),
		strings.Contains(lower, "non-fast-forward"),
		strings.Contains(lower, "fetch first"),
		strings.Contains(lower, "failed to push some refs"):
		gitErr.Kind = KindRejected
		gitErr.Message = "The remote has commits you do not have yet. Pull first, then push again."
	case strings.Contains(lower, "would be overwritten by"),
		strings.Contains(lower, "please commit your changes or stash them"),
		strings.Contains(lower, "you have unstaged changes"),
		strings.Contains(lower, "your index contains uncommitted changes"):
		gitErr.Kind = KindDirty
		gitErr.Message = "Your uncommitted changes would be overwritten. Commit or discard them first."
	case strings.Contains(lower, "conflict"),
		strings.Contains(lower, "automatic merge failed"),
		strings.Contains(lower, "unmerged files"),
		strings.Contains(lower, "you have not concluded your merge"),
		strings.Contains(lower, "fix conflicts"):
		gitErr.Kind = KindConflict
		gitErr.Message = "There are merge conflicts. Resolve the conflicted files in your editor, then commit."
	case strings.Contains(lower, "please tell me who you are"),
		strings.Contains(lower, "unable to auto-detect email address"),
		strings.Contains(lower, "empty ident name"):
		gitErr.Kind = KindIdentity
		gitErr.Message = "Git has no name and email for this repository. Bind the repository to an identity, or set an active identity."
	case strings.Contains(lower, "has no upstream branch"),
		strings.Contains(lower, "no tracking information"),
		strings.Contains(lower, "no upstream configured"):
		gitErr.Kind = KindNoUpstream
		gitErr.Message = "This branch is not on the remote yet. Push it first."
	case strings.Contains(lower, "need to specify how to reconcile divergent branches"):
		gitErr.Kind = KindConflict
		gitErr.Message = "Your branch and the remote branch have both changed. Pull again to merge them."
	case strings.Contains(lower, "not a git repository"):
		gitErr.Kind = KindNotFound
		gitErr.Message = "This folder is not a Git repository."
	case strings.Contains(lower, "already exists and is not an empty directory"):
		gitErr.Message = "The destination folder already exists and is not empty. Choose another folder."
	case strings.Contains(lower, "nothing to commit"):
		gitErr.Message = "There is nothing to commit."
	default:
		gitErr.Message = lastMeaningfulLine(output)
	}
	return gitErr
}

// lastMeaningfulLine picks the line of Git output that best explains a
// failure: the last "fatal:" or "error:" line, otherwise the last line.
func lastMeaningfulLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		for _, prefix := range []string{"fatal: ", "error: "} {
			if rest, ok := strings.CutPrefix(line, prefix); ok && rest != "" {
				return strings.ToUpper(rest[:1]) + rest[1:]
			}
		}
	}
	last := strings.TrimSpace(lines[len(lines)-1])
	if last == "" {
		return "Git reported an error."
	}
	return last
}
