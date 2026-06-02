# Project Status

## Current State
bgit is a Go/Cobra CLI for managing multiple Git identities. The current implementation supports:

- identity add, list, use, update, delete
- repo bindings and workspace-based identity resolution
- managed SSH config entries and SSH key flows
- clone, remote fix/restore, status, active, prompt, doctor, sync, setup, and uninstall
- managed pre-push safety checks
- uninstall recovery for hooks, remotes, and Git identity
- automated validation with unit command coverage, integration tests, Docker tests, and guarded real-account acceptance tests

Code remains the source of truth for current behavior.

---

## Roadmap Alignment

Milestone reminder rule:
- Once a roadmap milestone is completed, update `roadmap.md`, update release notes/changelog, and release a new version.

---

## Completed

### R-001
- Status: done

Files Changed:
- core/config/config.go
- core/config/types.go
- internal/config/config.go
- internal/config/types.go
- cmd/add.go
- cmd/bind.go
- cmd/check.go
- cmd/clone.go
- cmd/delete.go
- cmd/doctor.go
- cmd/helpers.go
- cmd/init.go
- cmd/remote.go
- cmd/setup.go
- cmd/setup_ssh.go
- cmd/status.go
- cmd/sync.go
- cmd/uninstall.go
- cmd/update.go
- cmd/use.go
- cmd/workspace.go
- internal/identity/resolver.go
- internal/ssh/sshconfig.go
- internal/ui/output.go
- context.md
- project_status.md
- roadmap.md

What Changed:
- Extracted config ownership from `internal/config` into `core/config`.
- Moved config schema, config path helpers, persistence, legacy migration, and workspace/binding mutation logic into the new core package.
- Repointed commands and internal helpers to use `core/config` while preserving existing behavior.

Why:
- Aligns the codebase with the new roadmap structure and establishes the config core boundary before later SSH and identity extraction tasks.

---

### R-002
- Status: done

Files Changed:
- core/ssh/agent.go
- core/ssh/config.go
- core/ssh/connectivity.go
- core/ssh/keys.go
- internal/ssh/sshconfig.go
- internal/user/ssh.go
- cmd/add.go
- cmd/clone.go
- cmd/delete.go
- cmd/doctor.go
- cmd/setup.go
- cmd/setup_ssh.go
- cmd/sync.go
- cmd/uninstall.go
- cmd/update.go
- cmd/use.go
- context.md
- project_status.md
- roadmap.md

What Changed:
- Extracted SSH ownership from `internal/ssh` and `internal/user` into `core/ssh`.
- Moved SSH key generation, key validation, public key reading, managed SSH config updates, SSH agent helpers, and GitHub connectivity checks into the new core package.
- Repointed command flows to `core/ssh` while keeping existing CLI behavior intact.

Why:
- Aligns the codebase with the roadmap SSH boundary and removes the previous split ownership across unrelated internal packages and command-local helpers.

---

### R-003
- Status: done

Files Changed:
- core/identity/manage.go
- core/identity/resolver.go
- internal/identity/resolver.go
- cmd/active.go
- cmd/add.go
- cmd/bind.go
- cmd/check.go
- cmd/clone.go
- cmd/delete.go
- cmd/prompt.go
- cmd/remote.go
- cmd/status.go
- cmd/sync.go
- cmd/update.go
- cmd/use.go
- context.md
- project_status.md
- roadmap.md

What Changed:
- Extracted identity ownership into `core/identity`.
- Moved identity resolution out of `internal/identity` and added core add, update, delete, and activate operations there.
- Repointed commands to use the new core identity package while keeping prompts and user-facing output in the CLI layer.

Why:
- Aligns the codebase with the roadmap identity boundary and turns command handlers into thinner wrappers around shared identity logic.

---

### R-004
- Status: done

Files Changed:
- core/repo/repo.go
- cmd/bind.go
- cmd/check.go
- cmd/clone.go
- cmd/remote.go
- cmd/uninstall.go
- cmd/workspace.go
- context.md
- project_status.md
- roadmap.md

What Changed:
- Added `core/repo` for workspace registration/removal, repo binding/unbinding, clone auto-bind behavior, remote URL conversion, and repository owner resolution.
- Repointed the remaining repo-oriented CLI commands to call the extracted core modules instead of owning those behaviors directly.
- Kept prompts, confirmations, and display output in the CLI layer while moving business logic into core packages.

Why:
- Completes the CLI integration pass by making the command layer a thin wrapper around the extracted core modules.

---

### R-005
- Status: done

Files Changed:
- core/models/config.go
- core/models/identity.go
- core/models/repo.go
- core/models/results.go
- core/config/types.go
- core/identity/manage.go
- core/identity/resolver.go
- core/repo/repo.go
- core/ssh/agent.go
- core/ssh/connectivity.go
- context.md
- project_status.md
- roadmap.md

What Changed:
- Added `core/models` as the shared home for reusable identity, repository, and operational result structs.
- Moved shared structs out of package-local definitions and updated the core packages to reference the shared model layer through aliases.
- Kept persistence-specific `Config` in `core/config` while reusing shared entity structs for users, workspaces, and bindings.

Why:
- Establishes a shared model layer without changing behavior or starting the broader standardized response work planned for the next roadmap item.

---

### R-006
- Status: done

Files Changed:
- core/models/results.go
- core/identity/manage.go
- core/repo/repo.go
- core/ssh/agent.go
- core/ssh/connectivity.go
- cmd/add.go
- cmd/bind.go
- cmd/check.go
- cmd/clone.go
- cmd/doctor.go
- cmd/setup_ssh.go
- cmd/update.go
- cmd/workspace.go
- context.md
- project_status.md
- roadmap.md

What Changed:
- Standardized core mutation and operational APIs around structured result objects from `core/models`.
- Replaced mixed tuple-style returns in identity, repo, and SSH helpers with explicit result structs for user operations, workspace operations, binding removal, repo owner resolution, SSH command output, key loading, and connectivity checks.
- Updated the CLI layer to consume the standardized result objects without changing user-facing behavior.

Why:
- Makes the core layer more consistent and easier to consume before the later safety and export/import work.

---

### R-007
- Status: done

Files Changed:
- core/config/config.go
- core/config/validate.go
- context.md
- project_status.md
- roadmap.md

What Changed:
- Added centralized config validation for supported version, duplicate identities, required user fields, active user references, and workspace/binding user references.
- Normalized legacy config values before validation on both load and save.
- Replaced direct config file truncation with atomic temp-file writes and rename-based replacement.

Why:
- Makes config persistence safer and more resilient without changing the TOML format or user-facing behavior.

---

### R-008
- Status: done

Files Changed:
- tests/integration.sh
- context.md
- project_status.md
- roadmap.md

What Changed:
- Strengthened the host integration harness with tighter environment isolation, reusable command helpers, clearer failure output, and optional temp-dir preservation for debugging.
- Added regression coverage for invalid config version handling, duplicate user config rejection, binding override behavior, workspace removal, and no-op paths for bind/remote commands.
- Kept Docker and real-account flows compatible by improving the shared host integration path rather than introducing a separate framework.

Why:
- Improves repeatability and diagnostics for real CLI integration testing without changing product behavior.

---

### R-009
- Status: done

Files Changed:
- cmd/export.go
- core/export/archive.go
- core/export/export.go
- core/models/export.go
- tests/integration.sh
- context.md
- project_status.md
- roadmap.md

What Changed:
- Added `bgit export` as a focused archive-generation command that writes `.bgit` backup archives into the managed backup directory without adding import or encryption behavior.
- Added manifest models and a stable archive layout with `manifest.json`, `payload/config/config.toml`, and reserved payload directories so the archive shape is ready for the encryption layer planned in `R-010`.
- Added integration coverage that verifies the archive file is produced and contains the expected manifest and payload structure.

Why:
- Establishes the backup archive contract before encryption and import are layered on top of it.

---

### R-010
- Status: done

Files Changed:
- cmd/export.go
- core/export/archive.go
- core/export/encryption.go
- core/export/export.go
- core/models/export.go
- internal/ui/prompts.go
- tests/integration.sh
- context.md
- project_status.md
- roadmap.md

What Changed:
- Wrapped the existing `R-009` archive bytes in an encrypted outer `.bgit` file envelope while preserving the inner manifest and payload layout exactly.
- Added interactive password and confirmation prompts for `bgit export` without accepting passwords through CLI arguments.
- Added Argon2id key derivation metadata and AES-256-GCM encryption metadata to a small file header, followed by raw ciphertext for the payload.
- Updated integration coverage to verify encrypted export creation and confirm the file is no longer directly readable as a plaintext tar+gzip archive.

Why:
- Completes the export encryption layer while keeping the inner archive contract stable for the later import work in `R-011`.

---

### R-011
- Status: done

Files Changed:
- cmd/import.go
- core/export/archive.go
- core/export/encryption.go
- core/export/import.go
- core/models/export.go
- internal/ui/prompts.go
- tests/integration.sh
- context.md
- project_status.md
- roadmap.md

What Changed:
- Added `bgit import <archive.bgit>` with an interactive password prompt and no password CLI arguments.
- Added decrypt support for the existing `BGITEX10` envelope using stored Argon2id metadata and AES-256-GCM metadata.
- Added read support for the unchanged inner tar+gzip payload layout and restores `payload/config/config.toml` through validated atomic config saving.
- Added integration coverage for wrong-password failure and successful export/import restore after uninstall.

Why:
- Completes the encrypted export/import flow while preserving the archive format created by `R-009` and encrypted by `R-010`.

---

## Completed Outside Current Roadmap IDs

- uninstall recovery hardening
- post-uninstall read-only safety
- Docker-backed integration testing
- guarded real-account acceptance testing
- Windows install flow aligned to `install.ps1`
- removal of unused manual Windows installer packaging

---

## In Progress
None

---

## Pending
- R-012 through R-021 remain pending until implemented and recorded here with matching IDs.

---

## Blockers
None
