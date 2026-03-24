# Quickstart: simple-cli

**Branch**: `001-cli-long-life-session` | **Date**: 2026-03-23

Get `simple-cli` installed and running your first long-life session in under 5 minutes.

---

## Prerequisites

- A supported operating system: Windows 10/11, Ubuntu 20.04+, Debian 11+, macOS 12+
- Internet connection (for download) or a pre-downloaded binary

---

## Step 1 — Install

### macOS (Homebrew — recommended)

```sh
brew install simple-cli
```

### Linux (install script)

```sh
curl -fsSL https://github.com/your-org/simple-cli/releases/latest/download/install.sh | bash
```

This places the binary in `~/.local/bin` (or `/usr/local/bin` with sudo) and updates
your `~/.bashrc` and `~/.zshrc` automatically.

### Windows (PowerShell)

```powershell
irm https://github.com/your-org/simple-cli/releases/latest/download/install.ps1 | iex
```

This installs to `%LOCALAPPDATA%\simple-cli\bin` and registers it in your user `PATH`.
No administrator rights required.

### Windows (NSIS installer)

Download `simple-cli_windows_amd64_setup.exe` from the [Releases page](https://github.com/your-org/simple-cli/releases)
and run it. The installer adds `simple-cli` to your system PATH automatically.

---

## Step 2 — Verify Installation

Open a **new** terminal (the PATH update requires a fresh shell):

```sh
simple-cli --version
# simple-cli version 1.0.0
```

If `simple-cli` is not found, see [docs/installation.md](../docs/installation.md) for
manual PATH troubleshooting.

---

## Step 3 — Start Your First Session

```sh
simple-cli session start --name my-first-session
# Session 'my-first-session' started (id: 550e8400-e29b-41d4-a716-446655440000)
```

No name? An auto-generated name (e.g., `bold-river`) is assigned:

```sh
simple-cli session start
# Session 'bold-river' started (id: 7f1e9a2b-cafe-beef-dead-123456789abc)
```

---

## Step 4 — List Sessions

```sh
simple-cli session list
# ID         NAME               STATUS    CREATED
# 550e8400   my-first-session   active    2 minutes ago
# 7f1e9a2b   bold-river         active    1 minute ago
```

---

## Step 5 — Close Terminal and Resume

Close your terminal completely. Open a new one:

```sh
simple-cli session resume --name my-first-session
# Resumed session 'my-first-session'
```

Your session state is intact — `simple-cli` persists sessions to disk.

---

## Step 6 — Stop and Reset

Stop a session (retain state on disk):

```sh
simple-cli session stop --name my-first-session
# Session 'my-first-session' stopped
```

Reset (delete + recreate with clean state):

```sh
simple-cli session reset --name my-first-session --force
# Session 'my-first-session' reset (new id: a1b2c3d4-...)
```

---

## AI Agent / Automation Mode

Use `--output json` for machine-readable output:

```sh
simple-cli session start --name ci-run --output json
# {"status":"ok","data":{"id":"...","name":"ci-run","status":"active",...},"meta":{...}}
```

Or set the environment variable globally:

```sh
export SIMPLE_CLI_OUTPUT=json
simple-cli session list
```

Errors go to stderr; payload goes to stdout. See [contracts/output-schema.md](contracts/output-schema.md)
for the full schema and [contracts/exit-codes.md](contracts/exit-codes.md) for exit codes.

---

## Configuration File (optional)

Create `~/.config/simple-cli/config.yaml` (Linux/macOS) or
`%APPDATA%\simple-cli\config.yaml` (Windows):

```yaml
output: human # or json
log_level: info # debug | info | warn | error
no_color: false
```

All settings can also be overridden per-invocation with flags or environment variables.

---

## What's Next?

- [docs/configuration.md](../docs/configuration.md) — all flags, env vars, and config file options
- [docs/architecture.md](../docs/architecture.md) — package layout and design decisions
- [docs/ai-agent-guide.md](../docs/ai-agent-guide.md) — full AI agent integration guide
- [docs/installation.md](../docs/installation.md) — detailed install instructions and PATH troubleshooting
