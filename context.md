# Project Context

## Overview
bgit is a Go CLI for managing multiple Git identities on one machine. It switches global Git commit identity, manages SSH host aliases, supports repo/workspace identity resolution, and installs a global pre-push safety hook.

## Architecture
- CLI: Cobra commands in `cmd/`.
- Config: TOML config stored at `~/.bgit/config.toml`.
- Identity resolution: workspace, then repo binding, then global active user.
- Git integration: global `user.name`, `user.email`, and `core.hooksPath`.
- SSH integration: managed section in `~/.ssh/config` plus optional bgit-generated SSH keys.
- Database: none.
- Frontend: none.

## Key Modules
- `core/config`: config schema, config path helpers, persistence, legacy migration, active user state, and workspace/binding state mutation.
- `core/config` now validates config structure on load/save, normalizes legacy values, writes atomically via temp-file rename, and enforces supported config versions.
- `core/identity`: identity resolution plus identity add/update/delete/activate flows.
- `core/models`: shared reusable domain and result structs used across config, identity, repo, and SSH core packages.
- `core/export`: archive-generation logic for `.bgit` backups, including manifest generation and stable payload layout packaging for the later encryption/import work.
- `core/repo`: workspace and binding operations, remote URL conversion, clone auto-bind support, and repository owner resolution for safety checks.
- `core/ssh`: SSH key generation/validation, managed `~/.ssh/config` updates, SSH agent helpers, and GitHub SSH connectivity checks.
- Core APIs now prefer structured result objects from `core/models` over mixed tuple-style returns for mutations and operational checks.
- `cmd/export.go`: creates `.bgit` archives in the managed backup directory using the current export manifest/layout format. The archive is plaintext today and is intentionally shaped so `R-010` can add encryption around the existing payload.
- `cmd/setup.go`: initializes config, installs managed pre-push hook, stores previous hook path for later restore.
- `cmd/use.go`: switches active identity and stores the pre-bgit Git identity before the first managed switch.
- `cmd/uninstall.go`: restores repo remotes, removes managed SSH config, restores hooks/Git identity when possible, and removes bgit config.
- `internal/git`: global Git config wrappers.

## Testing
- `make test`: runs `go test ./...` with an isolated Go build cache under `/tmp`.
- `make test-integration`: runs real CLI integration tests on the host using isolated `HOME`/`XDG_CONFIG_HOME`, `GIT_CONFIG_NOSYSTEM=1`, disposable repositories, config-validation regression cases, and cleanup preservation via `BGIT_TEST_KEEP_TMP=1` when needed.
- `make test-docker`: builds `Dockerfile.test` and runs the same integration suite inside a disposable Linux container.
- `make test-real`: optional real-account acceptance test that requires explicit environment variables and `BGIT_REAL_CONFIRM=YES`; it backs up and restores the user's real bgit, SSH, and Git state.
- `docker-compose.test.yml` is available for manual Compose-based runs, but the Makefile uses plain Docker so Compose is not required.

## Uninstall And Recovery
Uninstall must support both new and old configs:
- New configs can restore previous Git identity and previous `core.hooksPath` from stored metadata.
- Old configs may not have restore metadata, so uninstall uses configured bindings/workspaces and warns with exact manual Git identity commands when the current global Git identity still matches a bgit user.
- SSH key deletion is opt-in through `bgit uninstall --remove-keys`.
- Read-only commands such as `list`, `status`, `active`, `prompt`, and `doctor` must not auto-create config after uninstall. They load config read-only and report that setup is needed when bgit is not configured.

## Constraints
- Code is the source of truth.
- `project_status.md` records actual implementation state.
- Roadmap items are planned only until verified in code.
