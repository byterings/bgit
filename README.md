# bgit — Multi Git Identity Manager

Never commit with the wrong Git identity again.

## Description

bgit is a CLI tool that helps developers manage multiple Git identities (work, personal, client) without manually editing `.gitconfig` or SSH configuration.

## Install

bgit comes as a **desktop app** and as a **command-line tool (CLI)**. Each works on its own, and they share the same settings, so you can install one or both.

| | Windows 10/11 | macOS 11+ | Ubuntu 22.04+ / Debian 12+ | Other Linux |
|---|---|---|---|---|
| Desktop app | [Installer (.exe)](https://github.com/byterings/bgit/releases/latest/download/bgit-desktop-windows-amd64-installer.exe) | [Disk image (.dmg)](https://github.com/byterings/bgit/releases/latest/download/bgit-desktop-darwin-universal.dmg) | [Package (.deb)](https://github.com/byterings/bgit/releases/latest/download/bgit-desktop-linux-amd64.deb) | [Program file](https://github.com/byterings/bgit/releases/latest/download/bgit-desktop-linux-amd64) |
| CLI | [PowerShell](#cli-on-windows) | [Terminal](#cli-on-macos) | [Terminal](#cli-on-linux) | [Terminal](#cli-on-linux) |

All downloads are on the [latest release](https://github.com/byterings/bgit/releases/latest) page. The desktop app is 64-bit only (x86-64 on Windows and Linux; Intel and Apple Silicon on macOS).

### Before you start: install Git

bgit uses the Git already on your computer. Check with `git --version` in a terminal.

- **Windows:** install [Git for Windows](https://git-scm.com/download/win) with the default options.
- **macOS:** run `xcode-select --install`, or install Git with [Homebrew](https://brew.sh): `brew install git`.
- **Linux:** `sudo apt install git` (Ubuntu, Debian), `sudo dnf install git` (Fedora), `sudo pacman -S git` (Arch). The `.deb` package installs Git for you.

### Desktop app on Windows

1. Download [`bgit-desktop-windows-amd64-installer.exe`](https://github.com/byterings/bgit/releases/latest/download/bgit-desktop-windows-amd64-installer.exe).
2. Open it. If Windows shows "Windows protected your PC", choose **More info**, then **Run anyway**. (This appears until releases are code signed.)
3. Allow the installer to make changes (it installs for all users, into `C:\Program Files\ByteRings\bgit`) and follow the steps.
4. Open **bgit** from the Start menu or the desktop shortcut.

The app uses Microsoft Edge WebView2, which Windows 10 and 11 normally already have. If it is missing, the installer downloads it, so keep the computer online while installing.

No admin rights? [`bgit-desktop-windows-amd64.exe`](https://github.com/byterings/bgit/releases/latest/download/bgit-desktop-windows-amd64.exe) is the same app without an installer: save it anywhere and run it (it needs WebView2).

**Update:** download and run the newest installer; it replaces the old version. **Uninstall:** Settings → Apps → Installed apps → bgit → Uninstall.

### Desktop app on macOS

1. Download [`bgit-desktop-darwin-universal.dmg`](https://github.com/byterings/bgit/releases/latest/download/bgit-desktop-darwin-universal.dmg). It runs on both Intel and Apple Silicon Macs.
2. Open the file and drag **bgit** onto **Applications**.
3. Open bgit from Applications or Launchpad.
4. If macOS says the app "cannot be opened" or "cannot be verified", open **System Settings → Privacy & Security**, scroll down, and choose **Open Anyway** next to bgit. Confirm once; later launches open normally. (This appears until releases are signed and notarized.)

If you prefer the terminal for step 4: `xattr -dr com.apple.quarantine /Applications/bgit.app`.

**Update:** download the newest `.dmg` and drag bgit to Applications again, replacing the old copy. **Uninstall:** quit bgit and move it from Applications to the Bin.

### Desktop app on Linux

**Ubuntu 22.04 or later, Debian 12 or later, Linux Mint 21 or later** (and other Debian-based systems):

```bash
curl -LO https://github.com/byterings/bgit/releases/latest/download/bgit-desktop-linux-amd64.deb
sudo apt install ./bgit-desktop-linux-amd64.deb
```

`apt` also installs what the app needs (GTK 3, WebKitGTK 4.1, Git, OpenSSH). Start **bgit** from your applications menu, or run `bgit-desktop` in a terminal.

**Update:** download the newest `.deb` and run the same `sudo apt install` command. **Uninstall:** `sudo apt remove bgit-desktop`.

**Other distributions (Fedora, Arch, openSUSE, …):** install GTK 3 and WebKitGTK 4.1, then run the program file:

```bash
# Fedora
sudo dnf install gtk3 webkit2gtk4.1 git openssh-clients
# Arch
sudo pacman -S gtk3 webkit2gtk-4.1 git openssh

curl -L https://github.com/byterings/bgit/releases/latest/download/bgit-desktop-linux-amd64 -o bgit-desktop
chmod +x bgit-desktop
mkdir -p ~/.local/bin && mv bgit-desktop ~/.local/bin/
bgit-desktop
```

### First launch

The first time it opens, the app walks you through setup: it checks Git, turns on push protection, and adds your first identity with an SSH key (or restores everything from a `.bgit` backup). If you already use the bgit CLI, the app picks up your existing identities and repositories; nothing needs to be set up again.

### Removing bgit completely

Uninstalling the app only removes the program. To also undo what bgit set up (the push check, SSH settings, and repository addresses), open the app first and choose **Settings → Advanced → Remove bgit from this computer**, then uninstall. Skip this step if you keep using the bgit CLI, because the CLI uses the same setup. CLI users can do the same with `bgit uninstall`.

### CLI on Linux

```bash
curl -L https://github.com/byterings/bgit/releases/latest/download/bgit-linux-amd64 -o bgit
chmod +x bgit
sudo mv bgit /usr/local/bin/
```

For Linux ARM devices, use `bgit-linux-arm64`.

### CLI on macOS

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

### CLI on Windows

```powershell
irm https://raw.githubusercontent.com/byterings/bgit/main/install.ps1 | iex
```

Check the CLI with `bgit --version`.

## Desktop App

The desktop app does everything without a terminal: first-run setup, identities and SSH keys, cloning, repository health, and everyday Git work (changes, commit, pull, push, branches, history). It does not need the CLI.

Navigation is a bar across the top: Dashboard, Repositories, Identities, Workspaces, Doctor, and Settings. The app opens in a dark theme; a light theme is in Settings. Useful shortcuts: `Ctrl+K` (`⌘K` on macOS) searches repositories, `Ctrl+1` to `Ctrl+5` switch areas, `Ctrl+Enter` commits, and `Ctrl+B` hides or shows the file list in a repository. In a repository, the header shows which identity commits there and why; the file list and commit area can be resized by dragging their edges.

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
