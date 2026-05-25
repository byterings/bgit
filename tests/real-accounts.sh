#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REAL_HOME="${HOME:?HOME is required}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
BACKUP_DIR="${BGIT_REAL_BACKUP_DIR:-$REAL_HOME/.bgit-real-test-backup-$TIMESTAMP}"
TMPROOT="$(mktemp -d)"
BIN="$TMPROOT/bgit-real-test"
RESTORED=0
RESTORE_READY=0

required_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "Missing required environment variable: $name" >&2
    exit 1
  fi
}

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

assert_equals() {
  local expected="$1"
  local actual="$2"
  if [[ "$expected" != "$actual" ]]; then
    printf 'Expected: %s\nActual:   %s\n' "$expected" "$actual" >&2
    exit 1
  fi
}

run_bgit() {
  HOME="$REAL_HOME" "$BIN" "$@"
}

git_global_get() {
  HOME="$REAL_HOME" git config --global --get "$1" 2>/dev/null || true
}

git_global_set_or_unset() {
  local key="$1"
  local value_file="$2"
  local value
  value="$(cat "$value_file")"
  if [[ -n "$value" ]]; then
    HOME="$REAL_HOME" git config --global "$key" "$value"
  else
    HOME="$REAL_HOME" git config --global --unset "$key" >/dev/null 2>&1 || true
  fi
}

restore_state() {
  if [[ "$RESTORE_READY" != "1" ]]; then
    rm -rf "$TMPROOT"
    return
  fi

  if [[ "$RESTORED" == "1" ]]; then
    return
  fi
  RESTORED=1

  log "Restoring original bgit, SSH, and Git state"

  rm -rf "$REAL_HOME/.bgit"
  if [[ -d "$BACKUP_DIR/.bgit" ]]; then
    cp -a "$BACKUP_DIR/.bgit" "$REAL_HOME/.bgit"
  fi

  mkdir -p "$REAL_HOME/.ssh"
  if [[ -f "$BACKUP_DIR/ssh_config" ]]; then
    cp -a "$BACKUP_DIR/ssh_config" "$REAL_HOME/.ssh/config"
  else
    rm -f "$REAL_HOME/.ssh/config"
  fi

  git_global_set_or_unset user.name "$BACKUP_DIR/git_user_name"
  git_global_set_or_unset user.email "$BACKUP_DIR/git_user_email"
  git_global_set_or_unset core.hooksPath "$BACKUP_DIR/git_hooks_path"

  rm -rf "$TMPROOT"
  echo "Restored from backup: $BACKUP_DIR"
}

trap restore_state EXIT

required_env BGIT_TEST_ALIAS_1
required_env BGIT_TEST_ALIAS_2
required_env BGIT_TEST_REPO_1
required_env BGIT_TEST_REPO_2

if [[ ! -f "$REAL_HOME/.bgit/config.toml" ]]; then
  fail "real bgit config not found at $REAL_HOME/.bgit/config.toml"
fi

if [[ "${BGIT_REAL_CONFIRM:-}" != "YES" ]]; then
  cat >&2 <<'MSG'
This test touches your real bgit, Git, SSH, and GitHub repo state.
It creates a backup and restores it at the end, but you must opt in explicitly.

Required:
  BGIT_REAL_CONFIRM=YES
  BGIT_TEST_ALIAS_1=<alias from ~/.bgit/config.toml>
  BGIT_TEST_ALIAS_2=<alias from ~/.bgit/config.toml>
  BGIT_TEST_REPO_1=<GitHub test repo URL for alias 1>
  BGIT_TEST_REPO_2=<GitHub test repo URL for alias 2>

Optional:
  BGIT_REAL_PUSH=1       # create and push empty test commits to throwaway branches
MSG
  exit 1
fi

log "Backing up current state"
mkdir -p "$BACKUP_DIR"
cp -a "$REAL_HOME/.bgit" "$BACKUP_DIR/.bgit"
if [[ -f "$REAL_HOME/.ssh/config" ]]; then
  cp -a "$REAL_HOME/.ssh/config" "$BACKUP_DIR/ssh_config"
fi
git_global_get user.name > "$BACKUP_DIR/git_user_name"
git_global_get user.email > "$BACKUP_DIR/git_user_email"
git_global_get core.hooksPath > "$BACKUP_DIR/git_hooks_path"
RESTORE_READY=1

extract_user_field() {
  local alias="$1"
  local field="$2"
  awk -v alias="$alias" -v field="$field" '
    /^\[\[users\]\]/ {
      if (in_user && found) exit
      in_user=1
      found=0
      next
    }
    in_user && $0 ~ "^[[:space:]]*" field "[[:space:]]*=" {
      value=$0
      sub("^[^=]*=[[:space:]]*", "", value)
      gsub(/^"|"$/, "", value)
      if (found) {
        print value
        exit
      }
    }
    in_user && /^[[:space:]]*alias[[:space:]]*=/ {
      value=$0
      sub("^[^=]*=[[:space:]]*", "", value)
      gsub(/^"|"$/, "", value)
      if (value == alias) found=1
    }
  ' "$BACKUP_DIR/.bgit/config.toml"
}

add_identity_from_backup() {
  local alias="$1"
  local name email github ssh_key

  name="$(extract_user_field "$alias" name)"
  email="$(extract_user_field "$alias" email)"
  github="$(extract_user_field "$alias" github_username)"
  ssh_key="$(extract_user_field "$alias" ssh_key_path)"

  [[ -n "$name" ]] || fail "could not read name for alias $alias"
  [[ -n "$email" ]] || fail "could not read email for alias $alias"
  [[ -n "$github" ]] || fail "could not read github_username for alias $alias"
  [[ -n "$ssh_key" ]] || fail "could not read ssh_key_path for alias $alias"
  [[ -f "$ssh_key" ]] || fail "SSH key for $alias does not exist: $ssh_key"

  run_bgit add \
    --alias "$alias" \
    --name "$name" \
    --email "$email" \
    --github "$github" \
    --ssh-key "$ssh_key"
}

test_alias_repo() {
  local alias="$1"
  local repo_url="$2"
  local workdir="$TMPROOT/work-$alias"
  local branch="bgit-real-test-$alias-$TIMESTAMP"
  local output

  log "Testing alias '$alias' with repo '$repo_url'"
  output="$(run_bgit use "$alias")"
  assert_contains "$output" "Switched to identity: $alias"

  output="$(run_bgit clone "$repo_url" "$workdir")"
  assert_contains "$output" "Repository cloned successfully"

  cd "$workdir"
  output="$(run_bgit status)"
  assert_contains "$output" "Effective Identity"

  output="$(run_bgit active)"
  assert_contains "$output" "Active user: $alias"

  output="$(run_bgit prompt --plain)"
  assert_equals "$alias" "$output"

  output="$(run_bgit bind --user "$alias" --force)"
  if [[ "$output" != *"Bound repository to '$alias'"* && "$output" != *"Repository already bound to '$alias'"* ]]; then
    printf 'Unexpected bind output:\n%s\n' "$output" >&2
    exit 1
  fi

  output="$(run_bgit remote fix)"
  if [[ "$output" != *"Remote fixed for user '$alias'"* && "$output" != *"Remote URL already configured for $alias"* ]]; then
    printf 'Unexpected remote fix output:\n%s\n' "$output" >&2
    exit 1
  fi

  output="$(run_bgit check)"
  assert_contains "$output" "Safety checks passed"

  HOME="$REAL_HOME" git fetch origin >/dev/null

  if [[ "${BGIT_REAL_PUSH:-}" == "1" ]]; then
    HOME="$REAL_HOME" git switch -c "$branch"
    HOME="$REAL_HOME" git commit --allow-empty -m "bgit real acceptance test $TIMESTAMP"
    HOME="$REAL_HOME" git push origin "$branch"
  fi
}

log "Build local test binary"
cd "$ROOT_DIR"
GOCACHE="${GOCACHE:-/tmp/bgit-gocache}" go build -o "$BIN" .

log "Uninstall current bgit state with local binary"
output="$(run_bgit uninstall --force)"
assert_contains "$output" "bgit uninstall complete"

output="$(run_bgit list)"
assert_contains "$output" "bgit is not configured."

log "Re-add selected identities from backup"
add_identity_from_backup "$BGIT_TEST_ALIAS_1"
add_identity_from_backup "$BGIT_TEST_ALIAS_2"

output="$(run_bgit list)"
assert_contains "$output" "$BGIT_TEST_ALIAS_1"
assert_contains "$output" "$BGIT_TEST_ALIAS_2"

test_alias_repo "$BGIT_TEST_ALIAS_1" "$BGIT_TEST_REPO_1"
test_alias_repo "$BGIT_TEST_ALIAS_2" "$BGIT_TEST_REPO_2"

log "Final uninstall after real-account test"
output="$(run_bgit uninstall --force)"
assert_contains "$output" "bgit uninstall complete"

output="$(run_bgit status)"
assert_contains "$output" "bgit is not configured."

log "Real-account acceptance test passed"
