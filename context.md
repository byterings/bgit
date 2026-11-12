# Project Context

## Overview
bgit is a Go CLI for managing multiple Git identities on one machine. It switches global Git commit identity, manages SSH host aliases, supports repo/workspace identity resolution, and installs a global pre-push safety hook.

## Architecture
- CLI: Cobra commands in `cmd/`.
- Config: TOML config stored at `~/.bgit/config.toml`.
- Identity resolution: workspace, then repo binding, then global active user.
- Git integration: global `user.name`, `user.email`, and `core.hooksPath`; bound repositories also get repo-local `user.name` and `user.email` so normal `git commit` uses the bound identity even when the global active identity differs.
- SSH integration: managed section in `~/.ssh/config` plus optional bgit-generated SSH keys.
- Database: none.
- Frontend: Wails desktop app under `desktop/` with a first-run wizard and pages for Dashboard, Repositories (with a per-repository view for changes, history, branches, and sync), Identities, Workspaces, Backup & Restore, Doctor, and Settings.

## Key Modules
- `core/config`: config schema, config path helpers, persistence, legacy migration, active user state, and workspace/binding state mutation.
- `core/config` now validates config structure on load/save, normalizes legacy values, writes atomically via temp-file rename, and enforces supported config versions.
- `core/identity`: identity resolution plus identity add/update/delete/activate flows.
- `core/models`: shared reusable domain and result structs used across config, identity, repo, and SSH core packages.
- `core/export`: archive-generation and import logic for `.bgit` backups. It preserves the `R-009` inner tar+gzip archive layout, stores configured identity SSH key pairs under `payload/keys/`, wraps that payload in an encrypted outer file envelope, and can decrypt/import the archived config and keys onto another machine.
- `core/repo`: workspace and binding operations, remote URL conversion, clone auto-bind support, and repository owner resolution for safety checks.
- `core/setup`: one-time setup shared by the CLI and desktop app: config initialization, managed pre-push hook generation and installation, and global `core.hooksPath` get/set/restore. The hook can carry a fallback helper binary path for installs without `bgit` on PATH; with no helper it is the classic CLI hook.
- `core/check`: push safety decisions shared by `bgit check` and the desktop app: expected identity for a repository (binding, workspace, remote host alias, then active user), Git identity match, expected origin URL, plus a non-interactive `Inspect` report and `Fix`.
- `core/gitops`: Git command-line wrapper used only by the desktop app: status, diff, commit, discard, branches, history, fetch/pull/push, clone. Commands never prompt (SSH batch mode, no editor, no terminal prompts), have timeouts, and return categorised errors with plain-language messages. Commit and pull take the author identity explicitly.
- `core/uninstall`: undoes bgit's machine changes (repository remotes, managed SSH config, hooks path, Git identity, optional generated keys, config directory). `cmd/uninstall.go` prints around its steps; the desktop app calls `Run` and shows the summary.
- `core/ssh`: SSH key generation/validation, managed `~/.ssh/config` updates, SSH agent helpers, and GitHub SSH connectivity checks.
- Core APIs now prefer structured result objects from `core/models` over mixed tuple-style returns for mutations and operational checks.
- `cmd/export.go`: creates encrypted `.bgit` archives in the managed backup directory, prompting interactively for a password and confirmation before encrypting the unchanged inner archive payload.
- `cmd/import.go`: restores config from encrypted `.bgit` archives by prompting for the archive password, decrypting the payload, validating the archived config, and saving it atomically.
- `desktop/`: Wails desktop app with a separate entrypoint, a sidebar/navigation shell with dedicated Dashboard, Identities, Repository Bindings, Backup & Restore, and Doctor pages, backend methods for reading configured identities, active identity, effective identity state, desktop identity add/update/delete/activate operations, repository binding management through existing core repo modules, automatic identity detection/sync for selected repository or workspace paths, read-only diagnostic checks through existing core modules and config persistence, encrypted backup export/import through the existing core archive package, and app-native notifications/dialogs for user feedback and confirmations.
- `desktop/repos.go`, `desktop/gitclient.go`, `desktop/workspaces.go`, `desktop/settings.go`, `desktop/onboarding.go`, `desktop/system.go`: desktop backend by feature. Repository actions only accept repositories bgit manages (bound, or inside a workspace). Git failures are returned as results with a category so the UI can offer the next step (fix and push, pull first).
- `desktop/` entrypoint: started with arguments, the desktop binary runs the bgit CLI (`desktop/main.go`, `desktop/cli.go`). The managed pre-push hook written by the app calls `bgit` when it is on PATH and otherwise this binary, so app-only installs keep push protection.
- `desktop/frontend/dist/`: plain HTML/CSS/JS with no build step. `index.html` holds the shell and shared helpers; feature screens live in `js/*.js` and register themselves with `registerPage`. `app.css` holds their styles.
- `desktop/devharness.go` (build tag `devharness`): serves the frontend to a browser with the real backend on an isolated home directory, for developing and testing screens without WebKit/GTK.
- `internal/platform` background mode: the desktop app enables it so child processes do not open console windows on Windows.
- `cmd/setup.go`: initializes config, installs managed pre-push hook, stores previous hook path for later restore.
- `cmd/use.go`: switches active identity and stores the pre-bgit Git identity before the first managed switch.
- `cmd/uninstall.go`: restores repo remotes, removes managed SSH config, restores hooks/Git identity when possible, and removes bgit config.
- `internal/git`: global Git config wrappers.

## Packaging
- `.github/workflows/release.yml` builds the CLI binaries and the desktop app: an NSIS installer on Windows, a `.dmg` on macOS, and a `.deb` on Linux (`desktop/build/linux/package-deb.sh`).
- `desktop/build/` holds packaging inputs (icon, Linux desktop entry, version script). Wails writes its generated files and build output there too; those are ignored by Git.
- The Linux desktop entry must not pass arguments to the binary, because arguments switch it to CLI mode.
- Signing and notarization run only when the signing secrets are configured.

## Testing
- `make test`: runs `go test ./...` with an isolated Go build cache under `/tmp`.
- Go unit tests use `internal/testenv` to isolate HOME and Git global/system config in a temp directory, so they never touch real bgit, Git, or SSH state.
- `make test-desktop`: runs the desktop backend tests with the `devharness` build tag, which needs no WebKit/GTK. `internal/testenv.FakeGitHub` serves local bare repositories through Git's SSH transport so clone, pull, and push through identity host aliases run offline.
- `make test-ui`: runs the desktop UI tests in `desktop/uitest`. Headless Chrome drives the frontend against the dev harness on an isolated home, with GitHub replaced by local bare repositories behind a stand-in SSH command.
- `make test-integration`: runs real CLI integration tests on the host using isolated `HOME`/`XDG_CONFIG_HOME`, `GIT_CONFIG_NOSYSTEM=1`, disposable repositories, config-validation regression cases, and cleanup preservation via `BGIT_TEST_KEEP_TMP=1` when needed.
- `make test-docker`: builds `Dockerfile.test` and runs the same integration suite inside a disposable Linux container.
- `make test-backup-portability`: simulates Machine A and Machine B with separate temporary HOME directories, exports encrypted backups with SSH keys on Machine A, imports them on Machine B, and verifies config, key files, SSH config regeneration, and identity functionality.
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
