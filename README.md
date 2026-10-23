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

## Desktop Beta

Desktop release assets are attached to GitHub Releases:

- Linux: `bgit-desktop-linux-amd64`
- Windows: `bgit-desktop-windows-amd64.exe`
- macOS: `bgit-desktop-darwin-universal.zip`

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
