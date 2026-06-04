#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMPROOT="$(mktemp -d)"
MACHINE_A_HOME="$TMPROOT/machine-a"
MACHINE_B_HOME="$TMPROOT/machine-b"
BIN="$TMPROOT/bgit"
LAST_OUTPUT=""
KEEP_TMP="${BGIT_TEST_KEEP_TMP:-0}"
PASSWORD="portable-password"

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

run_home() {
  local home="$1"
  shift
  HOME="$home" \
  XDG_CONFIG_HOME="$home/.xdg-config" \
  GIT_CONFIG_NOSYSTEM=1 \
  "$@"
}

run_bgit_ok() {
  local home="$1"
  shift
  local output
  if ! output="$(run_home "$home" "$BIN" "$@" 2>&1)"; then
    LAST_OUTPUT="$output"
    fail "command failed: bgit $*"
  fi
  LAST_OUTPUT="$output"
  printf '%s' "$output"
}

run_bgit_ok_with_input() {
  local home="$1"
  local input="$2"
  shift 2
  local output
  if ! output="$(printf '%s' "$input" | run_home "$home" "$BIN" "$@" 2>&1)"; then
    LAST_OUTPUT="$output"
    fail "command failed: bgit $*"
  fi
  LAST_OUTPUT="$output"
  printf '%s' "$output"
}

generate_key() {
  local home="$1"
  local name="$2"
  mkdir -p "$home/.ssh"
  chmod 700 "$home/.ssh"
  run_home "$home" ssh-keygen -t ed25519 -f "$home/.ssh/$name" -N "" -C "$name@bgit" >/dev/null
}

log "Build test binary"
cd "$ROOT_DIR"
GOCACHE="${GOCACHE:-/tmp/bgit-gocache}" go build -o "$BIN" .

mkdir -p "$MACHINE_A_HOME" "$MACHINE_B_HOME"

log "Machine A: configure identities with SSH keys"
run_bgit_ok "$MACHINE_A_HOME" setup >/dev/null
generate_key "$MACHINE_A_HOME" "company_key"
generate_key "$MACHINE_A_HOME" "personal_key"

output="$(run_bgit_ok "$MACHINE_A_HOME" add --alias company --name "Byterings Admin" --email "admin@byterings.test" --github "byterings" --ssh-key "$MACHINE_A_HOME/.ssh/company_key")"
assert_contains "$output" "User 'company' added successfully"

output="$(run_bgit_ok "$MACHINE_A_HOME" add --alias personal --name "Personal User" --email "personal@example.test" --github "personal-user" --ssh-key "$MACHINE_A_HOME/.ssh/personal_key")"
assert_contains "$output" "User 'personal' added successfully"

output="$(run_bgit_ok "$MACHINE_A_HOME" use company)"
assert_contains "$output" "Switched to identity: company"

run_bgit_ok "$MACHINE_A_HOME" setup >/dev/null
assert_contains "$(cat "$MACHINE_A_HOME/.ssh/config")" "Host github.com-byterings"
assert_contains "$(cat "$MACHINE_A_HOME/.ssh/config")" "Host github.com-personal-user"

log "Machine A: export encrypted portable backup"
output="$(run_bgit_ok_with_input "$MACHINE_A_HOME" $'portable-password\nportable-password\n' export)"
assert_contains "$output" "Created bgit export archive"

ARCHIVE_PATH="$(find "$MACHINE_A_HOME/.bgit/backups" -maxdepth 1 -type f -name '*.bgit' | head -n 1)"
assert_exists "$ARCHIVE_PATH"
PORTABLE_ARCHIVE="$TMPROOT/portable.bgit"
cp "$ARCHIVE_PATH" "$PORTABLE_ARCHIVE"

if tar -tzf "$PORTABLE_ARCHIVE" >/dev/null 2>&1; then
  fail "expected portable backup to be encrypted"
fi

log "Machine B: import into clean HOME"
output="$(run_bgit_ok_with_input "$MACHINE_B_HOME" $'portable-password\n' import "$PORTABLE_ARCHIVE")"
assert_contains "$output" "Imported bgit archive"
assert_contains "$output" "Users restored: 2"
assert_contains "$output" "Active user: company"

output="$(run_bgit_ok "$MACHINE_B_HOME" list)"
assert_contains "$output" "company"
assert_contains "$output" "personal"

output="$(run_bgit_ok "$MACHINE_B_HOME" active)"
assert_contains "$output" "Active user: company"

MACHINE_B_CONFIG="$MACHINE_B_HOME/.bgit/config.toml"
assert_exists "$MACHINE_B_CONFIG"
assert_contains "$(cat "$MACHINE_B_CONFIG")" 'active_user = "company"'
assert_contains "$(cat "$MACHINE_B_CONFIG")" 'name = "Byterings Admin"'
assert_contains "$(cat "$MACHINE_B_CONFIG")" 'email = "admin@byterings.test"'
assert_contains "$(cat "$MACHINE_B_CONFIG")" 'github_username = "byterings"'
assert_contains "$(cat "$MACHINE_B_CONFIG")" 'name = "Personal User"'
assert_contains "$(cat "$MACHINE_B_CONFIG")" 'email = "personal@example.test"'
assert_contains "$(cat "$MACHINE_B_CONFIG")" 'github_username = "personal-user"'
assert_contains "$(cat "$MACHINE_B_CONFIG")" "ssh_key_path = \"$MACHINE_B_HOME/.ssh/bgit_company\""
assert_contains "$(cat "$MACHINE_B_CONFIG")" "ssh_key_path = \"$MACHINE_B_HOME/.ssh/bgit_personal\""
assert_not_contains "$(cat "$MACHINE_B_CONFIG")" "$MACHINE_A_HOME"

log "Machine B: verify restored SSH keys"
assert_exists "$MACHINE_B_HOME/.ssh/bgit_company"
assert_exists "$MACHINE_B_HOME/.ssh/bgit_company.pub"
assert_exists "$MACHINE_B_HOME/.ssh/bgit_personal"
assert_exists "$MACHINE_B_HOME/.ssh/bgit_personal.pub"
assert_equals "600" "$(stat -c '%a' "$MACHINE_B_HOME/.ssh/bgit_company")"
assert_equals "600" "$(stat -c '%a' "$MACHINE_B_HOME/.ssh/bgit_personal")"

log "Machine B: regenerate SSH config and validate functionality"
output="$(run_bgit_ok "$MACHINE_B_HOME" setup)"
assert_contains "$output" "Setup complete"

MACHINE_B_SSH_CONFIG="$MACHINE_B_HOME/.ssh/config"
assert_exists "$MACHINE_B_SSH_CONFIG"
assert_contains "$(cat "$MACHINE_B_SSH_CONFIG")" "BEGIN BGIT MANAGED"
assert_contains "$(cat "$MACHINE_B_SSH_CONFIG")" "Host github.com-byterings"
assert_contains "$(cat "$MACHINE_B_SSH_CONFIG")" "IdentityFile $MACHINE_B_HOME/.ssh/bgit_company"
assert_contains "$(cat "$MACHINE_B_SSH_CONFIG")" "Host github.com-personal-user"
assert_contains "$(cat "$MACHINE_B_SSH_CONFIG")" "IdentityFile $MACHINE_B_HOME/.ssh/bgit_personal"

output="$(run_bgit_ok "$MACHINE_B_HOME" status)"
assert_contains "$output" "company"
assert_contains "$output" "admin@byterings.test"

log "Backup portability test passed"
