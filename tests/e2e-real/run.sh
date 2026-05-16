#!/usr/bin/env bash
set -Eeuo pipefail

required=(
  BGIT_TEST_GITHUB_TOKEN
  BGIT_TEST_REPO
  BGIT_TEST_SSH_KEY_PERSONAL
  BGIT_TEST_SSH_KEY_WORK
)

for name in "${required[@]}"; do
  if [ -z "${!name:-}" ]; then
    echo "Skipping real GitHub smoke tests"
    exit 0
  fi
done

echo "Running real GitHub smoke tests against $BGIT_TEST_REPO"
echo "Real smoke test harness is intentionally minimal; extend this only with disposable test repos/accounts."

tmp_home="$(mktemp -d)"
trap 'rm -rf "$tmp_home"' EXIT

export HOME="$tmp_home"
mkdir -p "$HOME/.ssh"
chmod 700 "$HOME/.ssh"

printf '%s\n' "$BGIT_TEST_SSH_KEY_PERSONAL" > "$HOME/.ssh/bgit_personal"
printf '%s\n' "$BGIT_TEST_SSH_KEY_WORK" > "$HOME/.ssh/bgit_work"
chmod 600 "$HOME/.ssh/bgit_personal" "$HOME/.ssh/bgit_work"

go build -o /tmp/bgit-real-smoke .
/tmp/bgit-real-smoke --version

echo "Real GitHub smoke test prerequisites are present."
echo "Skipping destructive repo operations by default."
