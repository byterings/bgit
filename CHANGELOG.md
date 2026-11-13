# Changelog

All notable changes to bgit will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.12.0] - 2026-11-13 (Phase 10)

The desktop app now works on its own, with a redesigned interface: first-run setup, repositories with clone and one-click fixes, workspaces, everyday Git work, and installers for Windows, macOS, and Linux. The CLI is unchanged.

### Added
- **Redesigned desktop interface** - The desktop app has a new look built as one design system: top navigation with repository search, a dark theme by default with a light option, a three-part repository view (files, diff, commit) with History, Branches, Remotes, and Identity tabs, an identity manager, a diagnostic Doctor, and keyboard shortcuts
- **Desktop interface refinements** - The repository view gives the diff more room with a resizable, collapsible file list and commit area, marks the changed part of each line, and always shows the sync state; the header says which identity commits in the repository and why; navigation fits smaller windows with a More menu
- **Desktop first-run setup** - A setup guide checks for Git, turns on push protection, adds the first identity with an SSH key, and helps add the key to GitHub, or restores everything from a .bgit backup
- **Desktop repositories** - A Repositories page clones GitHub repositories with a chosen identity, adds existing local repositories, shows whether each one matches its identity, and fixes mismatches with one click
- **Desktop Git workflow** - Each repository has a view for changed files and diffs, commit, discard, branches, history, and fetch, pull, and push, always using the repository's bgit identity
- **CLI and desktop app share one setup** - The desktop app reads and writes the same configuration as the CLI. Opening it on a machine already set up with the CLI changes nothing, and changes made with either tool show up in the other
- **Desktop workspaces** - Workspace folders can be added and removed in the app, and repositories found inside them are listed with the workspace identity
- **GitHub key helper and connection test** - Identity details can copy the SSH key, open the GitHub key settings page, and test that GitHub accepts the key for the expected account
- **Desktop settings and removal** - A Settings page shows and repairs push protection and can remove bgit from the computer, restoring repository remotes, SSH config, Git hooks, and the previous Git identity
- **Desktop installers** - Releases include a Windows installer, a macOS disk image, and a Debian package for the desktop app, with optional code signing
- **Remotes and branch deletion in the desktop app** - A repository shows its remotes and whether origin is routed through the right identity, and branches can be deleted, with a second confirmation before deleting a branch whose commits exist nowhere else

### Fixed
- **Push protection without the CLI** - The desktop app can answer the pre-push check itself, so push protection works when only the desktop app is installed. If neither bgit nor the app can be found, the hook lets the push through with a notice instead of blocking it
- **Hook repair after restore or app move** - The desktop app reinstalls the push check after a backup is restored on a new machine or after the app is moved or updated
- **Remote alignment when adding a repository** - Adding or rebinding a repository in the desktop app now also rewrites its GitHub remote to the identity's SSH key, so the next push is not blocked
- **No console windows on Windows** - Git and SSH commands started by the desktop app no longer open console windows
- **macOS desktop release bundle** - The release workflow now builds the macOS app bundle; the previous step skipped packaging and could not produce it
- **Desktop window on small screens** - The window opens maximised on small screens such as 1366x768, the repository view fills the window so its buttons stay reachable, and re-drawing a page no longer replays animations

## [0.9.0] - 2026-06-17 (Phase 9)

Milestone release completing desktop repository mapping, identity detection, UI polish, structured navigation, and repository-local Git identity cleanup across CLI and desktop flows.

### Added
- **Desktop repository bindings** - Added desktop repository binding management with repository selection, identity assignment, binding updates, and removal actions
- **Desktop identity detection and sync** - Added desktop identity detection for repository and workspace paths, showing the effective bgit identity and allowing Git config sync from the UI
- **Desktop interaction polish** - Improved desktop spacing, alignment, copy actions, truncation, active-state highlighting, hover states, and smoother refresh behavior without changing backend logic
- **Desktop structured navigation** - Reorganized the desktop app into dedicated Dashboard, Identities, Repository Bindings, Backup & Restore, and Doctor pages with sticky sidebar navigation

### Fixed
- **Repo-local Git identity consistency** - Bound repositories are now re-synced on identity activation and update, binding removal clears stale local Git identity, and deleting an identity removes dependent bindings/workspaces and clears affected repositories
- **Scriptable confirmation flows** - CLI yes/no confirmation prompts now accept non-TTY input so automated validation and scripted flows work reliably
- **Desktop binding cleanup parity** - Desktop binding removal and identity deletion now follow the same repository-local Git cleanup behavior as the CLI
- **Desktop identity workflow clarity** - Desktop add/view flows now show SSH public keys directly and guide users to add generated keys to GitHub
- **Desktop doctor and responsive cleanup** - Doctor page duplication was removed and desktop layouts were tightened for smaller window sizes

## [0.8.0] - 2026-06-10 (Phase 8)

Milestone release expanding the desktop app with native feedback, encrypted backup and restore, SSH public key copy support, and GitHub avatar identity cards.

### Added
- **Desktop notification system** - Desktop identity actions now use app-native success, warning, and error notifications, confirmation dialogs, and action loading states
- **Desktop backup and restore** - Added desktop UI for encrypted .bgit export and import using the existing core archive format, file selection dialogs, password prompts, and restore summaries
- **Desktop SSH public key copy** - Identity rows now show SSH public key previews with a copy action so users can add generated keys to GitHub without manually opening key files
- **GitHub avatar identity cards** - Desktop identity rows now show GitHub avatars derived from configured GitHub usernames with initials fallback

### Fixed
- **Clearer desktop destructive actions** - Identity deletion and backup import now use explicit in-app confirmation dialogs with safety warnings
- **Actionable desktop error feedback** - Common backend, config, SSH, and identity failures now surface clearer user-facing hints in the desktop app

## [0.7.2] - 2026-06-10 (Phase 7)

Patch release fixing bound repository Git identity enforcement so normal git commits use the repository's bound bgit identity.

### Fixed
- **Repo-local Git identity for bound repositories** - bgit bind now writes repository-local user.name and user.email for the bound identity so normal git commit uses the correct author even when the global active identity differs
- **Bound repository safety checks** - bgit check now validates the effective repository Git identity and allows a bound repository to pass when its local Git identity matches the binding
- **Sync repair for bound repositories** - bgit sync --fix now repairs repo-local Git identity for bound repositories instead of requiring a global identity switch

## [0.7.1] - 2026-06-10 (Phase 7)

Release packaging update for the Phase 7 desktop beta, adding Linux desktop artifact generation to the GitHub release workflow.

### Added
- **Linux desktop release artifact** - The GitHub release workflow now builds the Wails desktop application on Linux and uploads bgit-desktop-linux-amd64 alongside the CLI binaries

### Fixed
- **Desktop asset release coverage** - Ensures the desktop beta binary is attached to the GitHub release instead of shipping only CLI binaries
- **Release workflow dependencies** - Installs the required Linux GTK, WebKit, Soup, Node, and Wails dependencies before building the desktop artifact

## [0.7.0-beta] - 2026-06-05 (Phase 7)

Beta milestone release introducing the Wails desktop app foundation, identity dashboard, identity management, and visual diagnostics.

### Added
- **Desktop app foundation** - Added a Wails desktop application under desktop/ with a separate entrypoint and backend bindings while preserving the existing CLI entrypoint
- **Desktop identity dashboard** - Added a desktop dashboard for configured identities, active profile, setup state, identity counts, SSH key status, and effective identity source
- **Desktop identity management** - Added desktop actions to add, edit, activate, and delete bgit identities using the existing core identity and SSH flows
- **Desktop doctor integration** - Added read-only desktop diagnostics for config health, SSH setup, SSH agent state, and Git identity alignment

### Fixed
- **Desktop project structure cleanup** - Moved active Wails configuration under desktop/ and removed obsolete manual Windows installer documentation from the release structure
- **Beta release metadata** - Prepared release metadata for v0.7.0-beta and marks beta releases as prereleases in the GitHub release workflow

## [0.6.0] - 2026-06-02 (Phase 6)

Milestone release adding encrypted bgit export and import archives for portable identity backups.

### Added
- **BGIT export archive system** - Added bgit export to package the current bgit backup payload into a stable .bgit archive structure
- **Encrypted export layer** - Wrapped .bgit archives in an encrypted envelope using Argon2id key derivation and AES-256-GCM payload encryption
- **BGIT import restore flow** - Added bgit import to decrypt encrypted .bgit archives, validate the archived config, and restore it atomically (`bgit import <archive.bgit>`)
- **Portable SSH key backup support** - Encrypted .bgit backups now include configured identity SSH private and public keys under payload/keys so backups can be imported on another machine without regenerating keys
- **Backup portability validation** - Added a Machine A to Machine B portability test that verifies identities, active user, SSH key files, target-machine key paths, and regenerated SSH config entries (`make test-backup-portability`)

### Fixed
- **Password-only interactive archive protection** - Export and import passwords are prompted interactively and are not accepted through command-line arguments
- **Export password recovery warning** - bgit export now warns users that the archive password is required for import and cannot be recovered if forgotten
- **Target-machine SSH key path restore** - Import rewrites restored SSH key paths to the target machine and regenerates bgit-managed SSH config entries after restoring keys
- **Encrypted archive integration coverage** - Integration tests now cover encrypted export creation, unreadable plaintext archive checks, wrong-password import failure, corrupted and empty archive rejection, and successful import restore

## [0.5.0] - 2026-06-02 (Phase 5)

Milestone release focused on shared core models, standardized core responses, safer config persistence, and stronger isolated integration testing.

### Added
- **Shared core model layer** - Added core/models to hold reusable identity, repository, and operational structs shared across config, identity, repo, and SSH modules
- **Standardized core result objects** - Core identity, repo, and SSH operations now return structured result objects instead of mixed tuple-style responses
- **Safer config persistence** - Config loads and saves now validate supported versions and references, normalize legacy values, and save atomically via temp-file replacement
- **Stronger isolated integration harness** - Integration tests now run with isolated HOME and XDG config paths, explicit Git system-config isolation, clearer failure output, and optional temp-dir preservation (`make test-integration`)

### Fixed
- **Config validation regressions caught earlier** - Integration coverage now includes invalid config version handling, duplicate user rejection, workspace removal, and binding override/no-op paths
- **More consistent core APIs** - Core consumers now work against predictable result shapes, reducing command-layer branching and special-case handling

## [0.4.0] - 2026-05-29 (Phase 4)

Milestone release that finishes the core extraction and CLI integration work for config, SSH, identity, and repository flows.

### Added
- **Core config module** - Moved config loading, persistence, migration, and active identity state into core/config as the shared config boundary
- **Core SSH module** - Moved SSH key generation, SSH config management, agent helpers, and connectivity checks into core/ssh
- **Core identity module** - Moved identity resolution and identity lifecycle operations into core/identity
- **Core repository module** - Added core/repo for workspace registration, repository binding, remote URL conversion, clone auto-bind support, and repository owner resolution

### Fixed
- **Thinner CLI command layer** - Commands now route their core behavior through extracted modules instead of duplicating repo, identity, and remote logic directly
- **Consistent repository safety decisions** - Repository owner detection, remote conversion, and clone binding now use shared core logic across check, remote, clone, bind, and uninstall flows

## [0.3.1] - 2026-05-25 (Phase 3)

Bugfix release focused on uninstall recovery, post-uninstall safety, and repeatable integration testing.

### Added
- **Docker-backed integration test harness** - Added repeatable host and Docker integration tests covering setup, identity flows, remote fixes, safety checks, and uninstall cleanup (`make test-docker`)
- **Real-account acceptance test harness** - Added an opt-in real-account test flow that backs up and restores real bgit, SSH, and Git state while validating selected identities and repositories (`make test-real`)

### Fixed
- **Uninstall recovery for existing users** - Resolved GitHub issue #5 reported by maheshmthorat: uninstall now restores configured repo remotes first, removes managed SSH config, restores or clears bgit-managed hooks, and restores backed-up Git identity when available so GitHub Desktop and normal Git commands are not left affected after uninstall
- **Post-uninstall read-only safety** - Read-only commands such as list, status, active, prompt, and doctor no longer recreate ~/.bgit or reinstall hooks after uninstall
- **Optional SSH key cleanup** - Added bgit uninstall --remove-keys so SSH key deletion is explicit instead of implicit

## [0.3.0] - 2026-03-04 (Phase 3)

Phase 3 usability and safety release focused on one-time setup and automatic push protection.

### Added
- **One-time setup command** - Added bgit setup to initialize configuration, install global pre-push safety checks, and prepare defaults (`bgit setup`)
- **Automatic push safety checks** - Added bgit check and managed pre-push hook integration to validate identity and remote alignment (`bgit check`)
- **Shell prompt integration** - Added prompt-friendly identity output for shell integrations (`bgit prompt --plain`)
- **Clone auto-bind by default** - Repositories cloned with bgit clone are automatically bound to the effective identity by default (`bgit clone <url> [directory]`)

### Fixed
- **Consistent first-run behavior** - Commands now trigger setup flow automatically when bgit has not been initialized
- **Uninstall cleanup** - Uninstall now clears managed global hooks path and removes managed SSH config entries

### Breaking Changes
- **Legacy command deprecations** - bgit init and bgit setup-ssh are deprecated in favor of bgit setup; bgit remote fix is now legacy/advanced guidance

## [0.2.1] - 2025-01-28 (Phase 2)

Code cleanup and improved installation experience.

### Added
- **Dynamic version fetching** - Install scripts now automatically fetch the latest version from GitHub

### Fixed
- **Standardized user prompts** - All confirmation prompts now use consistent UI patterns
- **Code cleanup** - Removed unnecessary comments and improved code readability

## [0.2.0] - 2025-01-28 (Phase 2)

Phase 2 release with workspace support and diagnostics.

### Added
- **Workspace management** - Create organized workspace folders with automatic identity binding (`bgit workspace`)
- **Repository binding** - Bind individual repositories to specific identities (`bgit bind`)
- **Identity status** - Show current identity status and bindings (`bgit status`)
- **Diagnostics** - Diagnose and auto-fix configuration issues (`bgit doctor`)
- **Active identity** - Show current active identity (`bgit active`)
- **Identity resolution** - Automatic identity resolution based on workspace > binding > global priority

## [0.1.0] - 2025-01-28 (Phase 1)

Initial release with core identity management features.

### Added
- **Multi-user identity management** - Add, list, use, and delete Git identities (`bgit add/use/list/delete`)
- **SSH key management** - Generate or import SSH keys for each identity (`bgit add`)
- **Git config handling** - Automatic user.name and user.email configuration (`bgit use`)
- **Clone with identity** - Clone repositories with the correct SSH configuration (`bgit clone`)
- **Remote management** - Fix and restore remote URLs for bgit compatibility (`bgit remote fix/restore`)
- **Configuration sync** - Validate and sync bgit configuration (`bgit sync`)
- **Safe uninstall** - Safely uninstall bgit and restore all repositories (`bgit uninstall`)
- **Cross-platform support** - Works on Linux, macOS, and Windows

