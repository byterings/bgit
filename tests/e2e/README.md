# bgit E2E Tests

This folder contains Docker-based end-to-end tests for bgit.

The normal E2E suite is safe to run after every code change:

- It runs inside Docker.
- It builds bgit from the current source tree.
- It uses an isolated `HOME` inside the container.
- It does not touch your real `~/.gitconfig`, `~/.ssh`, PATH, or GitHub accounts.
- It uses fake identities and local Git repositories only.

## Commands

Run unit tests:

```bash
make test
```

Run Docker E2E tests:

```bash
make e2e
```

Run everything:

```bash
make test-all
```

Run after every change:

```bash
watchexec -e go,sh,yml 'make test-all'
```

## What The E2E Suite Covers

The E2E script exercises:

- `bgit --version`
- `bgit setup`
- `bgit add`
- `bgit list`
- `bgit active`
- `bgit use`
- `bgit status`
- `bgit doctor`
- `bgit workspace`
- `bgit bind`
- `bgit sync`
- `bgit check`
- `bgit remote fix`
- `bgit remote restore`
- `bgit delete`
- `bgit uninstall`

It also includes a regression flow for GitHub issue #5:

1. Install/setup bgit.
2. Add fake identities.
3. Use a bgit identity.
4. Manually change global Git identity to simulate GitHub Desktop/user control.
5. Run uninstall.
6. Verify bgit-owned config is removed.
7. Verify non-bgit Git identity is preserved.
8. Verify normal Git commit/push still works against a local bare repo.

## Debugging

The script prints each command and assertion. On failure, it prints the path to the verbose log inside the container.

To debug interactively:

```bash
docker build -f tests/e2e/Dockerfile -t bgit-e2e:local .
docker run --rm -it --entrypoint bash bgit-e2e:local
bash tests/e2e/run.sh
```

## Real GitHub Smoke Tests

Real GitHub smoke tests live in `tests/e2e-real/run.sh` and are disabled by default.

They run only when all required environment variables are present:

- `BGIT_TEST_GITHUB_TOKEN`
- `BGIT_TEST_REPO`
- `BGIT_TEST_SSH_KEY_PERSONAL`
- `BGIT_TEST_SSH_KEY_WORK`

GitHub Actions can be run manually from the Actions tab with `workflow_dispatch`.
