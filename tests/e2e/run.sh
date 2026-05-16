#!/usr/bin/env bash
set -Eeuo pipefail

export HOME=/tmp/bgit-home
export GIT_CONFIG_GLOBAL="$HOME/.gitconfig"
export GIT_TERMINAL_PROMPT=0
BGIT_E2E_BIN_DIR="${BGIT_E2E_BIN_DIR:-/usr/local/bin}"
export PATH="$BGIT_E2E_BIN_DIR:/usr/local/bin:/usr/bin:/bin"

ROOT=/tmp/bgit-e2e
BARE="$ROOT/origin.git"
WORK="$ROOT/work"
WORKSPACES="$ROOT/workspaces"
LOG="$ROOT/e2e.log"

PASS=0
FAIL=0

section() {
  printf '\n== %s ==\n' "$*"
}

pass() {
  PASS=$((PASS + 1))
  printf 'ok - %s\n' "$*"
}

fail() {
  FAIL=$((FAIL + 1))
  printf 'not ok - %s\n' "$*"
  return 1
}

run_cmd() {
  local name="$1"
  shift
  printf '\n$ %s\n' "$*" | tee -a "$LOG"
  set +e
  "$@" >"$ROOT/out.txt" 2>&1
  local code=$?
  set -e
  cat "$ROOT/out.txt" | tee -a "$LOG"
  if [ "$code" -eq 0 ]; then
    pass "$name"
  else
    fail "$name exited $code"
  fi
}

run_cmd_stdin() {
  local input="$1"
  local name="$2"
  shift 2
  printf '\n$ %s\n' "$*" | tee -a "$LOG"
  set +e
  printf '%b' "$input" | "$@" >"$ROOT/out.txt" 2>&1
  local code=$?
  set -e
  cat "$ROOT/out.txt" | tee -a "$LOG"
  if [ "$code" -eq 0 ]; then
    pass "$name"
  else
    fail "$name exited $code"
  fi
}

assert_file_exists() {
  local path="$1"
  local name="$2"
  [ -e "$path" ] && pass "$name" || fail "$name"
}

assert_file_missing() {
  local path="$1"
  local name="$2"
  [ ! -e "$path" ] && pass "$name" || fail "$name"
}

assert_contains() {
  local path="$1"
  local needle="$2"
  local name="$3"
  grep -Fq "$needle" "$path" && pass "$name" || fail "$name"
}

assert_not_contains() {
  local path="$1"
  local needle="$2"
  local name="$3"
  if [ ! -e "$path" ] || ! grep -Fq "$needle" "$path"; then
    pass "$name"
  else
    fail "$name"
  fi
}

assert_git_config() {
  local key="$1"
  local expected="$2"
  local got
  got="$(git config --global --get "$key" || true)"
  [ "$got" = "$expected" ] && pass "git config $key is $expected" || fail "git config $key got '$got', expected '$expected'"
}

assert_output_contains() {
  local needle="$1"
  local name="$2"
  grep -Fq "$needle" "$ROOT/out.txt" && pass "$name" || fail "$name"
}

summary() {
  printf '\n== Summary ==\n'
  printf 'passed: %s\nfailed: %s\n' "$PASS" "$FAIL"
  if [ "$FAIL" -ne 0 ]; then
    printf '\nVerbose log: %s\n' "$LOG"
    exit 1
  fi
  printf 'E2E PASS\n'
}
trap summary EXIT

reset_env() {
  rm -rf "$HOME" "$ROOT"
  mkdir -p "$HOME/.ssh" "$ROOT"
  chmod 700 "$HOME/.ssh"
  : > "$LOG"
  git config --global init.defaultBranch main
  git config --global user.name "Original User"
  git config --global user.email "original@example.com"
}

make_key() {
  local name="$1"
  ssh-keygen -t ed25519 -N "" -C "$name@example.test" -f "$HOME/.ssh/$name" >/dev/null
  chmod 600 "$HOME/.ssh/$name"
}

create_local_repo() {
  mkdir -p "$ROOT"
  git init --bare "$BARE" >/dev/null
  git clone "$BARE" "$WORK" >/dev/null 2>&1
  (
    cd "$WORK"
    printf 'hello\n' > README.md
    git add README.md
    git commit -m "initial commit" >/dev/null
    git push origin main >/dev/null
  )
}

section "Prepare isolated HOME and local repos"
reset_env
make_key personal-test
make_key work-test
create_local_repo

section "Core commands"
run_cmd "bgit --version" bgit --version
assert_output_contains "bgit version" "version output is shown"

run_cmd "bgit setup" bgit setup
assert_file_exists "$HOME/.bgit/config.toml" "bgit config exists after setup"
assert_file_exists "$HOME/.bgit/hooks/pre-push" "managed pre-push hook exists"
assert_git_config "core.hooksPath" "$HOME/.bgit/hooks"
assert_contains "$HOME/.ssh/config" "BEGIN BGIT MANAGED" "ssh config has bgit-managed block"

run_cmd "bgit add personal" bgit add --alias personal --name "Personal Test" --email "personal-test@example.com" --github "personal-test" --ssh-key "$HOME/.ssh/personal-test"
assert_output_contains "User 'personal' added successfully" "personal identity added"

run_cmd "bgit add work" bgit add --alias work --name "Work Test" --email "work-test@example.com" --github "work-test" --ssh-key "$HOME/.ssh/work-test"
assert_output_contains "User 'work' added successfully" "work identity added"

run_cmd "bgit list" bgit list
assert_output_contains "personal" "list includes personal"
assert_output_contains "work" "list includes work"

run_cmd "bgit use personal" bgit use personal
assert_git_config "user.name" "Personal Test"
assert_git_config "user.email" "personal-test@example.com"
assert_contains "$HOME/.ssh/config" "Host github.com-personal-test" "ssh config contains personal host"
assert_contains "$HOME/.ssh/config" "Host github.com-work-test" "ssh config contains work host"

run_cmd "bgit active" bgit active
assert_output_contains "personal" "active shows personal"

run_cmd "bgit status" bgit status
assert_output_contains "personal" "status shows personal"

run_cmd "bgit doctor" bgit doctor
assert_output_contains "Checking bgit configuration" "doctor ran"

section "Workspace and binding"
mkdir -p "$WORKSPACES"
run_cmd "bgit workspace" bgit workspace --path "$WORKSPACES" --users personal,work
assert_file_exists "$WORKSPACES/personal" "personal workspace folder exists"
assert_file_exists "$WORKSPACES/work" "work workspace folder exists"
run_cmd "bgit workspace list" bgit workspace --list
assert_output_contains "$WORKSPACES/personal" "workspace list includes personal path"

(
  cd "$WORK"
  run_cmd "bgit bind" bgit bind --user personal
)
assert_contains "$HOME/.bgit/config.toml" "path = \"$WORK\"" "repo binding saved"

(
  cd "$WORK"
  run_cmd "bgit sync" bgit sync --fix
)
assert_output_contains "All checks passed" "sync reports clean state"

section "Remote conversion and check"
(
  cd "$WORK"
  git remote set-url origin "https://github.com/example/repo.git"
  run_cmd "bgit remote fix" bgit remote fix
  [ "$(git remote get-url origin)" = "git@github.com-personal-test:example/repo.git" ] && pass "remote fix converted to bgit SSH URL" || fail "remote fix did not convert URL"
  run_cmd "bgit check" bgit check
  assert_output_contains "Safety checks passed" "check passes after remote fix"
  run_cmd "bgit remote restore" bgit remote restore
  [ "$(git remote get-url origin)" = "git@github.com:example/repo.git" ] && pass "remote restore converted to standard SSH URL" || fail "remote restore did not convert URL"
)

section "Delete identity"
run_cmd "bgit delete work" bgit delete work --force
assert_output_contains "User 'work' deleted" "work identity deleted"
assert_not_contains "$HOME/.ssh/config" "Host github.com-work-test" "ssh config no longer contains deleted work host"

section "Uninstall cleanup"
(
  cd "$WORK"
  git remote set-url origin "git@github.com-personal-test:example/repo.git"
)
run_cmd "bgit uninstall dry-run" bgit uninstall --dry-run --force --verbose
assert_output_contains "dry run complete" "dry-run completes"

run_cmd "bgit uninstall" bgit uninstall --force --remove-config --verbose
assert_file_missing "$HOME/.bgit" "bgit config removed when requested"
assert_not_contains "$HOME/.ssh/config" "BEGIN BGIT MANAGED" "managed ssh block removed"
assert_git_config "core.hooksPath" ""
assert_git_config "user.name" "Original User"
assert_git_config "user.email" "original@example.com"

(
  cd "$WORK"
  git remote set-url origin "$BARE"
  printf 'after uninstall\n' >> README.md
  git add README.md
  git commit -m "normal git after uninstall" >/dev/null
  git push origin main >/dev/null
)
pass "normal git commit and push works after uninstall"

section "Regression issue #5: preserve non-bgit Git config"
reset_env
make_key personal-test
create_local_repo
run_cmd "issue5 setup" bgit setup
run_cmd "issue5 add personal" bgit add --alias personal --name "Personal Test" --email "personal-test@example.com" --github "personal-test" --ssh-key "$HOME/.ssh/personal-test"
run_cmd "issue5 use personal" bgit use personal
assert_git_config "user.name" "Personal Test"
assert_git_config "user.email" "personal-test@example.com"

git config --global user.name "Manual Desktop User"
git config --global user.email "desktop@example.com"
run_cmd "issue5 uninstall" bgit uninstall --force --remove-config --verbose
assert_file_missing "$HOME/.bgit" "issue5 bgit config removed"
assert_not_contains "$HOME/.ssh/config" "BEGIN BGIT MANAGED" "issue5 ssh config cleaned"
assert_git_config "user.name" "Manual Desktop User"
assert_git_config "user.email" "desktop@example.com"
assert_git_config "core.hooksPath" ""

(
  cd "$WORK"
  git remote set-url origin "$BARE"
  printf 'desktop normal\n' >> README.md
  git add README.md
  git commit -m "desktop normal git" >/dev/null
  git push origin main >/dev/null
)
pass "issue5 normal git works after uninstall"
