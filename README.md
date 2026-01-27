# bgit - Multi-Git Identity Manager

**bgit** is a CLI tool for managing multiple Git identities on one system. Works on **Linux, macOS, and Windows**.

Switch Git identities with one command. Keep using normal `git` commands while bgit handles identity switching.

> **Version 0.1.0** - Initial release with core identity management.

## Why bgit?

If you have multiple GitHub accounts (work, personal, side projects), you know the pain:
- Manually editing `.gitconfig` and `.ssh/config`
- Accidentally pushing with the wrong identity
- Complex SSH host configurations
- Forgetting which account you're using

**bgit solves this:**
- One command to switch identities: `bgit use work`
- Automatic Git and SSH config management
- Keep using normal `git` commands
- Clear indication of active identity

## Installation

**Note**: Users do **NOT** need Go installed. Download pre-built binaries from [Releases](https://github.com/byterings/bgit/releases).

### Linux / macOS

```bash
curl -sSL https://raw.githubusercontent.com/byterings/bgit/main/install.sh | bash
```

Before running, you can inspect the script:
```bash
curl -sSL https://raw.githubusercontent.com/byterings/bgit/main/install.sh | less
```

### Windows

```powershell
irm https://raw.githubusercontent.com/byterings/bgit/main/install.ps1 | iex
```

Before running, you can inspect the script at:
https://github.com/byterings/bgit/blob/main/install.ps1

### Manual Installation

Download the binary for your platform from [Releases](https://github.com/byterings/bgit/releases):

```bash
# Linux (AMD64)
curl -L https://github.com/byterings/bgit/releases/download/v0.1.0/bgit-linux-amd64 -o bgit
chmod +x bgit
sudo mv bgit /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/byterings/bgit/releases/download/v0.1.0/bgit-darwin-arm64 -o bgit
chmod +x bgit
sudo mv bgit /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/byterings/bgit/releases/download/v0.1.0/bgit-darwin-amd64 -o bgit
chmod +x bgit
sudo mv bgit /usr/local/bin/
```

## Quick Start

### 1. Add your identities

```bash
# Interactive mode (recommended)
bgit add

# Or use flags
bgit add \
  --name "John Doe" \
  --email "john@work.com" \
  --github "john-work"
```

During setup, bgit can:
- Generate new SSH keys (Ed25519)
- Import existing SSH keys
- Skip SSH setup (add later)

### 2. Switch between identities

```bash
# Switch to work account
bgit use work

# Switch to personal account
bgit use personal
```

That's it. All `git` commands now use the active identity.

### 3. List your identities

```bash
bgit list
```

Output:
```
Configured users:

-> work                 john@work.com                  John Work
  personal             john@personal.com              John Personal
```

The `->` shows your active identity.

## What bgit Modifies

bgit manages two configuration areas:

### 1. Git Global Config

Updates `~/.gitconfig` (Linux/macOS) or `%USERPROFILE%\.gitconfig` (Windows):
```ini
[user]
    name = John Work
    email = john@work.com
```

**Only `user.name` and `user.email` are modified.** Other settings are untouched.

### 2. SSH Config

Adds a managed section to `~/.ssh/config`:
```
# ---- BEGIN BGIT MANAGED ----
Host github.com-john-work
  HostName github.com
  User git
  IdentityFile ~/.ssh/bgit_work
  IdentitiesOnly yes
# ---- END BGIT MANAGED ----
```

**Note:** The SSH host uses your GitHub username (e.g., `github.com-john-work`), not the alias.

**bgit only modifies content between these markers.** Your existing SSH config entries are preserved.

### 3. bgit Config

Stores its own configuration in `~/.bgit/config.toml`:
```toml
version = "1.0"
active_user = "work"

[[users]]
  alias = "work"
  name = "John Work"
  email = "john@work.com"
  github_username = "john-work"
  ssh_key_path = "/home/user/.ssh/bgit_work"
```

## Uninstall / Rollback

### Safe Uninstall (Recommended)

Use `bgit uninstall` to safely remove bgit:

```bash
bgit uninstall
```

This will:
1. Find all repositories with bgit remote URLs
2. Restore them to standard GitHub format
3. Remove bgit SSH config entries
4. Remove bgit configuration

Then manually delete the binary:
```bash
# Linux/macOS
sudo rm /usr/local/bin/bgit

# Windows
# Use Add/Remove Programs or: Remove-Item "$env:LOCALAPPDATA\bgit" -Recurse -Force
```

### Manual Uninstall

If you prefer manual removal:

1. **Restore repos**: Run `bgit remote restore` in each repository
2. **Remove binary**: `sudo rm /usr/local/bin/bgit`
3. **Remove config**: `rm -rf ~/.bgit`
4. **Clean SSH config**: Remove the `# ---- BEGIN BGIT MANAGED ----` section from `~/.ssh/config`
5. **Remove SSH keys** (optional): `rm ~/.ssh/bgit_*`
6. **Restore git config**:
   ```bash
   git config --global user.name "Your Name"
   git config --global user.email "your@email.com"
   ```

## Commands

| Command | Description |
|---------|-------------|
| `bgit add` | Add a new Git identity |
| `bgit list` | List all configured identities |
| `bgit use <alias>` | Switch to a different identity |
| `bgit clone <url>` | Clone repo with correct SSH config |
| `bgit remote fix` | Fix current repo's remote for active user |
| `bgit remote restore` | Restore remote to standard GitHub format |
| `bgit delete <alias>` | Remove an identity |
| `bgit update <alias>` | Update an identity's SSH key |
| `bgit sync [--fix]` | Validate configs match active user |
| `bgit setup-ssh` | (Windows) Start SSH agent and load keys |
| `bgit uninstall` | Safely uninstall bgit and restore all repos |

## SSH Key Management

When you add a user, bgit can:

1. **Generate new SSH key** - Creates Ed25519 key pair at `~/.ssh/bgit_<alias>`
2. **Import existing key** - Use your current SSH key
3. **Skip for now** - Add SSH key manually later

### Cloning Repositories

Use `bgit clone` to automatically use the correct SSH configuration:

```bash
bgit use work
bgit clone https://github.com/company/repo.git
```

This works with any GitHub URL (HTTPS or SSH) and converts it automatically.

### Fixing Existing Repositories

If you have an existing repo, fix its remote:

```bash
cd existing-repo
bgit use work
bgit remote fix
git push   # Now works with the correct identity
```

### Restoring Remotes

Before uninstalling bgit or to use standard git:

```bash
cd repo
bgit remote restore   # Restores to git@github.com:user/repo.git
```

## Troubleshooting

### SSH Permission Issues

SSH requires specific file permissions to work correctly:

```bash
# Fix SSH directory permissions (must be 700)
chmod 700 ~/.ssh

# Fix SSH key permissions (must be 600)
chmod 600 ~/.ssh/bgit_*
```

### Common Issues

**"Permission denied (publickey)"**
1. Ensure SSH key is added to your GitHub account
2. Check file permissions

**"Could not open a connection to your authentication agent"**
```bash
eval $(ssh-agent)
ssh-add ~/.ssh/bgit_*
```

## Limitations

- **GitHub-focused**: SSH config uses `github.com` hosts. GitLab/Bitbucket may require manual SSH config.
- **Config format may change**: The `~/.bgit/config.toml` format may change in future versions.

## Roadmap

### Phase 1 (v0.1.0) - Current
- [x] Global user management
- [x] SSH + Git config handling
- [x] User switching
- [x] Sync/validation
- [x] Cross-platform support
- [x] `bgit clone` - clone with correct SSH config
- [x] `bgit remote fix/restore` - manage remote URLs
- [x] `bgit uninstall` - safe uninstallation

### Phase 2 (Planned)
- [ ] `bgit workspace` - organized folders with auto-binding
- [ ] `bgit bind` - repo-bound identity (sticky ownership)
- [ ] `bgit status` - show identity status and bindings
- [ ] `bgit doctor` - diagnostics and auto-fix

### Phase 3 (Future)
- [ ] Shell prompt integration
- [ ] Pre-push safety checks

### Phase 4 (Future)
- [ ] UI for all operations

## FAQ

**Q: Does bgit wrap git commands?**
No. bgit only manages configuration. You use normal `git` commands.

**Q: What if I already have SSH keys?**
bgit can import existing keys. Provide the path when adding a user.

**Q: Is my existing .gitconfig safe?**
Yes. bgit only modifies `user.name` and `user.email`.

**Q: Is my existing SSH config safe?**
Yes. bgit only modifies content within its managed section markers.

**Q: Can I use bgit with GitLab/Bitbucket?**
Git config changes work anywhere. SSH config currently uses `github.com` hosts, so other providers may need manual SSH config adjustments.

**Q: How do I see what bgit will change?**
Run `bgit sync` to see current status without making changes.

## Contributing

Contributions welcome. This is an early-stage project focused on doing one thing well.

## License

MIT License - see [LICENSE](LICENSE)

---

**One command. Zero mistakes.**
