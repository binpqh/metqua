# Installation Guide

**simple-cli** ships first-class installers for Windows, Linux, and macOS.
All installers automatically register `simple-cli` on your PATH.

---

## macOS

### Homebrew (recommended)

```sh
brew install simple-cli
```

PATH is managed automatically by Homebrew. Verify:

```sh
simple-cli --version
```

### Shell Script (fallback)

```sh
curl -fsSL https://github.com/binpqh/simple-cli/releases/latest/download/install.sh | sh
```

Installs to `~/.local/bin` and appends to `~/.zshrc`. Open a new terminal to apply.

### PKG Installer (enterprise / MDM)

Download `simple-cli_darwin_universal.pkg` from the [Releases page](https://github.com/binpqh/simple-cli/releases)
and double-click it. The postinstall script writes `/etc/paths.d/simple-cli`.

---

## Linux

### Shell Script (recommended)

```sh
curl -fsSL https://github.com/binpqh/simple-cli/releases/latest/download/install.sh | sh
```

Installs to `~/.local/bin` (or `/usr/local/bin` with sudo) and appends to `~/.bashrc` and `~/.zshrc`.

### .deb Package (Debian/Ubuntu)

```sh
curl -fsSL https://github.com/binpqh/simple-cli/releases/latest/download/simple-cli_linux_amd64.deb \
     -o simple-cli.deb
sudo dpkg -i simple-cli.deb
```

### .rpm Package (Fedora/RHEL/Rocky)

```sh
sudo rpm -i https://github.com/binpqh/simple-cli/releases/latest/download/simple-cli_linux_amd64.rpm
```

Both packages install to `/usr/local/bin` and register `/etc/profile.d/simple-cli.sh`.

---

## Windows

### PowerShell (recommended, no elevation required)

```powershell
irm https://github.com/binpqh/simple-cli/releases/latest/download/install.ps1 | iex
```

Installs to `%LOCALAPPDATA%\simple-cli\bin` and registers user-scope PATH.
Open a new PowerShell session to apply.

### NSIS Installer (GUI or enterprise push)

Download `simple-cli_windows_amd64_setup.exe` from the [Releases page](https://github.com/binpqh/simple-cli/releases)
and run it. The installer registers machine-scope PATH via the Windows Registry.

---

## Verify Installation

After any install method, open a **new terminal** and run:

```sh
simple-cli --version
# simple-cli version 1.0.0 (commit abc1234, built 2026-03-24T10:00:00Z)
```

---

## PATH Troubleshooting

| Platform    | Check                                    | Fix                                                                  |
| ----------- | ---------------------------------------- | -------------------------------------------------------------------- |
| Linux/macOS | `echo $PATH \| grep simple-cli`          | Open a new terminal shell; source `~/.bashrc` / `~/.zshrc`           |
| macOS (PKG) | `cat /etc/paths.d/simple-cli`            | Should contain `/usr/local/bin`; if missing re-run installer         |
| Windows     | `$env:PATH -split ';' \| sls simple-cli` | Open a new PowerShell; log out and back in for machine-scope changes |

If `simple-cli --version` still fails after opening a new terminal, add the install directory to your PATH manually:

**Linux/macOS** (`~/.bashrc` or `~/.zshrc`):

```sh
export PATH="$PATH:$HOME/.local/bin"
```

**Windows** (PowerShell, user scope):

```powershell
[Environment]::SetEnvironmentVariable("PATH",
  [Environment]::GetEnvironmentVariable("PATH","User") + ";$env:LOCALAPPDATA\simple-cli\bin",
  "User")
```

---

## Uninstall

```sh
# Linux/macOS
rm "$(which simple-cli)"
# Remove PATH lines from ~/.bashrc, ~/.zshrc, ~/.profile

# Windows PowerShell
Remove-Item "$env:LOCALAPPDATA\simple-cli" -Recurse -Force
# Remove the PATH entry via system Properties > Environment Variables
```

Or use `Control Panel > Programs > Uninstall a Program` if you used the NSIS installer.
