#!/usr/bin/env bash
# Checks that the bgit CLI and the desktop app work on the same machine and
# the same configuration, in either install order.
#
# The CLI is the real binary. The desktop app is its real backend, driven
# through the dev harness over HTTP (the same calls the window makes).
# Everything runs in an isolated home directory.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMPROOT="$(mktemp -d)"
BIN_DIR="$TMPROOT/bin"
TOOLS_DIR="$TMPROOT/tools"
HARNESS="$TMPROOT/bgit-desktop-harness"
PORT="$((20000 + RANDOM % 20000))"
TEST_HOME=""
APP_PID=""
KEEP_TMP="${BGIT_TEST_KEEP_TMP:-0}"

# Tests must never reach a real SSH agent.
unset SSH_AUTH_SOCK

cleanup() {
  local status=$?
  app_stop
  if [[ "$KEEP_TMP" == "1" || "$status" -ne 0 ]]; then
    printf '\n[test] preserving temp dir: %s\n' "$TMPROOT" >&2
    return
  fi
  rm -rf "$TMPROOT"
}
trap cleanup EXIT

log() { printf '\n==> %s\n' "$1"; }

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

assert_contains() {
  if [[ "$1" != *"$2"* ]]; then
    printf 'Expected to find: %s\nIn:\n%s\n' "$2" "$1" >&2
    exit 1
  fi
}

assert_not_contains() {
  if [[ "$1" == *"$2"* ]]; then
    printf 'Did not expect to find: %s\nIn:\n%s\n' "$2" "$1" >&2
    exit 1
  fi
}

assert_equals() {
  if [[ "$1" != "$2" ]]; then
    printf 'Expected: %s\nActual:   %s\n' "$1" "$2" >&2
    exit 1
  fi
}

# PATH with the CLI installed, and a PATH that has no bgit at all (the machine
# running the tests may have a real bgit installed).
cli_path() { printf '%s:%s' "$BIN_DIR" "$TOOLS_DIR"; }
no_cli_path() { printf '%s' "$TOOLS_DIR"; }

run_env() {
  local path="$1"
  shift
  env -i \
    HOME="$TEST_HOME" \
    XDG_CONFIG_HOME="$TEST_HOME/.config" \
    GIT_CONFIG_GLOBAL="$TEST_HOME/.gitconfig" \
    GIT_CONFIG_NOSYSTEM=1 \
    GIT_TERMINAL_PROMPT=0 \
    PATH="$path" \
    "$@"
}

cli() { run_env "$(cli_path)" bgit "$@"; }
git_home() { run_env "$(cli_path)" git "$@"; }

# Fingerprint of everything bgit manages on the machine.
machine_state() {
  local file
  for file in .bgit/config.toml .bgit/hooks/pre-push .gitconfig .ssh/config; do
    if [[ -f "$TEST_HOME/$file" ]]; then
      printf '%s %s\n' "$(cksum < "$TEST_HOME/$file")" "$file"
    else
      printf 'missing %s\n' "$file"
    fi
  done
}

# app_start cli|no-cli: starts the desktop backend the way the app starts.
app_start() {
  local path
  if [[ "$1" == "cli" ]]; then path="$(cli_path)"; else path="$(no_cli_path)"; fi
  (cd "$ROOT_DIR" && exec env -i PATH="$path" "$HARNESS" -addr "127.0.0.1:$PORT" -home "$TEST_HOME") > "$TMPROOT/app.log" 2>&1 &
  APP_PID=$!
  local attempt
  for attempt in $(seq 1 50); do
    if curl -s -o /dev/null "http://127.0.0.1:$PORT/"; then
      return
    fi
    sleep 0.1
  done
  cat "$TMPROOT/app.log" >&2
  fail "the desktop backend did not start"
}

app_stop() {
  if [[ -n "$APP_PID" ]]; then
    kill "$APP_PID" 2>/dev/null || true
    wait "$APP_PID" 2>/dev/null || true
    APP_PID=""
  fi
}

# app <Method> [json-args]: calls a desktop backend method like the window does.
app() {
  local method="$1"
  local args="${2:-[]}"
  local output
  output="$(curl -s -X POST -H 'Content-Type: application/json' -H 'X-Bgit-Harness: 1' --data "$args" "http://127.0.0.1:$PORT/__harness/call/$method")"
  if [[ "$output" == '{"error":'* ]]; then
    printf 'desktop call %s failed: %s\n' "$method" "$output" >&2
    exit 1
  fi
  printf '%s' "$output"
}

new_home() {
  TEST_HOME="$TMPROOT/$1"
  mkdir -p "$TEST_HOME/.ssh"
  chmod 700 "$TEST_HOME/.ssh"
}

new_key() {
  ssh-keygen -q -t ed25519 -N "" -f "$TEST_HOME/.ssh/$1" > /dev/null
  printf '%s' "$TEST_HOME/.ssh/$1"
}

new_repo() {
  local repo="$TEST_HOME/$1"
  mkdir -p "$repo"
  git_home -C "$repo" init -q -b main
  git_home -C "$repo" remote add origin "$2"
  printf '%s' "$repo"
}

log "Build the CLI and the desktop backend"
mkdir -p "$BIN_DIR" "$TOOLS_DIR"
cd "$ROOT_DIR"
GOCACHE="${GOCACHE:-/tmp/bgit-gocache}" go build -o "$BIN_DIR/bgit" .
GOCACHE="${GOCACHE:-/tmp/bgit-gocache}" go build -tags devharness -o "$HARNESS" ./desktop
for tool in git sh bash ssh ssh-keygen ssh-add env cat sed grep rm mkdir dirname basename uname; do
  if command -v "$tool" > /dev/null 2>&1; then
    ln -s "$(command -v "$tool")" "$TOOLS_DIR/$tool"
  fi
done

# ---------------------------------------------------------------------------
log "CLI first: set up identities, a bound repository, and a workspace"
new_home "cli-first"
git_home config --global user.name "Before Bgit"
git_home config --global user.email "before@example.com"

cli setup > /dev/null
cli add --alias company --name "Company Dev" --email "dev@company.example" --github "company-gh" --ssh-key "$(new_key id_company)" > /dev/null
cli add --alias personal --name "Personal Dev" --email "dev@personal.example" --github "personal-gh" --ssh-key "$(new_key id_personal)" > /dev/null
cli use company > /dev/null

REPO_ONE="$(new_repo "code/one" "https://github.com/org/one.git")"
(cd "$REPO_ONE" && cli bind --user personal > /dev/null)
mkdir -p "$TEST_HOME/ws"
cli workspace --path "$TEST_HOME/ws" --users company > /dev/null

BEFORE_APP="$(machine_state)"
assert_contains "$(cat "$TEST_HOME/.bgit/hooks/pre-push")" "bgit check --from-hook"
assert_not_contains "$(cat "$TEST_HOME/.bgit/hooks/pre-push")" "BGIT_HELPER"

log "Opening the app on a CLI machine changes nothing"
app_start cli
STATUS="$(app GetStatus)"
ONBOARDING="$(app GetOnboardingState)"
app GetDoctorStatus > /dev/null
REPOS="$(app ListRepositories)"
assert_equals "$BEFORE_APP" "$(machine_state)"

log "The app shows the CLI's identities, repository, workspace, and protection"
assert_contains "$ONBOARDING" '"needed":false'
assert_contains "$ONBOARDING" '"pushProtection":true'
assert_contains "$STATUS" '"activeAlias":"company"'
assert_contains "$STATUS" '"identityCount":2'
assert_contains "$STATUS" '"bindingCount":1'
assert_contains "$STATUS" '"workspaceCount":1'
assert_contains "$REPOS" "\"path\":\"$REPO_ONE\""
assert_contains "$REPOS" '"alias":"personal"'

log "The app and the CLI agree that the repository needs its remote fixed"
assert_contains "$REPOS" '"health":"attention"'
if (cd "$REPO_ONE" && cli check > "$TMPROOT/check.log" 2>&1); then
  fail "bgit check should fail before the remote is fixed"
fi
assert_contains "$(cat "$TMPROOT/check.log")" "origin remote is not aligned"

log "A fix made in the app satisfies the CLI check"
app FixRepository "[\"$REPO_ONE\"]" > /dev/null
assert_contains "$(cd "$REPO_ONE" && cli check 2>&1)" "Safety checks passed"
assert_equals "git@github.com-personal-gh:org/one.git" "$(git_home -C "$REPO_ONE" remote get-url origin)"

log "An identity added in the app works in the CLI"
app AddIdentity '[{"alias":"client","name":"Client Dev","email":"dev@client.example","githubUsername":"client-gh","sshKeyPath":"","generateSSHKey":true}]' > /dev/null
assert_contains "$(cli list 2>&1)" "client"
assert_contains "$(cli use client 2>&1)" "Switched to identity: client"
assert_contains "$(app GetStatus)" '"activeAlias":"client"'
assert_equals "dev@client.example" "$(git_home config --global user.email)"
assert_contains "$(cat "$TEST_HOME/.ssh/config")" "Host github.com-client-gh"

log "Changes made with the CLI show up in the app"
cli use personal > /dev/null
assert_contains "$(app GetStatus)" '"activeAlias":"personal"'
REPO_TWO="$(new_repo "code/two" "git@github.com-company-gh:org/two.git")"
(cd "$REPO_TWO" && cli bind --user company > /dev/null)
REPOS="$(app ListRepositories)"
assert_contains "$REPOS" "\"path\":\"$REPO_TWO\""
assert_contains "$(app GetStatus)" '"bindingCount":2'

log "A commit made in the app uses the identity the CLI bound"
printf 'hello\n' > "$REPO_ONE/file.txt"
COMMIT="$(app CommitChanges "[{\"path\":\"$REPO_ONE\",\"summary\":\"Add file\",\"description\":\"\",\"files\":[]}]")"
assert_contains "$COMMIT" '"ok":true'
assert_equals "Personal Dev <dev@personal.example>" "$(git_home -C "$REPO_ONE" log -1 --format='%an <%ae>')"

log "The CLI hook is still the CLI's hook"
assert_not_contains "$(cat "$TEST_HOME/.bgit/hooks/pre-push")" "BGIT_HELPER"
assert_contains "$(cli doctor 2>&1)" "Config file valid"
app_stop

# ---------------------------------------------------------------------------
log "CLI removed from PATH: the app keeps push protection working"
app_start no-cli
assert_contains "$(cat "$TEST_HOME/.bgit/hooks/pre-push")" "BGIT_HELPER="
assert_contains "$(app GetOnboardingState)" '"pushProtection":true'
app_stop

log "CLI installed again: its setup and checks still work"
cli setup > /dev/null
assert_not_contains "$(cat "$TEST_HOME/.bgit/hooks/pre-push")" "BGIT_HELPER"
assert_contains "$(cd "$REPO_ONE" && cli check 2>&1)" "Safety checks passed"

# ---------------------------------------------------------------------------
log "App first: set up and add an identity without the CLI"
new_home "app-first"
app_start no-cli
app RunSetup > /dev/null
app AddIdentity '[{"alias":"work","name":"Work Dev","email":"dev@work.example","githubUsername":"work-gh","sshKeyPath":"","generateSSHKey":true}]' > /dev/null
REPO_APP="$(new_repo "projects/app" "https://github.com/org/app.git")"
app AddRepository "[{\"path\":\"$REPO_APP\",\"alias\":\"work\"}]" > /dev/null
assert_contains "$(cat "$TEST_HOME/.bgit/hooks/pre-push")" "BGIT_HELPER="
app_stop

log "The CLI installed afterwards reads everything the app created"
assert_contains "$(cli list 2>&1)" "work"
assert_contains "$(cli active 2>&1)" "work"
assert_contains "$(cd "$REPO_APP" && cli check 2>&1)" "Safety checks passed"
assert_contains "$(cd "$REPO_APP" && cli status 2>&1)" "work"
cli add --alias side --name "Side Dev" --email "dev@side.example" --github "side-gh" --ssh-key skip > /dev/null

log "The app's hook prefers the CLI once it is installed"
if ! (cd "$REPO_APP" && run_env "$(cli_path)" sh "$TEST_HOME/.bgit/hooks/pre-push"); then
  fail "the app-written hook should pass through the installed CLI"
fi

app_start cli
assert_contains "$(app GetStatus)" '"identityCount":2'
assert_contains "$(app GetOnboardingState)" '"needed":false'

# ---------------------------------------------------------------------------
log "Uninstalling with the CLI returns the app to first-run state"
cli uninstall --force --skip-repos > /dev/null
ONBOARDING="$(app GetOnboardingState)"
assert_contains "$ONBOARDING" '"configured":false'
assert_contains "$ONBOARDING" '"needed":true'
app GetStatus > /dev/null
app GetDoctorStatus > /dev/null
if [[ -e "$TEST_HOME/.bgit" ]]; then
  fail "reading state in the app must not recreate the bgit folder"
fi

log "Removing bgit in the app leaves the CLI unconfigured and Git restored"
app RunSetup > /dev/null
app AddIdentity '[{"alias":"again","name":"Again Dev","email":"dev@again.example","githubUsername":"again-gh","sshKeyPath":"","generateSSHKey":true}]' > /dev/null
assert_contains "$(cli list 2>&1)" "again"
REMOVED="$(app UninstallBgit '[{"removeKeys":false}]')"
assert_contains "$REMOVED" '"configRemoved":true'
assert_contains "$(cli status 2>&1)" "bgit is not configured."
assert_equals "" "$(git_home config --global --get core.hooksPath || true)"
app_stop

log "All coexistence tests passed"
