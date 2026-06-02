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
- Status: pending

---

### R-011

- Title: BGIT Import System
- Description: Implement encrypted `.bgit` import and restore flow
- Priority: High
- Status: pending

---

### R-012

- Title: Desktop App Foundation
- Description: Initialize Wails desktop application structure and backend integration
- Priority: High
- Status: pending

---

### R-013

- Title: Desktop Identity Dashboard
- Description: Create desktop UI for viewing identities and active profile
- Priority: Medium
- Status: pending

---

### R-014

- Title: Desktop Identity Management
- Description: Add identity add/remove/update/activate functionality in desktop UI
- Priority: Medium
- Status: pending

---

### R-015

- Title: Desktop Doctor Integration
- Description: Add visual diagnostics and SSH status checks in desktop UI
- Priority: Medium
- Status: pending

---

### R-016

- Title: Desktop Notification System
- Description: Replace raw terminal-style output with notifications, dialogs, and status indicators
- Priority: Medium
- Status: pending

---

### R-017

- Title: GitHub Avatar Integration
- Description: Show GitHub profile avatars and identity cards in desktop UI
- Priority: Low
- Status: pending

---

### R-018

- Title: Repo Identity Mapping
- Description: Add repository-to-identity binding and management
- Priority: Medium
- Status: pending

---

### R-019

- Title: Automatic Identity Switching
- Description: Detect repositories and switch identities automatically
- Priority: Medium
- Status: pending

---

### R-020

- Title: Cross Platform Packaging
- Description: Add Linux, Windows, and macOS packaging support for desktop application
- Priority: Medium
- Status: pending

---

### R-021

- Title: Release Stabilization
- Description: Final cleanup, testing, bug fixes, documentation updates, and release preparation
- Priority: High
- Status: pending

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

---

### v1.0.0

Release after:

- R-020
- R-021
