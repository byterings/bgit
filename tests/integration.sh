#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMPROOT="$(mktemp -d)"
TEST_HOME="$TMPROOT/home"
TEST_XDG_CONFIG_HOME="$TMPROOT/xdg-config"
BIN="$TMPROOT/bgit"
LAST_OUTPUT=""
LAST_STATUS=0
KEEP_TMP="${BGIT_TEST_KEEP_TMP:-0}"

cleanup() {
  local status=$?
  if [[ "$KEEP_TMP" == "1" || "$status" -ne 0 ]]; then
    printf '\n[test] preserving temp dir: %s\n' "$TMPROOT" >&2
    return
  fi
  rm -rf "$TMPROOT"
}
trap cleanup EXIT

log() {
  printf '\n==> %s\n' "$1"
}

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  if [[ -n "${LAST_OUTPUT:-}" ]]; then
    printf '\nLast command output:\n%s\n' "$LAST_OUTPUT" >&2
  fi
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  if [[ "$haystack" == *"$needle"* ]]; then
    printf 'Did not expect output to contain: %s\nActual output:\n%s\n' "$needle" "$haystack" >&2
    exit 1
  fi
}

assert_not_exists() {
  local path="$1"
  if [[ -e "$path" ]]; then
    fail "expected path to be absent: $path"
  fi
}

assert_exists() {
  local path="$1"
  if [[ ! -e "$path" ]]; then
    fail "expected path to exist: $path"
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

run_env() {
  HOME="$TEST_HOME" \
  XDG_CONFIG_HOME="$TEST_XDG_CONFIG_HOME" \
  GIT_CONFIG_NOSYSTEM=1 \
  "$@"
}

run_bgit() {
  run_env "$BIN" "$@"
}

run_bgit_ok() {
  local output
  if ! output="$(run_bgit "$@" 2>&1)"; then
    LAST_OUTPUT="$output"
    fail "command failed: bgit $*"
  fi
  LAST_OUTPUT="$output"
  printf '%s' "$output"
}

run_bgit_fail() {
  local output
  if output="$(run_bgit "$@" 2>&1)"; then
    LAST_OUTPUT="$output"
    fail "command unexpectedly succeeded: bgit $*"
  fi
  LAST_OUTPUT="$output"
  printf '%s' "$output"
}

git_global() {
  run_env git config --global "$@"
}

write_config() {
  local content="$1"
  mkdir -p "$TEST_HOME/.bgit"
  printf '%s\n' "$content" > "$TEST_HOME/.bgit/config.toml"
}

log "Build test binary"
cd "$ROOT_DIR"
GOCACHE="${GOCACHE:-/tmp/bgit-gocache}" go build -o "$BIN" .

mkdir -p "$TEST_HOME" "$TEST_XDG_CONFIG_HOME"

log "Read-only commands do not initialize bgit"
output="$(run_bgit_ok list)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit_ok status)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit_ok active)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit_ok prompt --plain)"
assert_equals "none" "$output"

output="$(run_bgit_ok doctor)"
assert_contains "$output" "Config file not found"
assert_not_exists "$TEST_HOME/.bgit"

log "Invalid config is rejected"
write_config 'version = "999.0"'
output="$(run_bgit_fail list)"
assert_contains "$output" "unsupported config version"
rm -rf "$TEST_HOME/.bgit"

write_config '
version = "1.0"
[[users]]
alias = "dup"
name = "One"
email = "dup@example.com"
github_username = "dup1"

[[users]]
alias = "dup"
name = "Two"
email = "dup2@example.com"
github_username = "dup2"
'
output="$(run_bgit_fail status)"
assert_contains "$output" "duplicate user alias 'dup'"
rm -rf "$TEST_HOME/.bgit"

log "Fresh setup preserves previous hooks path"
mkdir -p "$TMPROOT/original-hooks"
git_global user.name "Original User"
git_global user.email "original@example.com"
git_global core.hooksPath "$TMPROOT/original-hooks"

output="$(run_bgit_ok setup)"
assert_contains "$output" "Setup complete"
assert_equals "$TEST_HOME/.bgit/hooks" "$(git_global --get core.hooksPath)"

log "Add and use first identity"
output="$(run_bgit_ok add --alias company --name "Byterings Admin" --email "byterings@gmail.com" --github "byterings" --ssh-key skip)"
assert_contains "$output" "User 'company' added successfully"

output="$(run_bgit_ok add --alias personal --name "Personal User" --email "personal@example.com" --github "personal-user" --ssh-key skip)"
assert_contains "$output" "User 'personal' added successfully"

output="$(run_bgit_ok list)"
assert_contains "$output" "company"
assert_contains "$output" "personal"

output="$(run_bgit_ok use company)"
assert_contains "$output" "Switched to identity: company"
assert_equals "Byterings Admin" "$(git_global --get user.name)"
assert_equals "byterings@gmail.com" "$(git_global --get user.email)"

output="$(run_bgit_ok active)"
assert_contains "$output" "Active user: company"

output="$(run_bgit_ok prompt --plain)"
assert_equals "company" "$output"

log "Workspace and repo binding flows"
WORKSPACES="$TMPROOT/workspaces"
mkdir -p "$WORKSPACES"
output="$(run_bgit_ok workspace --path "$WORKSPACES" --users company)"
assert_contains "$output" "Workspace ready"

output="$(run_bgit_ok workspace --list)"
assert_contains "$output" "$WORKSPACES/company"

REPO="$TMPROOT/repo"
mkdir -p "$REPO"
run_env git -C "$REPO" init
run_env git -C "$REPO" remote add origin git@github.com:byterings/bgit.git

cd "$REPO"
output="$(run_bgit_ok bind --user company)"
assert_contains "$output" "Bound repository to 'company'"

output="$(run_bgit_ok bind --user company)"
assert_contains "$output" "Repository already bound to 'company'"

output="$(run_bgit_fail bind --user personal)"
assert_contains "$output" "repository already bound to 'company'"

output="$(run_bgit_ok bind --user personal --force)"
assert_contains "$output" "Overriding existing binding from 'company' to 'personal'"
assert_contains "$output" "Bound repository to 'personal'"

output="$(run_bgit_ok bind --user company --force)"
assert_contains "$output" "Overriding existing binding from 'personal' to 'company'"

output="$(run_bgit_ok bind --remove)"
assert_contains "$output" "Removed binding for 'company'"

output="$(run_bgit_ok bind --user company)"
assert_contains "$output" "Bound repository to 'company'"

log "Remote fix and safety check"
output="$(run_bgit_ok remote fix)"
assert_contains "$output" "Remote fixed for user 'company'"
assert_equals "git@github.com-byterings:byterings/bgit.git" "$(run_env git -C "$REPO" remote get-url origin)"

output="$(run_bgit_ok remote fix)"
assert_contains "$output" "Remote URL already configured for company"

output="$(run_bgit_ok check)"
assert_contains "$output" "Safety checks passed"

log "Export archive layout"
cd "$ROOT_DIR"
output="$(run_bgit_ok export)"
assert_contains "$output" "Created bgit export archive"

BACKUP_DIR="$TEST_HOME/.bgit/backups"
ARCHIVE_PATH="$(find "$BACKUP_DIR" -maxdepth 1 -type f -name '*.bgit' | head -n 1)"
assert_exists "$ARCHIVE_PATH"

ARCHIVE_LISTING="$(tar -tzf "$ARCHIVE_PATH")"
assert_contains "$ARCHIVE_LISTING" "manifest.json"
assert_contains "$ARCHIVE_LISTING" "payload/"
assert_contains "$ARCHIVE_LISTING" "payload/config/"
assert_contains "$ARCHIVE_LISTING" "payload/config/config.toml"
assert_contains "$ARCHIVE_LISTING" "payload/keys/"

MANIFEST_CONTENT="$(tar -xOzf "$ARCHIVE_PATH" manifest.json)"
assert_contains "$MANIFEST_CONTENT" '"format_version": "1"'
assert_contains "$MANIFEST_CONTENT" '"layout_version": "1"'
assert_contains "$MANIFEST_CONTENT" '"status": "plaintext"'
assert_contains "$MANIFEST_CONTENT" '"planned_version": "R-010"'
assert_contains "$MANIFEST_CONTENT" '"alias": "company"'

CONFIG_CONTENT="$(tar -xOzf "$ARCHIVE_PATH" payload/config/config.toml)"
assert_contains "$CONFIG_CONTENT" 'active_user = "company"'
assert_contains "$CONFIG_CONTENT" 'alias = "company"'

log "Workspace removal path"
cd "$ROOT_DIR"
output="$(run_bgit_ok workspace --remove company)"
assert_contains "$output" "Removed workspace binding for 'company'"

output="$(run_bgit_ok workspace --list)"
assert_not_contains "$output" "$WORKSPACES/company"

log "Uninstall restores Git state and removes managed config"
cd "$REPO"
output="$(run_bgit_ok uninstall --force)"
assert_contains "$output" "bgit uninstall complete"
assert_equals "$TMPROOT/original-hooks" "$(git_global --get core.hooksPath)"
assert_equals "Original User" "$(git_global --get user.name)"
assert_equals "original@example.com" "$(git_global --get user.email)"
assert_equals "git@github.com:byterings/bgit.git" "$(run_env git -C "$REPO" remote get-url origin)"
assert_not_exists "$TEST_HOME/.bgit"

if [[ -f "$TEST_HOME/.ssh/config" ]] && grep -q "BEGIN BGIT MANAGED" "$TEST_HOME/.ssh/config"; then
  fail "managed SSH config block still exists after uninstall"
fi

log "Read-only commands stay read-only after uninstall"
output="$(run_bgit_ok list)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit_ok status)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit_ok active)"
assert_contains "$output" "bgit is not configured."

output="$(run_bgit_ok prompt --plain)"
assert_equals "none" "$output"

output="$(run_bgit_ok doctor)"
assert_contains "$output" "Config file not found"
assert_not_exists "$TEST_HOME/.bgit"
assert_equals "$TMPROOT/original-hooks" "$(git_global --get core.hooksPath)"

log "All integration tests passed"
