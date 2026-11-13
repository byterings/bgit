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
- desktop first-run setup wizard, repositories (clone, add existing, health, fix), workspaces, settings, and removal
- desktop everyday Git work: changes, diff, commit, discard, branches, history, fetch, pull, push
- desktop notifications, confirmation dialogs, and action status indicators
- desktop encrypted backup export/import with file selection, password prompts, and restore summaries
- desktop SSH public key display and copy action for GitHub setup
- desktop GitHub avatar display with initials fallback
- desktop release workflow artifacts for Linux, Windows, and macOS
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

Desktop redesign (R-033), which replaces the sidebar layout described in the older notes below:

- Navigation is a top bar: bgit mark, Dashboard, Repositories, Identities, Workspaces, Doctor, a repository search box, the push protection status, session activity, the default identity switcher, and Settings. There is no sidebar.
- Dark theme by default, with Light and System choices in Settings. The choice is stored in the app's local storage, not in the bgit configuration.
- An open repository has its own header (branch menu, sync status, Fetch, Pull, Push) and tabs: Changes (file list, diff with syntax colours, commit panel), History, Branches, Remotes, Identity.
- Identities is a list with a detail view (Git configuration, GitHub connection, SSH key, repositories). Switching the default identity in the top bar states that bound repositories and workspaces keep their identity.
- Doctor groups checks into areas and says why each area matters. Settings has General, Appearance, Push protection, Backup & Restore, and Advanced sections.
- Refinement (R-034): in the Changes tab the file list and commit area share one sidebar whose width (and the commit area's height) can be dragged; the sidebar collapses to a narrow rail (Ctrl/Cmd+B) that keeps the selected-file count and a commit button. The header shows "Commits as <identity> · Bound/Workspace/From remote/Default", which opens a card explaining why that identity applies and how it relates to the default identity; the top-bar switcher is labelled "Default". Tabs that do not fit the window move into a "More" menu.
- Not in the app because the backend has no support for them, and no placeholder controls were added: creating a new repository, merging branches from the app, adding or editing remotes, a commit graph, side-by-side diffs, staging individual lines, and editing an existing workspace.

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

Recent desktop packaging update:

- GitHub release workflow now builds and uploads desktop artifacts for Linux, Windows, and macOS alongside the existing CLI binaries.
- Release notes now use the same CLI installation commands as the site documentation and list the released desktop artifacts explicitly.
- README installation instructions now match the site installation steps for Linux, macOS, and Windows.

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

### R-020

- Status: done

Files Changed:

- .github/workflows/release.yml
- README.md
- project_status.md
- roadmap.md

What Changed:

- Split the release workflow into metadata, CLI build, Linux desktop build, Windows desktop build, macOS desktop build, and final release jobs.
- Added native desktop release artifact generation for Linux, Windows, and macOS using Wails on their corresponding GitHub Actions runners.
- Aggregated CLI and desktop artifacts into the final GitHub release so all desktop platforms are published from the same workflow.
- Updated generated release notes to use the same CLI installation steps as the site documentation and to list all desktop artifacts explicitly.
- Updated README installation instructions to match the documented Linux, macOS, and Windows install flow.

Why:

- Completes cross-platform desktop release packaging while keeping the CLI release behavior intact and aligning repo documentation with the published installation steps.

---

### R-022

- Status: done

Files Changed:

- core/setup/setup.go
- core/setup/setup_test.go
- core/check/check.go
- core/check/check_test.go
- core/models/setup.go
- internal/git/gitconfig.go
- internal/testenv/testenv.go
- cmd/setup.go
- cmd/check.go
- cmd/uninstall.go
- context.md
- project_status.md
- roadmap.md

What Changed:

- Moved config initialization, managed pre-push hook installation, and global hooks path handling from `cmd/setup.go` into `core/setup`.
- Moved the push safety decisions (expected identity, Git identity match, expected origin URL) from `cmd/check.go` into `core/check`, and added a non-interactive `Inspect` report and `Fix` for desktop use.
- The managed hook can now be generated with a fallback helper binary. Without a helper the hook is byte-for-byte the previous CLI hook.
- CLI commands keep their prompts and messages and call the shared packages for the logic.
- Added the first Go unit tests in the repository, running against an isolated HOME and Git config.

Why:

- The desktop app could not call setup or check logic while it lived in `cmd/`. Sharing it keeps the CLI and desktop app on one implementation.

Validation:

- `go test ./core/... ./internal/...` passes.
- `make test-integration` passes.
- A before/after comparison of CLI output for setup, check, bind, sync, and uninstall scenarios showed identical output.

---

### R-023

- Status: done

Files Changed:

- desktop/main.go
- desktop/cli.go
- desktop/console_windows.go
- desktop/console_other.go
- desktop/onboarding.go
- desktop/system.go
- desktop/devharness.go
- desktop/app.go
- desktop/backend.go
- desktop/models.go
- desktop/frontend/dist/index.html
- desktop/frontend/dist/app.css
- desktop/frontend/dist/js/onboarding.js
- desktop/frontend/dev.mjs
- core/setup/setup.go
- core/setup/setup_test.go
- core/ssh/agent.go
- core/ssh/keys.go
- internal/git/gitconfig.go
- internal/platform/proc.go
- internal/platform/proc_windows.go
- internal/platform/proc_other.go
- cmd/root.go
- .gitignore
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added a first-run wizard that checks for Git, runs setup, adds the first identity with a generated or existing SSH key, helps add the key to GitHub, and offers restore from a `.bgit` backup instead.
- The desktop binary now acts as the bgit CLI when started with arguments, and desktop setup writes a pre-push hook that falls back to the desktop binary when `bgit` is not on PATH. If neither exists the hook lets the push through with a notice instead of blocking every push.
- The app repairs the hook on startup and after a backup import (app moved or updated, or backup restored on a new machine). A hooks path the user set on purpose is left alone.
- The first identity added in the app becomes the active identity, and an existing bgit key for the same GitHub username is reused instead of failing.
- Desktop child processes no longer open console windows on Windows. The CLI is unaffected.
- Frontend feature code now lives in separate files under `desktop/frontend/dist/` and registers pages through a small registry in `index.html`.
- Added a dev harness (`go run -tags devharness ./desktop -home <dir>`) that serves the frontend to a browser with the real Go backend on an isolated home directory.
- Desktop Doctor hints now describe what to do in the app instead of CLI commands.

Why:

- Lets a user go from install to a working identity without the CLI, and keeps push protection working for app-only installs.

Validation:

- Go unit tests cover hook generation, fallback behavior, and hook repair.
- The wizard was run end to end in headless Chrome against the dev harness: setup, hook install, validation errors, key generation, GitHub step, and finish.
- CLI integration tests pass and CLI output is unchanged.
- Not yet verified: the built Wails app on Linux, Windows, or macOS. Windows console attachment for CLI mode is untested.

---

### R-027

- Status: done

Files Changed:

- core/ssh/connectivity.go
- core/ssh/connectivity_test.go
- core/models/results.go
- desktop/system.go
- desktop/app.go
- desktop/models.go
- desktop/frontend/dist/index.html
- desktop/frontend/dist/js/onboarding.js
- project_status.md
- roadmap.md

What Changed:

- Added a per-identity connection test that never prompts, reports wrong-account keys, and explains common failures in plain language.
- Added "Copy key and open GitHub" and "Test connection" to the identity details dialog and the wizard. Adding an identity now opens its details so the key can be added to GitHub straight away.
- External links are limited to a fixed list of HTTPS sites.

Why:

- Adding the key to GitHub was the one manual step with no guidance or confirmation.

Validation:

- Unit tests cover response classification. Both actions were exercised in the dev harness.
- Not verified against a real GitHub account: the test environment has no registered key, so only failure paths ran.

---

### R-030

- Status: done

Files Changed:

- core/gitops/run.go
- core/gitops/status.go
- core/gitops/diff.go
- core/gitops/commit.go
- core/gitops/branches.go
- core/gitops/log.go
- core/gitops/sync.go
- core/gitops/gitops_test.go
- core/models/git.go
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added `core/gitops`, a wrapper around the Git command line for status, per-file diff, commit (all changes or selected files), discard, branches, history, commit details, fetch, pull, push, clone, and abort merge.
- Every command runs without prompts: SSH runs in batch mode, editors and terminal prompts are disabled, and network commands have timeouts.
- Failures are returned with a category (auth, network, rejected, safety, conflict, dirty, identity, and others) and a plain-language message.
- Commits and pull merges take the author identity as an argument, so they do not depend on the Git config in effect.
- Paths and commit ids coming from the UI are validated so they cannot leave the repository or be read as Git options.

Why:

- The desktop app needs everyday Git operations with structured results. The CLI does not use this package.

Validation:

- `go test ./core/gitops/` covers each operation against real repositories and local bare remotes, including partial commits, discard, rejected push, merge on pull, and conflict then abort.

---

### R-024

- Status: done

Files Changed:

- desktop/repos.go
- desktop/app.go
- desktop/models.go
- desktop/backend_test.go
- desktop/frontend/dist/index.html
- desktop/frontend/dist/app.css
- desktop/frontend/dist/js/repositories.js
- internal/testenv/testenv.go
- Makefile
- context.md
- project_status.md
- roadmap.md

What Changed:

- Added a Repositories page that replaces the Repository Bindings page.
- Clone: takes a GitHub link (HTTPS, SSH, `github.com/owner/repo`, or `owner/repo`), an identity, and a folder; clones through the identity's SSH host alias, binds the clone, and sets its commit identity.
- Add existing: binds a local repository and, in the same step, sets its commit identity and rewrites its GitHub remote to the identity's SSH host alias. Changing a repository's identity uses the same flow.
- Removing a repository from bgit works even when its folder no longer exists.
- Added `make test-desktop` and an offline fake GitHub for tests (`internal/testenv`) that serves local bare repositories through Git's SSH transport.

Why:

- A repository cloned over HTTPS and only bound had its pushes blocked by the safety check until the remote was fixed by hand.

Validation:

- Desktop backend tests cover clone (including an identity without access and invalid input), add existing, and identity change.
- The page was driven end to end in headless Chrome against the dev harness.
- Not verified against real GitHub.

---

### R-025

- Status: done

Files Changed:

- desktop/workspaces.go
- desktop/backend.go
- desktop/app.go
- desktop/models.go
- desktop/backend_test.go
- desktop/frontend/dist/js/workspaces.js
- desktop/frontend/dist/app.css
- project_status.md
- roadmap.md

What Changed:

- Added a Workspaces page to list, add, and remove workspace folders, using the existing core workspace registration.
- Nested or duplicate workspaces are rejected with a reason. Removing a workspace never deletes the folder.
- Repositories found inside workspaces (up to three folder levels, skipping hidden and dependency folders) appear on the Repositories page.

Why:

- Workspaces were only manageable from the CLI.

Validation:

- Desktop backend tests and a headless Chrome run cover add, reject, list, and remove.

---

### R-026

- Status: done

Files Changed:

- desktop/repos.go
- desktop/frontend/dist/js/repositories.js
- desktop/frontend/dist/js/repository.js
- desktop/backend_test.go
- project_status.md
- roadmap.md

What Changed:

- Each repository shows a health state from the shared push safety check: Ready, Needs fix, Blocked, or Missing, with the reasons in plain language.
- Fix aligns one repository; Fix all aligns every fixable one.
- Repositories with a remote that is not on GitHub are shown as Blocked, because the existing push safety check rejects them.

Why:

- Gives the desktop app the equivalent of `bgit check` and `bgit sync --fix` across all repositories.

Validation:

- Desktop backend tests cover bound, workspace, non-GitHub, and missing repositories, and Fix all.

---

### R-028

- Status: done

Files Changed:

- core/uninstall/uninstall.go
- core/uninstall/exit.go
- core/uninstall/uninstall_test.go
- core/setup/setup.go
- core/setup/setup_test.go
- cmd/uninstall.go
- desktop/settings.go
- desktop/onboarding.go
- desktop/app.go
- desktop/models.go
- desktop/backend_test.go
- desktop/frontend/dist/index.html
- desktop/frontend/dist/js/settings.js
- desktop/frontend/dist/app.css
- context.md
- project_status.md
- roadmap.md

What Changed:

- Moved the uninstall steps (restore remotes, remove managed SSH config, restore hooks path and Git identity, remove generated keys) from `cmd/uninstall.go` into `core/uninstall`. The CLI keeps its prompts and messages.
- Added a Settings page with push protection status and repair, the setup guide, app information, and "Remove bgit from this computer" with an optional key deletion and a summary of what was restored.
- Turning push protection on from the app records a custom hooks folder it replaces, so removal restores it. The CLI keeps its record-once behavior.

Why:

- App-only users had no way to undo bgit's changes before uninstalling the app.

Validation:

- `core/uninstall` unit tests and desktop backend tests cover full removal, key handling, and legacy configs.
- CLI uninstall output was compared before and after across full and legacy scenarios and is identical.

---

### R-031

- Status: done

Files Changed:

- desktop/gitclient.go
- desktop/app.go
- desktop/models.go
- desktop/backend_test.go
- core/gitops/status.go
- core/gitops/commit.go
- core/gitops/gitops_test.go
- core/models/git.go
- desktop/frontend/dist/js/repository.js
- desktop/frontend/dist/app.css
- project_status.md
- roadmap.md

What Changed:

- Added a repository view with a Changes tab: changed files with checkboxes, a line-numbered diff, a commit box, and discard for one file or all.
- Commits use the repository's bgit identity explicitly, even if the repository's Git config drifted.
- Selecting a subset of files commits only those; everything else is left as it was.
- A conflicted file counts as resolved once its conflict markers are gone, so a merge can be finished from the app without staging by hand.

Why:

- Lets users commit with the right identity without a terminal or a second Git client.

Validation:

- Go tests cover full and partial commits, discard of every change kind, and committing a resolved merge.
- The view was driven end to end in headless Chrome.

---

### R-032

- Status: done

Files Changed:

- desktop/gitclient.go
- desktop/app.go
- desktop/models.go
- desktop/backend_test.go
- desktop/frontend/dist/js/repository.js
- desktop/frontend/dist/app.css
- project_status.md
- roadmap.md

What Changed:

- Added Fetch, Pull, and Push with incoming and outgoing commit counts. A branch that is not on the remote is published on first push.
- Push checks the repository against its identity first and offers "Fix and push". A rejected push offers "Pull now".
- A pull that stops on conflicts shows a merge banner with "Cancel merge".
- Added a branch picker (local and remote branches), new branch creation, and a History tab with commit details and per-file diffs.

Why:

- Completes everyday Git work in the app.

Validation:

- Desktop backend tests push through the managed pre-push hook and confirm a failing check is reported as a safety block.
- Headless Chrome runs cover push, publish, fetch, pull, rejected push, conflict, cancel merge, and fix and push, against a local stand-in for GitHub.
- Not verified against real GitHub or with passphrase-protected SSH keys.

---

### R-029

- Status: done (not yet run in CI)

Files Changed:

- .github/workflows/release.yml
- desktop/build/appicon.png
- desktop/build/set-product-version.mjs
- desktop/build/linux/package-deb.sh
- desktop/build/linux/bgit-desktop.desktop
- desktop/build/linux/bgit-desktop.png
- desktop/wails.json
- desktop/frontend/wailsjs/go/main/App.d.ts
- desktop/frontend/wailsjs/go/main/App.js
- desktop/frontend/wailsjs/go/models.ts
- .gitignore
- README.md
- context.md
- project_status.md
- roadmap.md

What Changed:

- Windows: the release workflow builds an NSIS installer (`bgit-desktop-windows-amd64-installer.exe`) next to the portable executable.
- macOS: the workflow builds the `bgit.app` bundle and publishes it as a `.dmg` with an Applications shortcut, plus a `.zip`. The previous macOS step passed `-nopackage` and then looked for an app bundle, which that flag does not produce.
- Linux: `desktop/build/linux/package-deb.sh` builds a `.deb` that installs the binary, a desktop entry, and an icon, and depends on GTK 3, WebKitGTK 4.1, Git, and the OpenSSH client.
- Code signing is optional and driven by repository secrets: `WINDOWS_CERTIFICATE` and `WINDOWS_CERTIFICATE_PASSWORD` for Windows; `MACOS_CERTIFICATE`, `MACOS_CERTIFICATE_PASSWORD`, `MACOS_SIGN_IDENTITY`, `APPLE_ID`, `APPLE_TEAM_ID`, and `APPLE_APP_PASSWORD` for macOS signing and notarization. Without them, unsigned artifacts are published.
- Desktop builds now carry the release version, and the app icon is generated from the existing bgit logo.
- Updated `softprops/action-gh-release` from v1 to v2 because the workflow linter reports v1 as too old to run.
- Regenerated the Wails frontend bindings for the new backend methods.

Why:

- Raw binaries are hard to install and launch for users who do not use a terminal.

Validation:

- The `.deb` script was run locally with a stand-in binary and the package metadata and contents were inspected.
- The workflow passes `actionlint`.
- Not verified: the workflow has not run on GitHub, so the Windows installer, the macOS disk image, and both signing paths are untested. No `.rpm` or AppImage is produced.

---

### R-033

- Status: done

Files Changed:

- desktop/frontend/dist/index.html
- desktop/frontend/dist/css/tokens.css
- desktop/frontend/dist/css/base.css
- desktop/frontend/dist/css/components.css
- desktop/frontend/dist/css/shell.css
- desktop/frontend/dist/css/pages.css
- desktop/frontend/dist/js/ui.js
- desktop/frontend/dist/js/core.js
- desktop/frontend/dist/js/shell.js
- desktop/frontend/dist/js/highlight.js
- desktop/frontend/dist/js/dashboard.js
- desktop/frontend/dist/js/repositories.js
- desktop/frontend/dist/js/repository.js
- desktop/frontend/dist/js/identities.js
- desktop/frontend/dist/js/workspaces.js
- desktop/frontend/dist/js/doctor.js
- desktop/frontend/dist/js/settings.js
- desktop/frontend/dist/js/onboarding.js
- desktop/frontend/dist/app.css (removed)
- desktop/frontend/wailsjs/go/main/App.d.ts
- desktop/frontend/wailsjs/go/main/App.js
- desktop/frontend/wailsjs/go/models.ts
- desktop/app.go
- desktop/backend.go
- desktop/gitclient.go
- desktop/models.go
- desktop/main.go
- core/gitops/remotes.go
- core/gitops/run.go
- core/gitops/gitops_test.go
- core/models/git.go
- desktop/uitest/lib.js
- desktop/uitest/client.test.js
- desktop/uitest/dashboard.test.js
- desktop/uitest/identities.test.js
- desktop/uitest/layout.test.js
- desktop/uitest/repos.test.js
- desktop/uitest/settings.test.js
- desktop/uitest/wizard.test.js
- changelog.json
- CHANGELOG.md
- README.md
- context.md
- project_status.md
- roadmap.md

What Changed:

- The desktop frontend was rewritten as one design system: colour, spacing, and type tokens in `css/tokens.css` (dark by default, light override), shared controls in `css/base.css` and `css/components.css`, and one file per screen under `js/`.
- Navigation moved from a sidebar to a top bar with repository search (Ctrl/Cmd+K), push protection status, session activity, the default identity switcher, and Settings.
- Repository workspace: header with branch menu and Fetch, Pull, Push; Changes tab with file list, syntax-coloured diff, and commit panel showing who the commit will be recorded as; History, Branches, Remotes, and Identity tabs.
- Identity manager with list and detail view; Workspaces as cards listing their repositories; Doctor grouped into areas with the reason each matters; Settings split into sections; setup guide with a step indicator. All earlier flows are kept: clone, add existing, fix, change identity, remove, backup export and restore, push protection repair, removal.
- Dialogs, menus, context menus, tooltips, notifications, empty states, and loading states are shared components. A dialog cannot be dismissed while its action is still running.
- Keyboard: Ctrl/Cmd+K search, Ctrl/Cmd+1-5 navigation, Ctrl/Cmd+, settings, Ctrl/Cmd+Enter commit, / to filter the repository list, Esc to close.
- Two small backend additions the new screens needed, both read from or delegated to Git: `ListRemotes` (Remotes tab) and `DeleteBranch` (Branches tab). Deleting a branch whose commits exist nowhere else is refused first and needs a second, explicit confirmation.
- Doctor messages from the desktop backend now read "2 identities configured" instead of "2 identity(s) configured". CLI output is unchanged.
- The window background is the dark theme colour, so the window does not flash white while loading.
- UI tests were rewritten for the new layout and extended: seven scenarios cover setup, identities, the repository list, the repository workspace (commit, discard, push, history, branches, remotes, fetch, pull, conflict, fix and push), dashboard and Doctor, settings (workspaces, push protection, backup and restore, removal), and layout (window sizes, themes, shortcuts, outside changes).

Why:

- The earlier screens worked but were built one at a time and did not look or behave like one product. The redesign makes the app usable without Git or CLI knowledge while keeping everything it could already do.

Validation:

- `make test`, `make test-desktop`, `make test-integration`, `make test-coexistence`, and `make test-ui` (7 scenarios) pass.
- `go vet` passes for Linux, Windows, and macOS targets of the desktop package.
- No file under `cmd/` or `internal/` changed, and the CLI does not import `core/gitops`, so CLI behaviour and output are unchanged.
- The built app was opened on a throwaway home on Linux (WebKitGTK 2.52) and the dashboard checked; the repository and identities screens were checked in the same web engine through the dev harness.
- Not verified: Windows (WebView2) and macOS (WKWebView) rendering, and clicking through the built app window by hand; interaction is covered by the headless Chrome tests against the real backend.

### R-034

- Status: done

Files Changed:

- desktop/frontend/dist/css/tokens.css
- desktop/frontend/dist/css/components.css
- desktop/frontend/dist/css/shell.css
- desktop/frontend/dist/css/pages.css
- desktop/frontend/dist/js/ui.js
- desktop/frontend/dist/js/core.js
- desktop/frontend/dist/js/shell.js
- desktop/frontend/dist/js/dashboard.js
- desktop/frontend/dist/js/repositories.js
- desktop/frontend/dist/js/repository.js
- desktop/frontend/dist/js/settings.js
- desktop/frontend/dist/js/workspaces.js
- desktop/uitest/lib.js
- desktop/uitest/layout.test.js
- README.md
- changelog.json
- CHANGELOG.md
- context.md
- project_status.md
- roadmap.md

What Changed:

- Repository workspace, Changes tab: the file list and the commit area are one sidebar, so the diff gets the rest of the width at every window size. A splitter sets the sidebar width and another sets the commit area height (drag, arrow keys, double-click to reset; remembered in local storage). The sidebar collapses to a rail (button or Ctrl/Cmd+B) that shows how many files are selected and opens the commit area at the summary.
- File list: arrow keys move the selection and Space includes or leaves out a file; a file left out of the commit is shown struck through and dimmed. The commit area is flat (no box inside a box), with "Commit to <branch>", who the commit is recorded as, summary with a length counter, description, and the commit button.
- Diff: the changed characters inside a replaced line are marked, long lines can be wrapped (remembered), the header steps to the previous or next file, binary files get their own message, and hunk headers and line numbers are easier to read. History has resizable commit list and file list columns.
- Header: the sync state is always visible as a coloured status (up to date, n to push, n to pull, not published, no remote); Fetch, Pull, and Push keep their counts and become icon buttons in narrow windows; Pull is outlined when commits are waiting.
- Identity: the header shows "Commits as <identity>" with where it comes from (Bound, Workspace, From remote, Default). Clicking it opens a card with the name and email on commits, the GitHub account, why this identity applies, whether the default identity applies here, any mismatch with a Fix button, and Change identity. The top-bar switcher is labelled "Default" and, inside a repository, marks that repository's identity. The Identity tab lists the default identity next to the repository's own. Workspaces explains the order bgit uses to pick an identity (bound repository, workspace, remote address, default), matching `core/check.ResolveExpected`.
- Navigation: tabs that do not fit move into a "More" menu, which is marked when the open page or a warning is inside it.
- Consistency: secondary and tertiary text meet 4.5:1 contrast on the app's surfaces in both themes; Settings and Workspaces use the full page width like the other screens; the backup explanation is an intro line instead of a separate card; the dashboard has a Health card (push protection and the last Doctor result).
- States: a failed repository list load shows the error with "Try again" on the dashboard and the Repositories page; a repository that cannot be read offers "Try again"; a clean working tree with unpushed commits offers Push; loading lists share one skeleton.

Why:

- The redesign's direction was kept; this pass makes the most-used screen give the diff more room, makes the identity that applies to a repository impossible to confuse with the default identity, and keeps every control usable in smaller windows.

Validation:

- `make test`, `make test-desktop`, `make test-integration`, `make test-coexistence`, and `make test-ui` (7 scenarios) pass.
- The layout scenario now checks 1366x768, 1280x720, 1024x768, and 960x640: the repository page fits without page scrolling, the commit button is in view, nothing scrolls sideways, no two controls in the top bar or repository header overlap, every main area is reachable, and every screen fits. It also checks the More menu at 860 wide, dragging and keyboard use of the splitters, collapsing the sidebar, keyboard file navigation, the wrap toggle, and the identity card.
- The same geometry was measured in WebKitGTK 2.52 (the Linux app's engine) at 1366x768, 1024x768, 960x640, and 860x600: no overlaps, no sideways scrolling, the commit button inside the window, and More only at 860.
- No backend or CLI file changed.
- Not verified: Windows (WebView2) and macOS (WKWebView) rendering.

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

### R-021 Release Stabilization

Done so far:

- Changelog entry for 0.12.0 added to `changelog.json` and `CHANGELOG.md` regenerated. `version.txt` is still 0.9.0, so nothing is released until it is changed.
- README describes the desktop app and how to install it.
- Review fixes: the desktop binary only enters CLI mode for a leading subcommand, so Wails' own flags and the macOS launch argument open the window; the app prefers its installed location on macOS when launched from a download; background refreshes no longer replace the commit message inputs.
- Desktop UI tests added under `desktop/uitest` (`make test-ui`): five scenarios, 79 checks, covering the wizard, identities, repositories, the repository view, workspaces, settings, and removal.
- Unused frontend code from the old Repository Bindings and Identity Detection panels was removed.
- `make desktop-harness` failed to serve the page because the frontend folder was resolved relative to the wrong directory; the harness now finds it from the repository root or from `desktop/`.
- `BGIT_DESKTOP_LOG=trace` makes the built app print every call between the window and the backend.

- Small screens (found on a 1366x768 laptop): the default window was taller than the screen, so the bottom of the app was out of reach. The window now opens maximised on small screens, the repository view fills the window with its own scrolling panes, dialogs never grow past the window, and spacing is tighter on short windows.
- Smoothness: a page now fades in only when the page changes (every re-draw used to replay the animation), page switches are immediate, scroll position is restored without a visible jump, and the top bar no longer blurs the page behind it. Lists re-draw only when their data changed.
- CLI and app together: opening the app on a machine set up with the CLI changes nothing (config, hook, Git config, and SSH config stay byte-for-byte the same). The app no longer turns push protection on by itself when Git has no hooks path; the dashboard shows that it is off instead. Returning to the window picks up changes made with the CLI or Git.
- `tests/coexistence.sh` (`make test-coexistence`) runs the real CLI binary and the desktop backend against one configuration in both install orders, including fix, commit, hook hand-over, and removal with either tool.
- Linux window now has the bgit icon and groups under the installed launcher. `BGIT_DESKTOP_GPU=on` opts in to GPU drawing; software drawing stays the default because GPU drawing leaves the window blank on some drivers.

Still to do before v1.0.0:

- Build and run the real Wails app on Windows and macOS, and click through every screen in the real window on Linux (the dashboard was captured from the real window; the repository view and lists were captured in the same web engine through the dev harness; nobody has clicked through the real window with a mouse yet). On Linux (Ubuntu 24.04, WebKitGTK 4.1) the app builds, opens, loads all frontend files, and its first backend calls succeed; CLI mode of the built binary (`status`, `setup`, `check --from-hook`) works. The screens themselves have only been driven in Chrome through the dev harness.
- Run the release workflow once and install each artifact (Windows installer, macOS disk image, Linux `.deb`).
- Test against real GitHub accounts: add key, test connection, clone, push, pull.
- Test Windows-specific code that cannot run here: console attachment in CLI mode, hidden child process windows, and the hook calling the installed `.exe`.
- Decide on code signing and add the signing secrets if wanted.
- Update the website docs for the desktop app (separate repository).

---

## Pending

- R-021 remaining items above.

Known limits of the desktop app (verified in code, 2026-11-12):

- Only GitHub remotes are managed. The shared push safety check rejects other hosts, so those repositories show as Blocked.
- Git operations never prompt. An SSH key with a passphrase must already be loaded in the SSH agent.
- Stash, rebase, tags, and resolving conflicts inside the app are not implemented. Conflicts are resolved in the user's editor.
- Setting the global `core.hooksPath` to bgit's hooks folder means Git ignores hooks inside individual repositories (for example `.git/hooks/pre-commit`). This is existing CLI behavior.

---

## Blockers

None
