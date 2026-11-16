// Package testenv gives tests an isolated HOME, Git config, and helpers for
// creating disposable repositories. It is only imported by tests.
package testenv

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Isolate points HOME and Git's global/system config at a temp directory so a
// test can never read or change the developer's real bgit, Git, or SSH state.
func Isolate(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, ".gitconfig"))
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	// A hooks path or signing setup from the host must not leak into tests.
	t.Setenv("GIT_CONFIG_COUNT", "0")
	t.Setenv("SSH_AUTH_SOCK", "")
	return home
}

// Git runs a git command in dir and fails the test on error.
func Git(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

// InitRepo creates an empty repository on branch main and returns its path.
func InitRepo(t *testing.T, parent, name string) string {
	t.Helper()

	repo := filepath.Join(parent, name)
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}
	Git(t, repo, "init", "-q", "-b", "main")
	// Resolve symlinks (macOS temp dirs) so paths match what Git reports.
	if resolved, err := filepath.EvalSymlinks(repo); err == nil {
		repo = resolved
	}
	return repo
}

// WriteFile writes content to a file inside dir, creating parent directories.
func WriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
	return path
}

// FakeGitHub makes Git's SSH transport talk to local bare repositories instead
// of github.com, so clone, fetch, pull, and push through bgit host aliases can
// be tested offline with the real Git protocol.
//
// A repository created with AddRepo("work-gh", "org", "app") is reachable as
// git@github.com-work-gh:org/app.git. A host alias without repositories
// behaves like an account without access.
type FakeGitHub struct {
	t    *testing.T
	root string
}

// NewFakeGitHub installs the fake SSH command for the current test.
func NewFakeGitHub(t *testing.T, home string) *FakeGitHub {
	t.Helper()

	root := filepath.Join(home, "fake-github")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("failed to create fake GitHub root: %v", err)
	}

	// The script drops SSH options, then runs the requested Git service in
	// the folder that stands for the host alias.
	script := filepath.Join(root, "ssh")
	content := `#!/bin/sh
while [ $# -gt 0 ]; do
  case "$1" in
    -o) shift 2 ;;
    -*) shift ;;
    *) break ;;
  esac
done
host="${1#git@}"
shift
if [ ! -d "` + root + `/hosts/$host" ]; then
  echo "git@github.com: Permission denied (publickey)." >&2
  exit 255
fi
cd "` + root + `/hosts/$host" || exit 255
exec sh -c "$*"
`
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatalf("failed to write fake ssh: %v", err)
	}
	t.Setenv("GIT_SSH_COMMAND", script)
	return &FakeGitHub{t: t, root: root}
}

// AddRepo creates an empty bare repository reachable through the host alias
// of githubUsername and returns its path on disk.
func (f *FakeGitHub) AddRepo(githubUsername, owner, name string) string {
	f.t.Helper()

	bare := filepath.Join(f.root, "hosts", "github.com-"+githubUsername, owner, name+".git")
	if err := os.MkdirAll(filepath.Dir(bare), 0755); err != nil {
		f.t.Fatalf("failed to create fake repo dir: %v", err)
	}
	Git(f.t, f.root, "init", "-q", "--bare", "-b", "main", bare)
	return bare
}

// Seed pushes one commit containing README.md to a bare repository.
func (f *FakeGitHub) Seed(bare string) {
	f.t.Helper()

	work := filepath.Join(f.root, "seed-"+filepath.Base(bare))
	if err := os.MkdirAll(work, 0755); err != nil {
		f.t.Fatalf("failed to create seed dir: %v", err)
	}
	Git(f.t, work, "init", "-q", "-b", "main")
	WriteFile(f.t, work, "README.md", "line one\nline two\nline three\n")
	Git(f.t, work, "add", "-A")
	Git(f.t, work, "-c", "user.name=Seed User", "-c", "user.email=seed@example.com", "commit", "-q", "-m", "initial commit")
	// Seeding is not part of the scenario under test, so hooks are skipped.
	Git(f.t, work, "-c", "core.hooksPath="+os.DevNull, "push", "-q", bare, "main")
}

// StubBgit puts a stand-in `bgit` first on PATH that exits with exitCode and
// records its arguments in the returned marker file. Tests use it so the
// managed pre-push hook never reaches a real bgit installed on the machine.
func StubBgit(t *testing.T, home string, exitCode int) string {
	t.Helper()

	binDir := filepath.Join(home, "stub-bin")
	marker := filepath.Join(home, "bgit-stub-args")
	script := "#!/bin/sh\necho \"$*\" > '" + marker + "'\nexit " + strconv.Itoa(exitCode) + "\n"
	path := WriteFile(t, binDir, "bgit", script)
	if err := os.Chmod(path, 0755); err != nil {
		t.Fatalf("failed to make stub bgit executable: %v", err)
	}
	if !strings.HasPrefix(os.Getenv("PATH"), binDir+string(os.PathListSeparator)) {
		t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	return marker
}
