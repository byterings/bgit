# Windows Installer Package

This folder contains all files needed to create the Windows installer for bgit.

## Contents

- `bgit-installer.iss` - Inno Setup script for creating the installer
- `INSTALLER-README.txt` - Detailed instructions for compiling the installer
- `LICENSE` - MIT License (shown during installation)
- `WINDOWS-USER-GUIDE.md` - User guide for Windows users
- `README.md` - This file

## How to Create Installer

### Prerequisites

1. Windows machine with Inno Setup installed
   - Download from: https://jrsoftware.org/isinfo.php

### Steps

1. **Build Windows executable:**
   ```bash
   # From project root on Linux/Mac
   ./build-release.sh v0.1.0

   # Or manually
   GOOS=windows GOARCH=amd64 go build -o bgit.exe
   ```

2. **Copy bgit.exe to this folder:**
   ```bash
   cp release/bgit-windows-amd64.exe windows-installer/bgit.exe
   ```

3. **On Windows machine:**
   - Open Inno Setup Compiler
   - File → Open → Select `bgit-installer.iss`
   - Build → Compile (or press F9)
   - Output: `bgit-installer-v0.1.0.exe`

4. **Test the installer:**
   - Double-click `bgit-installer-v0.1.0.exe`
   - Follow installation wizard
   - Test: Open Command Prompt → `bgit --version`

5. **Distribute:**
   - Upload to GitHub Releases
   - Users can download and install with one click

## Version Updates

When releasing a new version:

1. Update version in `bgit-installer.iss` (line 3):
   ```iss
   AppVersion=0.1.0  → AppVersion=1.1.0
   ```

2. Update output filename (line 9):
   ```iss
   OutputBaseFilename=bgit-installer-v0.1.0  → OutputBaseFilename=bgit-installer-v1.1.0
   ```

3. Build new bgit.exe with new version
4. Compile installer
5. Upload to releases

## What the Installer Does

- Installs bgit.exe to `C:\Program Files\bgit\`
- Adds bgit to System PATH automatically
- Creates Start Menu shortcuts
- Creates Uninstaller
- Shows LICENSE during installation
- Requires Administrator privileges

## Notes

- The `bgit.exe` file is NOT included in git (too large, binary artifact)
- Build fresh `bgit.exe` for each release
- Keep this folder structure intact for easy releases
- All other files are tracked in git and stay up to date
