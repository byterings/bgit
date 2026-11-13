package setup

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/internal/testenv"
)

const classicHook = `#!/bin/sh
# ---- BEGIN BGIT MANAGED ----
bgit check --from-hook
status=$?
if [ $status -ne 0 ]; then
  echo "bgit: push blocked by safety checks"
fi
exit $status
# ---- END BGIT MANAGED ----
`

func TestHookScriptWithoutHelperIsClassicCLIHook(t *testing.T) {
	// The CLI hook content is a compatibility contract: it must not change.
	if got := HookScript(""); got != classicHook {
		t.Fatalf("classic hook changed:\n%s", got)
	}
}

func TestEnsureConfigInitializedCreatesConfigOnce(t *testing.T) {
	testenv.Isolate(t)

	cfg, created, err := EnsureConfigInitialized()
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}
	if !created || cfg == nil {
		t.Fatalf("expected a newly created config, got created=%v cfg=%v", created, cfg)
	}

	_, created, err = EnsureConfigInitialized()
	if err != nil {
		t.Fatalf("second init failed: %v", err)
	}
	if created {
		t.Fatal("second init must not report a new config")
	}
}

func TestApplyInstallsHookAndRemembersPreviousHooksPath(t *testing.T) {
	home := testenv.Isolate(t)
	previous := filepath.Join(home, "my-hooks")
	testenv.Git(t, home, "config", "--global", "core.hooksPath", previous)

	cfg, _, err := EnsureConfigInitialized()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	result, err := Apply(cfg, Options{})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if result.ReplacedHooksPath != previous {
		t.Fatalf("ReplacedHooksPath = %q, want %q", result.ReplacedHooksPath, previous)
	}

	content, err := os.ReadFile(result.HookPath)
	if err != nil {
		t.Fatalf("hook not written: %v", err)
	}
	if string(content) != classicHook {
		t.Fatalf("unexpected hook content:\n%s", content)
	}

	hooksPath, err := GlobalHooksPath()
	if err != nil {
		t.Fatalf("GlobalHooksPath failed: %v", err)
	}
	if hooksPath != result.HooksDir {
		t.Fatalf("core.hooksPath = %q, want %q", hooksPath, result.HooksDir)
	}

	saved, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if !saved.SetupCompleted || !saved.PreviousHooksPathSet || saved.PreviousHooksPath != previous {
		t.Fatalf("setup metadata not saved: %+v", saved)
	}

	// A second run must not overwrite the remembered pre-bgit hooks path.
	second, err := Apply(saved, Options{})
	if err != nil {
		t.Fatalf("second apply failed: %v", err)
	}
	if second.ReplacedHooksPath != "" {
		t.Fatalf("second apply reported a replaced path: %q", second.ReplacedHooksPath)
	}
	if saved.PreviousHooksPath != previous {
		t.Fatalf("previous hooks path was overwritten: %q", saved.PreviousHooksPath)
	}

	if err := RestoreGlobalHooksPath(saved); err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if hooksPath, _ = GlobalHooksPath(); hooksPath != previous {
		t.Fatalf("restored core.hooksPath = %q, want %q", hooksPath, previous)
	}
}

func TestRestoreUnsetsHooksPathWhenNoneExistedBefore(t *testing.T) {
	testenv.Isolate(t)

	cfg, _, err := EnsureConfigInitialized()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := Apply(cfg, Options{}); err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if err := RestoreGlobalHooksPath(cfg); err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if hooksPath, _ := GlobalHooksPath(); hooksPath != "" {
		t.Fatalf("core.hooksPath should be unset, got %q", hooksPath)
	}
}

func TestApplyWithHelperWritesFallbackHook(t *testing.T) {
	home := testenv.Isolate(t)
	helper := filepath.Join(home, "apps", "it's bgit", "bgit-desktop")

	cfg, _, err := EnsureConfigInitialized()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := Apply(cfg, Options{HookHelperPath: helper}); err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	got, err := InstalledHookHelperPath()
	if err != nil {
		t.Fatalf("InstalledHookHelperPath failed: %v", err)
	}
	if got != filepath.ToSlash(helper) {
		t.Fatalf("helper path = %q, want %q", got, helper)
	}
}

// runHook executes the generated hook with a controlled PATH and reports its
// exit code and output.
func runHook(t *testing.T, script, pathEnv string) (int, string) {
	t.Helper()

	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not available")
	}
	hookPath := filepath.Join(t.TempDir(), "pre-push")
	if err := os.WriteFile(hookPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write hook: %v", err)
	}

	cmd := exec.Command(sh, hookPath)
	cmd.Env = []string{"PATH=" + pathEnv}
	output, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(output)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("hook did not run: %v", err)
	}
	return exitErr.ExitCode(), string(output)
}

func writeFakeBgit(t *testing.T, dir, name string, exitCode int) string {
	t.Helper()

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	path := filepath.Join(dir, name)
	script := "#!/bin/sh\necho \"" + name + " $*\"\nexit " + string(rune('0'+exitCode)) + "\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake binary: %v", err)
	}
	return path
}

func TestFallbackHookBehavior(t *testing.T) {
	base := t.TempDir()
	// Minimal PATH: only what the hook itself needs, never the host's bgit.
	emptyPath := filepath.Join(base, "empty")
	if err := os.MkdirAll(emptyPath, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("prefers bgit on PATH", func(t *testing.T) {
		pathDir := filepath.Join(base, "onpath")
		writeFakeBgit(t, pathDir, "bgit", 0)
		helper := writeFakeBgit(t, filepath.Join(base, "h1"), "helper", 1)

		code, output := runHook(t, HookScript(helper), pathDir)
		if code != 0 || !strings.Contains(output, "bgit check --from-hook") {
			t.Fatalf("code=%d output=%q", code, output)
		}
	})

	t.Run("falls back to helper and blocks on failure", func(t *testing.T) {
		helper := writeFakeBgit(t, filepath.Join(base, "dir with space"), "helper", 1)

		code, output := runHook(t, HookScript(helper), emptyPath)
		if code != 1 {
			t.Fatalf("expected exit 1, got %d (%q)", code, output)
		}
		if !strings.Contains(output, "helper check --from-hook") || !strings.Contains(output, "push blocked by safety checks") {
			t.Fatalf("unexpected output: %q", output)
		}
	})

	t.Run("helper success lets the push through", func(t *testing.T) {
		helper := writeFakeBgit(t, filepath.Join(base, "h3"), "helper", 0)

		code, output := runHook(t, HookScript(helper), emptyPath)
		if code != 0 || strings.Contains(output, "push blocked") {
			t.Fatalf("code=%d output=%q", code, output)
		}
	})

	t.Run("missing bgit and helper does not block pushes", func(t *testing.T) {
		code, output := runHook(t, HookScript(filepath.Join(base, "gone", "helper")), emptyPath)
		if code != 0 {
			t.Fatalf("expected exit 0, got %d (%q)", code, output)
		}
		if !strings.Contains(output, "safety check skipped") {
			t.Fatalf("expected a skip notice, got %q", output)
		}
	})
}

func TestEnsureHookRepairsMissingAndStaleHooks(t *testing.T) {
	home := testenv.Isolate(t)
	// Keep PATH free of any real bgit so the classic hook counts as broken.
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	if err := os.Symlink(gitPath, filepath.Join(binDir, "git")); err != nil {
		t.Skipf("cannot symlink git: %v", err)
	}
	t.Setenv("PATH", binDir)

	helper := writeFakeBgit(t, filepath.Join(home, "app"), "bgit-desktop", 0)

	cfg, _, err := EnsureConfigInitialized()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Not set up yet: nothing to repair.
	if changed, err := EnsureHook(cfg, helper); err != nil || changed {
		t.Fatalf("EnsureHook before setup: changed=%v err=%v", changed, err)
	}

	// A config imported from another machine says setup is complete, but Git
	// has no hooks path here. That is not repaired silently: turning push
	// protection on is an explicit action.
	cfg.SetupCompleted = true
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if changed, err := EnsureHook(cfg, helper); err != nil || changed {
		t.Fatalf("EnsureHook with no hooks path: changed=%v err=%v", changed, err)
	}
	if got, _ := GlobalHooksPath(); got != "" {
		t.Fatalf("hooks path must stay unset, got %q", got)
	}

	// Once protection is on, a deleted hook file is put back.
	result, err := Apply(cfg, Options{HookHelperPath: helper})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(result.HookPath); err != nil {
		t.Fatal(err)
	}
	changed, err := EnsureHook(cfg, helper)
	if err != nil || !changed {
		t.Fatalf("EnsureHook with a missing hook file: changed=%v err=%v", changed, err)
	}
	state, err := InspectHook()
	if err != nil {
		t.Fatal(err)
	}
	if !state.Active() || state.HelperPath != filepath.ToSlash(helper) {
		t.Fatalf("hook should be active with the helper: %+v", state)
	}

	// Up to date: no change.
	if changed, err := EnsureHook(cfg, helper); err != nil || changed {
		t.Fatalf("EnsureHook when current: changed=%v err=%v", changed, err)
	}

	// The app moved: the baked helper is outdated and must be rewritten.
	moved := writeFakeBgit(t, filepath.Join(home, "app2"), "bgit-desktop", 0)
	if changed, err := EnsureHook(cfg, moved); err != nil || !changed {
		t.Fatalf("EnsureHook after move: changed=%v err=%v", changed, err)
	}
	if got, _ := InstalledHookHelperPath(); got != filepath.ToSlash(moved) {
		t.Fatalf("helper = %q, want %q", got, moved)
	}

	// A hooks path the user chose on purpose is never overwritten.
	custom := filepath.Join(home, "custom-hooks")
	testenv.Git(t, home, "config", "--global", "core.hooksPath", custom)
	if changed, err := EnsureHook(cfg, helper); err != nil || changed {
		t.Fatalf("EnsureHook with custom hooks path: changed=%v err=%v", changed, err)
	}
	if got, _ := GlobalHooksPath(); got != custom {
		t.Fatalf("custom hooks path was replaced with %q", got)
	}
	if state, _ := InspectHook(); state.Active() {
		t.Fatalf("hook must not be reported active with a custom hooks path: %+v", state)
	}
}

func TestApplyCanRememberALaterCustomHooksPath(t *testing.T) {
	home := testenv.Isolate(t)

	cfg, _, err := EnsureConfigInitialized()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(cfg, Options{}); err != nil {
		t.Fatal(err)
	}

	// After setup the user points Git at their own hooks folder.
	custom := filepath.Join(home, "team-hooks")
	testenv.Git(t, home, "config", "--global", "core.hooksPath", custom)

	// The CLI default keeps the first record (no hooks path before bgit).
	if _, err := Apply(cfg, Options{}); err != nil {
		t.Fatal(err)
	}
	if cfg.PreviousHooksPath != "" {
		t.Fatalf("default setup must keep the first record, got %q", cfg.PreviousHooksPath)
	}

	testenv.Git(t, home, "config", "--global", "core.hooksPath", custom)
	result, err := Apply(cfg, Options{RememberReplacedHooksPath: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.ReplacedHooksPath != custom || cfg.PreviousHooksPath != custom {
		t.Fatalf("replaced=%q remembered=%q, want %q", result.ReplacedHooksPath, cfg.PreviousHooksPath, custom)
	}
	if err := RestoreGlobalHooksPath(cfg); err != nil {
		t.Fatal(err)
	}
	if got, _ := GlobalHooksPath(); got != custom {
		t.Fatalf("restored hooks path = %q, want %q", got, custom)
	}
}

func TestEnsureHookLeavesAWorkingCLIHookAlone(t *testing.T) {
	home := testenv.Isolate(t)
	// A machine set up with the CLI: classic hook, and bgit is on PATH.
	testenv.StubBgit(t, home, 0)
	helper := writeFakeBgit(t, filepath.Join(home, "app"), "bgit-desktop", 0)

	cfg, _, err := EnsureConfigInitialized()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Apply(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}

	changed, err := EnsureHook(cfg, helper)
	if err != nil || changed {
		t.Fatalf("EnsureHook on a CLI machine: changed=%v err=%v", changed, err)
	}
	content, err := os.ReadFile(result.HookPath)
	if err != nil || string(content) != classicHook {
		t.Fatalf("the CLI hook must stay byte-for-byte the same, got:\n%s", content)
	}
}
