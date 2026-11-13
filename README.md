# bgit — Multi Git Identity Manager

Never commit with the wrong Git identity again.

## Description

bgit is a CLI tool that helps developers manage multiple Git identities (work, personal, client) without manually editing `.gitconfig` or SSH configuration.

## Install

### Linux

```bash
curl -L https://github.com/byterings/bgit/releases/latest/download/bgit-linux-amd64 -o bgit
chmod +x bgit
sudo mv bgit /usr/local/bin/
```

For Linux ARM devices, use `bgit-linux-arm64`.

### macOS

Intel Macs:

```bash
curl -L https://github.com/byterings/bgit/releases/latest/download/bgit-darwin-amd64 -o bgit
chmod +x bgit
sudo mv bgit /usr/local/bin/
```

Apple Silicon Macs:

```bash
curl -L https://github.com/byterings/bgit/releases/latest/download/bgit-darwin-arm64 -o bgit
chmod +x bgit
sudo mv bgit /usr/local/bin/
```

### Windows

```powershell
irm https://raw.githubusercontent.com/byterings/bgit/main/install.ps1 | iex
```

## Desktop App

The desktop app does everything without a terminal: first-run setup, identities and SSH keys, cloning, repository health, and everyday Git work (changes, commit, pull, push, branches, history). It does not need the CLI.

Navigation is a bar across the top: Dashboard, Repositories, Identities, Workspaces, Doctor, and Settings. The app opens in a dark theme; a light theme is in Settings. Useful shortcuts: `Ctrl+K` (`⌘K` on macOS) searches repositories, `Ctrl+1` to `Ctrl+5` switch areas, `Ctrl+Enter` commits, and `Ctrl+B` hides or shows the file list in a repository. In a repository, the header shows which identity commits there and why; the file list and commit area can be resized by dragging their edges.

Download it from [GitHub Releases](https://github.com/byterings/bgit/releases/latest):

- **Windows:** run `bgit-desktop-windows-amd64-installer.exe`
- **macOS:** open `bgit-desktop-darwin-universal.dmg` and drag bgit to Applications
- **Linux (Debian, Ubuntu):** `sudo apt install ./bgit-desktop-linux-amd64.deb`

The first launch walks you through setup. Git must be installed.

Until releases are code signed, Windows SmartScreen and macOS Gatekeeper warn before the first launch. On Windows choose "More info", then "Run anyway". On macOS open System Settings, Privacy & Security, and choose "Open Anyway".

Before uninstalling the app, open Settings and choose "Remove bgit from this computer" so your repositories and Git configuration are restored.

## Quick Start

```bash
bgit add work
bgit use work
git commit
```

## Example Workflow

```bash
git config user.email
personal@gmail.com

bgit use work

✔ switched to work identity
```

## Features

- Manage multiple git identities
- Repo binding for identity
- Workspace-based identity
- SSH key management
- Diagnostics with `bgit doctor`

## Documentation

https://bgitcli.com/docs
