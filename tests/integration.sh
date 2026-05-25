#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMPROOT="$(mktemp -d)"
TEST_HOME="$TMPROOT/home"
BIN="$TMPROOT/bgit"

cleanup() {
  rm -rf "$TMPROOT"
}
trap cleanup EXIT

log() {
  printf '\n==> %s\n' "$1"
}

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  if [[ "$haystack" != *"$needle"* ]]; then
    printf 'Expected output to contain: %s\nActual output:\n%s\n' "$needle" "$haystack" >&2
    exit 1
  fi
}

assert_not_exists() {
  local path="$1"
  if [[ -e "$path" ]]; then
    fail "expected path to be absent: $path"
  fi
}

assert_equals() {
  local expected="$1"
  local actual="$2"
  if [[ "$expected" != "$actual" ]]; then
    printf 'Expected: %s\nActual:   %s\n' "$expected" "$actual" >&2
    exit 1
  fi
}

run_bgit() {
  HOME="$TEST_HOME" "$BIN" "$@"
}

git_global() {
  HOME="$TEST_HOME" git config --global "$@"
}

log "Build test binary"
cd "$ROOT_DIR"
GOCACHE="${GOCACHE:-/tmp/bgit-gocache}" go build -o "$BIN" .

mkdir -p "$TEST_HOME"

log "Read-only commands do not initialize bgit"
output="$(run_bgit list)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit status)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit active)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit prompt --plain)"
assert_equals "none" "$output"

output="$(run_bgit doctor)"
assert_contains "$output" "Config file not found"
assert_not_exists "$TEST_HOME/.bgit"

log "Fresh setup preserves previous hooks path"
mkdir -p "$TMPROOT/original-hooks"
git_global user.name "Original User"
git_global user.email "original@example.com"
git_global core.hooksPath "$TMPROOT/original-hooks"

output="$(run_bgit setup)"
assert_contains "$output" "Setup complete"
assert_equals "$TEST_HOME/.bgit/hooks" "$(git_global --get core.hooksPath)"

log "Add and use first identity"
output="$(run_bgit add --alias company --name "Byterings Admin" --email "byterings@gmail.com" --github "byterings" --ssh-key skip)"
assert_contains "$output" "User 'company' added successfully"

output="$(run_bgit list)"
assert_contains "$output" "company"

output="$(run_bgit use company)"
assert_contains "$output" "Switched to identity: company"
assert_equals "Byterings Admin" "$(git_global --get user.name)"
assert_equals "byterings@gmail.com" "$(git_global --get user.email)"

output="$(run_bgit active)"
assert_contains "$output" "Active user: company"

output="$(run_bgit prompt --plain)"
assert_equals "company" "$output"

log "Workspace and repo binding flows"
WORKSPACES="$TMPROOT/workspaces"
mkdir -p "$WORKSPACES"
output="$(run_bgit workspace --path "$WORKSPACES" --users company)"
assert_contains "$output" "Workspace ready"

REPO="$TMPROOT/repo"
mkdir -p "$REPO"
HOME="$TEST_HOME" git -C "$REPO" init
HOME="$TEST_HOME" git -C "$REPO" remote add origin git@github.com:byterings/bgit.git

cd "$REPO"
output="$(run_bgit bind --user company)"
assert_contains "$output" "Bound repository to 'company'"

log "Remote fix and safety check"
output="$(run_bgit remote fix)"
assert_contains "$output" "Remote fixed for user 'company'"
assert_equals "git@github.com-byterings:byterings/bgit.git" "$(HOME="$TEST_HOME" git -C "$REPO" remote get-url origin)"

output="$(run_bgit check)"
assert_contains "$output" "Safety checks passed"

log "Uninstall restores Git state and removes managed config"
output="$(run_bgit uninstall --force)"
assert_contains "$output" "bgit uninstall complete"
assert_equals "$TMPROOT/original-hooks" "$(git_global --get core.hooksPath)"
assert_equals "Original User" "$(git_global --get user.name)"
assert_equals "original@example.com" "$(git_global --get user.email)"
assert_equals "git@github.com:byterings/bgit.git" "$(HOME="$TEST_HOME" git -C "$REPO" remote get-url origin)"
assert_not_exists "$TEST_HOME/.bgit"

if [[ -f "$TEST_HOME/.ssh/config" ]] && grep -q "BEGIN BGIT MANAGED" "$TEST_HOME/.ssh/config"; then
  fail "managed SSH config block still exists after uninstall"
fi

log "Read-only commands stay read-only after uninstall"
output="$(run_bgit list)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit status)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit active)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit prompt --plain)"
assert_equals "none" "$output"

output="$(run_bgit doctor)"
assert_contains "$output" "Config file not found"
assert_not_exists "$TEST_HOME/.bgit"
assert_equals "$TMPROOT/original-hooks" "$(git_global --get core.hooksPath)"

log "All integration tests passed"
