# Changelog

All notable changes to bgit will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

