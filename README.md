# bgit — Multi Git Identity Manager

Never commit with the wrong Git identity again.

## Description

bgit is a CLI tool that helps developers manage multiple Git identities (work, personal, client) without manually editing `.gitconfig` or SSH configuration.

## Install

### Linux / macOS

```bash
curl -fsSL https://bgitcli.com/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/byterings/bgit/main/install.ps1 | iex
```

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
