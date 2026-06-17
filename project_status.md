# Project Status

## Current State

bgit is a Go/Cobra CLI for managing multiple Git identities. The current implementation supports:

- identity add, list, use, update, delete
- repo bindings and workspace-based identity resolution
- repo-local Git identity enforcement for bound repositories
- managed SSH config entries and SSH key flows
- clone, remote fix/restore, status, active, prompt, doctor, sync, setup, and uninstall
- managed pre-push safety checks
- uninstall recovery for hooks, remotes, and Git identity
- desktop notifications, confirmation dialogs, and action status indicators
- desktop encrypted backup export/import with file selection, password prompts, and restore summaries
- desktop SSH public key display and copy action for GitHub setup
- desktop GitHub avatar display with initials fallback
- automated validation with unit command coverage, comprehensive import/export integration tests, backup portability tests, Docker tests, and guarded real-account acceptance tests

Code remains the source of truth for current behavior.

Recent fix:

- Bound repositories now write and validate repo-local Git `user.name` and `user.email` for the bound identity.
- `bgit check` validates the effective Git identity for the repository, so a bound repo can pass even when the global active user differs, provided the repo-local Git identity matches the binding.
- `bgit sync --fix` repairs repo-local Git identity for bound repositories.
- `bgit use` now re-syncs repo-local Git identity for repositories already bound to the activated alias, so stale bound repos are repaired during normal CLI identity switches.
- Removing a binding now clears stale repo-local Git identity so the repository no longer keeps committing as the old bound user.
- Deleting an identity now automatically removes its dependent repo bindings and workspace bindings, and clears repo-local Git identity from the affected repositories.
- CLI confirmation prompts now accept non-TTY yes/no input, so scripted validation can cover delete, remote, check, and sync confirmation flows reliably.

Recent desktop update:

- Desktop identity actions now surface success, warning, and error feedback through app-native toast notifications.
- Identity deletion uses in-app confirmation dialogs instead of browser-native confirmation prompts.
- Long-running desktop actions show disabled/loading states and diagnostics show actionable error guidance.

Recent desktop backup update:

- Desktop users can create encrypted `.bgit` backups using the existing core export archive flow.
- Desktop users can import encrypted `.bgit` backups using the existing core import/restore flow.
- Import/export UI includes archive file selection, password prompts, unrecoverable-password warning, import replacement warning, and restore summaries.

Recent desktop SSH key update:

- Desktop identity rows show available SSH public key content with a copy action for GitHub setup.
- Desktop no longer relies on showing only the private key file path for generated/imported identities.
- Missing or unconfigured public keys are shown with clear status text.

Recent desktop avatar update:

- Desktop identity rows now show GitHub avatars derived from each configured GitHub username.
- Avatar rendering uses direct GitHub avatar URLs and initials fallback, without adding GitHub API calls or tokens.

Recent desktop repo binding update:

- Desktop users can view configured repository bindings.
- Desktop users can choose or enter a local Git repository path and bind it to an identity.
- Desktop repository binding uses the existing core repo binding flow and writes repo-local Git identity settings for the selected identity.
- Desktop users can change or remove existing repository bindings with in-app confirmation.
- Desktop activation and identity update flows now resync repo-local Git identity for repositories bound to the affected alias, so bound repositories keep using their bound commit identity even when the global active user differs.

Recent desktop activation fix:

- Global Git config commands now run from a stable existing directory so desktop identity activation does not fail when the app process current working directory is unavailable.

Recent desktop identity detection update:

- Desktop users can choose a repository or workspace path and see the effective bgit identity resolved from existing workspace, binding, and global identity rules.
- Desktop shows whether the relevant Git config currently matches the detected identity.
- Desktop users can sync the detected identity, using repo-local Git config for bound repositories and global Git config for workspace/global identities.

Recent desktop UI polish update:

- Desktop spacing, control sizing, table row alignment, and long-value truncation were tightened without changing navigation or business logic.
- Row actions now use compact icon buttons with accessible labels and tooltips for identity and repository binding actions.
- Panels, rows, controls, and toasts now use subtle transitions to reduce abrupt visual updates.
- Review fixes added one consistent inline SVG icon set, copy buttons for long values, stronger active identity highlighting, generic copied feedback, and scroll-position preservation during dashboard updates.

Recent desktop navigation update:

- Desktop UI is organized into a sticky sidebar shell with dedicated Dashboard, Identities, Repository Bindings, Backup & Restore, and Doctor pages.
- Active identity and quick health indicators are shown in a persistent top area while module content scrolls independently.
- Add Identity now opens in a modal instead of being permanently displayed on the page.
- Navigation refinements added page transition animation, cleaner page scroll behavior, and simplified identity cards with secondary SSH key details behind a compact disclosure.

Recent desktop UX polish update:

- Sidebar navigation now updates active state without rebuilding the full sidebar on every render.
- Page scroll position is preserved per section, while page switches restore the target section's prior scroll position.
- Topbar rendering is cached to avoid unnecessary DOM replacement during routine data refreshes.
- Dashboard now emphasizes the active identity and includes recent activity for identity, backup, repository, and doctor actions.
- Doctor results default to summary-first collapsible sections so long pass lists do not dominate the page.
- Button loading states preserve icon markup and stable button sizing to reduce layout jumps.

Recent desktop UI refinement update:

- Dashboard header was simplified to show only the active identity avatar, alias, and email as the persistent header item.
- The header now includes a circular back action using desktop page navigation history, with active identity alias and email aligned from the same left edge.
- The back action is positioned at the far left of the header while active identity information remains aligned at the far right.
- The Dashboard hides the back action because it is the root desktop page.
- Dashboard duplicate Active Profile card was removed and Quick Actions now use an equal 2x2 launchpad layout.
- Recent Activity now shows only recorded desktop actions, limited to the latest five entries, without generated placeholder timestamps.
- Identities now render as full-width horizontal records with avatar, alias, name, email, GitHub username, SSH key copy action, status, and icon-only actions.
- Identity records now display the SSH public key directly instead of its filesystem path, use stable aligned columns, and avoid duplicating the active badge in the status column.
- A read-only identity details modal displays and copies the SSH public key instead of exposing its filesystem path.
- Successful identity creation now prompts the user to copy the generated public key and add it to GitHub SSH keys.
- Doctor now exposes a single page-level Run Checks action and no longer duplicates aggregate pass, warning, and error badges inside the results panel.
- Desktop responsiveness now includes tablet and mobile layouts for horizontal navigation, identity records, forms, dialogs, diagnostics, long values, action controls, and notifications.
- Desktop chooser controls are now attached to their related fields for repository binding, identity detection, backup export/import, and SSH private-key selection.
- Add Identity now prevents conflicting SSH key choices by making generated keys and existing private-key paths mutually exclusive.
- Repository binding edit mode now shows an explicit editing state with Update Binding and Cancel actions.
- Backup, binding, detection, and edit/save actions now use consistent primary action styling and explicit submit flows.

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

### R-011A

- Status: done

Files Changed:

- core/export/archive.go
- core/export/export.go
- core/export/import.go
- core/models/export.go
- tests/backup-portability.sh
- Makefile
- context.md
- project_status.md
- roadmap.md

What Changed:

- Extended encrypted `.bgit` backups to include configured identity SSH private and public keys inside the encrypted payload under `payload/keys/<alias>` and `payload/keys/<alias>.pub`.
- Import now restores archived key pairs to the target machine under `~/.ssh/bgit_<alias>`, secures private keys with `0600`, rewrites imported `ssh_key_path` values, saves the validated config, and regenerates managed SSH config entries.
- Added `make test-backup-portability` to simulate Machine A export and Machine B import with isolated HOME directories and validate identities, active user, config data, SSH key files, SSH config entries, and target-machine key paths.

Why:

- Makes encrypted backups portable across machines without regenerating SSH keys.

---

### R-012

- Status: done

Files Changed:

- desktop/main.go
- desktop/app.go
- desktop/backend.go
- desktop/models.go
- desktop/frontend/dist/index.html
- desktop/frontend/package.json
- desktop/frontend/build.mjs
- desktop/frontend/dev.mjs
- wails.json
- Makefile
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added a Wails desktop foundation with a separate `desktop` entrypoint so the existing CLI entrypoint remains unchanged.
- Added minimal backend methods for desktop consumers to read configured identities, active identity, effective identity, and summary counts through existing core modules.
- Added a minimal embedded frontend and Makefile targets for desktop dev/build workflows.

Why:

- Establishes the desktop application structure and backend integration required before building the dashboard and management UI in later roadmap tasks.

---

### R-013

- Status: done

Files Changed:

- desktop/frontend/dist/index.html
- project_status.md
- roadmap.md

What Changed:

- Replaced the desktop placeholder screen with a read-only identity dashboard.
- Added active profile, identity counts, setup state, identity table, SSH key status badges, effective identity source, refresh behavior, and empty/unconfigured states.
- Kept all identity mutation behavior out of scope for `R-014`.

Why:

- Provides the first usable desktop screen for viewing configured identities and the active profile.

---

### R-014

- Status: done

Files Changed:

- desktop/app.go
- desktop/backend.go
- desktop/models.go
- desktop/frontend/dist/index.html
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added Wails backend methods for desktop identity add, update, activate, and delete operations.
- Added request/result models for desktop identity mutations and refreshed dashboard state after successful operations.
- Extended the desktop UI with add identity, edit identity, activate identity, and delete identity controls.
- Kept identity mutations backed by existing config, SSH, and core identity behavior while leaving CLI behavior unchanged.

Why:

- Completes desktop identity management so users can manage configured bgit identities without leaving the desktop app.

---

### R-015

- Status: done

Files Changed:

- desktop/app.go
- desktop/backend.go
- desktop/models.go
- desktop/frontend/dist/index.html
- desktop/frontend/wailsjs/go/main/App.d.ts
- desktop/frontend/wailsjs/go/main/App.js
- desktop/frontend/wailsjs/go/models.ts
- .gitignore
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added a desktop `GetDoctorStatus` backend method with read-only diagnostic sections for config, SSH setup, SSH agent, and Git identity alignment.
- Added desktop doctor models for sectioned pass/warn/fail checks with suggested fixes.
- Extended the desktop UI with a visual Doctor panel, summary counts, sectioned check rows, and refresh behavior.
- Added a narrow gitignore exception so the embedded desktop HTML is tracked.
- Kept existing CLI `bgit doctor` behavior unchanged.

Why:

- Completes visual diagnostics in the desktop app without adding desktop auto-fix behavior or changing CLI doctor output.

---

### R-016

- Status: done

Files Changed:

- desktop/frontend/dist/index.html
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added app-native toast notifications for desktop success, warning, and error feedback.
- Replaced browser-native confirmation prompts with in-app dialogs for destructive identity actions.
- Added disabled/loading states for long-running desktop actions and actionable error guidance for common backend/config/SSH failures.

Why:

- Makes desktop feedback easier to understand and closer to bgit's CLI clarity without changing CLI behavior.

---

### R-016A

- Status: done

Files Changed:

- desktop/app.go
- desktop/backend.go
- desktop/models.go
- desktop/frontend/dist/index.html
- desktop/frontend/wailsjs/go/main/App.d.ts
- desktop/frontend/wailsjs/go/main/App.js
- desktop/frontend/wailsjs/go/models.ts
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added desktop backup export and import backend methods backed by the existing encrypted core archive package.
- Added Wails file dialogs for choosing export destinations and import archives.
- Added desktop request/result models for backup export/import status, restored counts, active user, refreshed dashboard state, and diagnostics.
- Added a Backup & Restore desktop panel with password confirmation, import password entry, password safety warning, import replacement warning, archive path display, and restore summary output.

Why:

- Brings the existing CLI backup/restore capability into the desktop app while preserving the established encrypted archive format and core import/export behavior.

---

### R-016B

- Status: done

Files Changed:

- desktop/backend.go
- desktop/models.go
- desktop/frontend/dist/index.html
- desktop/frontend/wailsjs/go/models.ts
- project_status.md
- roadmap.md

What Changed:

- Added SSH public key fields to desktop identity views.
- Loaded public key content from each configured identity's `.pub` file without exposing private key content.
- Replaced the identity table's SSH file path display with a public key preview, GitHub-oriented copy action, and missing/unconfigured key states.

Why:

- Lets users copy the generated/imported SSH public key directly from the desktop app and add it to GitHub without manually opening key files.

---

### R-017

- Status: done

Files Changed:

- desktop/backend.go
- desktop/models.go
- desktop/frontend/dist/index.html
- desktop/frontend/wailsjs/go/models.ts
- project_status.md
- roadmap.md

What Changed:

- Added GitHub avatar URLs to desktop identity views based on configured GitHub usernames.
- Rendered avatars in desktop identity rows with lazy-loaded GitHub image URLs.
- Added initials fallback for missing usernames or image load failures.

Why:

- Makes desktop identity cards easier to scan visually while avoiding GitHub API tokens or blocking backend network calls.

---

### R-018

- Status: done

Files Changed:

- desktop/app.go
- desktop/backend.go
- desktop/models.go
- desktop/frontend/dist/index.html
- desktop/frontend/wailsjs/go/main/App.d.ts
- desktop/frontend/wailsjs/go/main/App.js
- desktop/frontend/wailsjs/go/models.ts
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added desktop repository binding models, status rows, and action results.
- Added Wails backend methods to choose a repository directory, bind a repository to an identity, and remove a repository binding.
- Reused the existing core repo binding implementation and refreshed repo-local Git `user.name` and `user.email` when binding through the desktop app.
- Added a Repository Bindings desktop panel with path entry, directory picker, identity selector, binding list, change action, and remove confirmation.

Why:

- Completes desktop repository-to-identity mapping while preserving existing CLI behavior and core repo ownership.

---

### R-019

- Status: done

Files Changed:

- desktop/app.go
- desktop/backend.go
- desktop/models.go
- desktop/frontend/dist/index.html
- desktop/frontend/wailsjs/go/main/App.d.ts
- desktop/frontend/wailsjs/go/main/App.js
- desktop/frontend/wailsjs/go/models.ts
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added desktop identity detection models and Wails backend methods for choosing a repository/workspace path, detecting the effective identity, and syncing the detected identity.
- Reused existing core identity resolution rules instead of changing CLI/core resolution behavior.
- Added an Identity Detection desktop panel that shows detected source, resolved identity, repository path, Git config scope, Git config values, and match status.
- Added a desktop Sync Identity action that fixes repo-local Git identity for bound repositories and global Git identity for workspace/global resolution.

Why:

- Completes automatic identity switching visibility and correction in the desktop app while preserving existing CLI behavior.

Milestone:

- R-018 and R-019 are complete. R-019-A was added as a final UI polish task before the v0.9.0 release.

---

### R-019-A

- Status: done

Files Changed:

- desktop/frontend/dist/index.html
- project_status.md
- roadmap.md

What Changed:

- Standardized desktop control sizing, spacing, panel radius usage, table alignment, and status badge sizing.
- Added subtle transitions for controls, panels, table rows, and toast entry to make refreshes and state changes feel less abrupt.
- Replaced identity and repository binding row text actions with a single inline SVG icon set using consistent sizing, tooltips, hover states, and accessible labels.
- Added copy buttons, copied tooltip state, and toast feedback for copyable long values including public keys, repository paths, backup paths, and detection paths.
- Added stronger active identity highlighting in the dashboard and identity list with a green indicator, active badge, and highlighted row.
- Added truncation and title tooltips for long identity values, repository paths, backup paths, detection paths, and public key values.
- Preserved scroll position across dashboard re-renders to reduce abrupt update jumps.
- Improved table row spacing and hover states while preserving existing section order and page structure.
- Improved Doctor section spacing and visual grouping without changing diagnostics behavior.

Why:

- Improves desktop readability and interaction quality without changing business logic, navigation, or backend behavior.

Milestone:

- R-018, R-019, and R-019-A are complete. R-019-B was added as a desktop navigation redesign task before the v0.9.0 release.

---

### R-019-B

- Status: done

Files Changed:

- desktop/frontend/dist/index.html
- project_status.md
- roadmap.md

What Changed:

- Reorganized the desktop app from one long management page into a structured shell with sticky sidebar navigation and independently scrolling content.
- Added dedicated pages for Dashboard, Identities, Repository Bindings, Backup & Restore, and Doctor while reusing existing frontend actions and backend calls.
- Added a persistent top area that shows active identity details, identity/binding counts, doctor health status, and refresh action.
- Moved Add Identity into a modal/popup form instead of showing it permanently.
- Added dashboard quick actions and kept existing identity detection, repository binding, backup/import, and doctor functionality available through dedicated pages.
- Refined page switching with smooth content transitions, cleaner scroll reset on navigation, and preserved scroll behavior for data refreshes.
- Reduced identity card density by showing core identity details first and moving SSH public key details into a compact expandable section.

Why:

- Improves desktop application structure and navigation while preserving existing functionality, import/export behavior, identity handling, and core logic.

Milestone:

- R-018, R-019, R-019-A, and R-019-B are now complete, so the v0.9.0 milestone has been reached. Update roadmap/release notes/changelog/version and release a new version before moving to the next milestone.

---

## Completed Outside Current Roadmap IDs

- uninstall recovery hardening
- post-uninstall read-only safety
- Docker-backed integration testing
- guarded real-account acceptance testing
- Windows install flow aligned to `install.ps1`
- removal of unused manual Windows installer packaging
- comprehensive BGIT import/export automated coverage for encrypted export, import restore, wrong-password failure, corrupted/empty archive rejection, isolated temporary environment use, SSH config regeneration, and config round-trip validation
  tested on local machine A and remote machine B
- R-019-C desktop UX polish for smoother navigation, per-page scroll preservation, recent activity, summary-first Doctor details, cached topbar/sidebar updates, and stable action loading states

---

## In Progress

None

---

## Pending

- R-020 through R-021 remain pending until implemented and recorded here with matching IDs.

---

## Blockers

None
