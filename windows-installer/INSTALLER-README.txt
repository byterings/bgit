========================================
bgit Windows Installer Package
========================================

This package contains everything you need to create the Windows installer.

Contents:
---------
1. bgit.exe                  - Windows executable (build before creating installer)
2. bgit-installer.iss        - Inno Setup script
3. LICENSE                   - MIT License file
4. WINDOWS-USER-GUIDE.md     - User guide
5. INSTALLER-README.txt      - This file
6. README.md                 - Developer documentation

Note: bgit.exe is NOT included in git. Build it fresh using:
  GOOS=windows GOARCH=amd64 go build -o windows-installer/bgit.exe .

========================================
How to Create Installer:
========================================

Step 1: Install Inno Setup
---------------------------
Download from: https://jrsoftware.org/isinfo.php
Install Inno Setup on your Windows machine.

Step 2: Open Project
--------------------
1. Open Inno Setup Compiler
2. File → Open → Select "bgit-installer.iss"

Step 3: Compile
---------------
1. Click "Build → Compile" (or press F9)
2. Wait for compilation to finish
3. Installer will be created: "bgit-installer-v0.1.0.exe"

Step 4: Test
------------
1. Double-click "bgit-installer-v0.1.0.exe"
2. Follow installation wizard
3. Check "Add bgit to system PATH"
4. Complete installation
5. Open new Command Prompt
6. Test: bgit --version

========================================
What the Installer Does:
========================================

✓ Installs bgit.exe to C:\Program Files\bgit\
✓ Adds bgit to System PATH automatically
✓ Creates Start Menu shortcuts
✓ Creates Uninstaller
✓ Shows LICENSE during installation
✓ Requires Administrator privileges

========================================
For Distribution:
========================================

Upload "bgit-installer-v0.1.0.exe" to:
- GitHub Releases
- Your website
- Package managers

Users just download and run - No manual setup needed!

========================================
Troubleshooting:
========================================

If compilation fails:
- Make sure bgit.exe exists in the same folder
- Make sure LICENSE file exists
- Check Inno Setup version (6.0 or later)

If PATH doesn't work after install:
- User needs to close and reopen Command Prompt
- Or restart Windows

========================================
ByteRings
https://github.com/byterings/bgit
========================================
