# Roadmap

## Format

- ID
- Title
- Description
- Priority
- Status

---

## Items

### R-001

- Title: Core Config Extraction
- Description: Move config loading, saving, validation, and active identity handling into `/core/config`
- Priority: High
- Status: done

---

### R-002

- Title: Core SSH Extraction
- Description: Move SSH key generation, SSH config updates, and auth testing into `/core/ssh`
- Priority: High
- Status: done

---

### R-003

- Title: Core Identity Extraction
- Description: Move identity add/remove/update/activate logic into `/core/identity`
- Priority: High
- Status: done

---

### R-004

- Title: CLI Core Integration
- Description: Update CLI commands to use extracted core modules instead of direct implementations
- Priority: High
- Status: done

---

### R-005

- Title: Shared Models
- Description: Create reusable shared structs for identities, config, requests, and responses
- Priority: Medium
- Status: done

---

### R-006

- Title: Standardized Result Responses
- Description: Standardize core module responses using structured result objects
- Priority: Medium
- Status: done

---

### R-007

- Title: Config Safety Improvements
- Description: Add config validation, atomic writes, and config version handling
- Priority: Medium
- Status: done

---

### R-008

- Title: Integration Test Improvements
- Description: Improve real Git integration tests using isolated temporary environments and cleanup handling
- Priority: High
- Status: done

---

### R-009

- Title: BGIT Export System
- Description: Implement `.bgit` export archive generation and packaging flow
- Priority: High
- Status: done

---

### R-010

- Title: Export Encryption Layer
- Description: Add Argon2id and AES-256-GCM encryption for `.bgit` archives
- Priority: High
- Status: done

---

### R-011

- Title: BGIT Import System
- Description: Implement encrypted `.bgit` import and restore flow
- Priority: High
- Status: done

---

### R-011A

- Title: Portable SSH Key Backup Support
- Description: Include configured identity SSH key pairs in encrypted `.bgit` backups and restore them onto imported machines with target-machine paths
- Priority: High
- Status: done

---

### R-012

- Title: Desktop App Foundation
- Description: Initialize Wails desktop application structure and backend integration
- Priority: High
- Status: done

---

### R-013

- Title: Desktop Identity Dashboard
- Description: Create desktop UI for viewing identities and active profile
- Priority: Medium
- Status: done

---

### R-014

- Title: Desktop Identity Management
- Description: Add identity add/remove/update/activate functionality in desktop UI
- Priority: Medium
- Status: done

---

### R-015

- Title: Desktop Doctor Integration
- Description: Add visual diagnostics and SSH status checks in desktop UI
- Priority: Medium
- Status: done

---

### R-016

- Title: Desktop Notification System
- Description: Replace raw terminal-style output with notifications, dialogs, and status indicators
- Priority: Medium
- Status: done

---

### R-016A

- Title: Desktop Backup and Restore
- Description: Add desktop UI for encrypted bgit export/import, including .bgit file selection, password prompts, restore summary, and clear safety messaging.
- Priority: High
- Status: done

---

### R-016B

- Title: Desktop SSH Public Key Display
- Description: Show each identity's SSH public key in the desktop UI with a copy action so users can add the key to GitHub without manually opening key files.
- Priority: High
- Status: done

---

### R-017

- Title: GitHub Avatar Integration
- Description: Show GitHub profile avatars and identity cards in desktop UI
- Priority: Low
- Status: done

---

### R-018

- Title: Repo Identity Mapping
- Description: Add repository-to-identity binding and management
- Priority: Medium
- Status: done

---

### R-019

- Title: Automatic Identity Switching
- Description: Detect repositories and switch identities automatically
- Priority: Medium
- Status: done

---

### R-019-A

- Title: UI Polish & Interaction Improvements
- Description: Improve desktop spacing, alignment, responsiveness, consistent SVG icon actions, copy affordances, active identity visibility, table readability, and smooth update interactions without changing layout, navigation, or business logic
- Priority: Medium
- Status: done

---

### R-019-B

- Title: Desktop UI Redesign & Navigation
- Description: Reorganize the desktop app into a structured sidebar navigation layout with dedicated Dashboard, Identities, Repository Bindings, Backup & Restore, and Doctor pages while preserving existing functionality and core behavior
- Priority: High
- Status: done

---

### R-020

- Title: Cross Platform Packaging
- Description: Add Linux, Windows, and macOS packaging support for desktop application
- Priority: Medium
- Status: done

---

### R-021

- Title: Release Stabilization
- Description: Final cleanup, testing, bug fixes, documentation updates, and release preparation
- Priority: High
- Status: in progress

---

## Desktop-First Direction (R-022 onward)

Goal: users can install the desktop app and manage identities, SSH keys, and repositories without using the CLI.

- The CLI stays as it is. CLI behavior and output must not change.
- The desktop app manages identities, keys, and repositories, and also covers everyday Git work (view changes, commit, push, pull, branches, history) so users do not need a terminal or a second Git client.
- Shared logic is extracted into `/core` so the CLI and desktop app use the same implementation.

---

### R-022

- Title: Core Setup and Check Extraction
- Description: Move setup, managed pre-push hook installation, and push safety check logic from `cmd/` into `/core` so the desktop app can reuse it, without changing CLI behavior or output
- Priority: High
- Status: done

---

### R-023

- Title: Desktop First-Run Onboarding
- Description: Add a first-run wizard that checks Git, runs setup, adds the first identity, and generates its SSH key; let the desktop binary answer `check --from-hook` so the pre-push hook works without a separate CLI install
- Priority: High
- Status: done

---

### R-024

- Title: Desktop Clone and Add Existing Repository
- Description: Clone a GitHub repository with a chosen identity and destination folder, and add an existing local repository by binding it and fixing its remote in one step
- Priority: High
- Status: done

---

### R-025

- Title: Desktop Workspaces
- Description: Add a desktop page to view, register, and remove workspace folders and their identities using the existing core repo modules
- Priority: Medium
- Status: done

---

### R-026

- Title: Desktop Repository Health
- Description: Show all bound repositories with identity and remote status, and provide a one-click fix backed by the existing check and sync logic
- Priority: Medium
- Status: done

---

### R-027

- Title: GitHub Key Helper and Connection Test
- Description: Add an action to open GitHub SSH key settings with the public key copied, and a per-identity SSH connection test
- Priority: Medium
- Status: done

---

### R-028

- Title: Desktop Uninstall and Cleanup
- Description: Add a settings page to restore remotes, hooks, SSH config, and Git identity and remove bgit configuration, matching CLI uninstall behavior
- Priority: Medium
- Status: done

---

### R-029

- Title: Desktop Installers and Signing
- Description: Produce real installers for Linux, Windows, and macOS instead of raw binaries, and add code signing where available
- Priority: High
- Status: done

---

### R-030

- Title: Core Git Operations
- Description: Add a `/core/gitops` module that wraps the Git CLI for status, diff, stage, unstage, discard, commit, branches, history, fetch, pull, and push, returning structured results
- Priority: High
- Status: done

---

### R-031

- Title: Desktop Changes and Commit
- Description: Add a repository view that lists changed files, shows diffs, stages or discards files, and commits with the repository's bgit identity
- Priority: High
- Status: done

---

### R-032

- Title: Desktop Sync, Branches, and History
- Description: Add fetch, pull, and push with ahead/behind status, branch switching and creation, and a commit history view; pushes run the same safety check as the CLI hook
- Priority: High
- Status: done

---

### R-033

- Title: Desktop UI/UX Redesign
- Description: Redesign the desktop frontend as one coherent design system: top navigation instead of a sidebar, dark theme by default with a light theme, a three-part repository workspace (files, diff, commit) with History, Branches, Remotes, and Identity tabs, an identity manager, diagnostic Doctor, categorised Settings, and a redesigned setup guide. No change to CLI behaviour or to how Git, backups, or identities work
- Priority: High
- Status: done

### R-034

- Title: Desktop UI/UX Refinement
- Description: Refine the redesigned desktop app without redesigning it: resizable and collapsible panes in the repository workspace, a diff that marks the changed part of each line and can wrap, clearer sync status, a "Commits as" identity control that explains where a repository's identity comes from and how it relates to the default identity, top navigation that overflows into a "More" menu in narrow windows, higher contrast for secondary text, fewer nested cards, retry actions on load errors, and window-size tests at 1366x768, 1280x720, 1024x768, and 960x640. No backend or CLI change
- Priority: High
- Status: done

## Release Milestones

- Note : once a milestone is reached, give a reminder to update the roadmap and release notes and to release a new version.

### v0.4.0

Release after:

- R-001
- R-002
- R-003
- R-004

---

### v0.5.0

Release after:

- R-005
- R-006
- R-007
- R-008

---

### v0.6.0

Release after:

- R-009
- R-010
- R-011

---

### v0.7.0

Release after:

- R-012
- R-013
- R-014
- R-015

---

### v0.8.0

Release after:

- R-016
- R-017

---

### v0.9.0

Release after:

- R-018
- R-019
- R-019-A
- R-019-B

---

Note: v0.10.0, v0.11.0, and v0.12.0 were completed together and ship as a single v0.12.0 release.

### v0.10.0

Desktop app works without the CLI.

Release after:

- R-020
- R-022
- R-023
- R-024

---

### v0.11.0

Desktop app covers the remaining CLI workflows.

Release after:

- R-025
- R-026
- R-027
- R-028

---

### v0.12.0

Desktop app covers everyday Git work.

Release after:

- R-030
- R-031
- R-032
- R-033
- R-034

---

### v1.0.0

Release after:

- R-029
- R-021
