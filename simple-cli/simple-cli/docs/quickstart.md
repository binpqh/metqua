# Quickstart

Get up and running with **simple-cli** in under 5 minutes.

---

## Prerequisites

- Go 1.22+ (only needed if building from source)
- Or: download a pre-built binary from [GitHub Releases](https://github.com/your-org/simple-cli/releases)

---

## Install

### Linux / macOS (shell installer)

```bash
curl -sSL https://github.com/your-org/simple-cli/releases/latest/download/install.sh | bash
```

### macOS (Homebrew)

```bash
brew install your-org/tap/simple-cli
```

### Windows (PowerShell)

```powershell
irm https://github.com/your-org/simple-cli/releases/latest/download/install.ps1 | iex
```

### Windows (NSIS installer)

Download `simple-cli-setup.exe` from the [Releases page](https://github.com/your-org/simple-cli/releases) and run it. The installer registers `simple-cli` on the system PATH automatically.

---

## Verify Installation

Open a **new** terminal (so PATH changes take effect):

```bash
simple-cli --version
# simple-cli dev (commit: unknown, built: unknown)
```

---

## Start Your First Session

```bash
simple-cli session start --name my-project
# Session 'my-project' started (id: 550e8400-...)
```

Optional: let the name be auto-generated:

```bash
simple-cli session start
# Session 'bold-river' started (id: ...)
```

---

## List Sessions

```bash
simple-cli session list
# ID         NAME                     STATUS     CREATED
# 550e8400   my-project               active     2026-03-23T10:00:00Z
```

---

## Close Terminal and Resume

Close your terminal window, then open a new one:

```bash
simple-cli session resume --name my-project
# Resumed session 'my-project'
```

Your session state is exactly where you left it.

---

## Stop a Session

```bash
simple-cli session stop --name my-project
# Session 'my-project' stopped
```

---

## Reset a Session

Deletes the session and starts a fresh one with the same name:

```bash
simple-cli session reset --name my-project --force
# Session 'my-project' reset (new id: ...)
```

Without `--force`, you will be prompted to confirm:

```bash
simple-cli session reset --name my-project
# Reset session 'my-project'? All state will be lost. [y/N] y
# Session 'my-project' reset (new id: ...)
```

---

## AI Agent / JSON Mode

Perfect for scripting or AI agent tool-use:

```bash
export SIMPLE_CLI_OUTPUT=json

simple-cli session start --name agent-workflow
# {"status":"ok","data":{"id":"...","name":"agent-workflow","status":"active",...},"meta":{...}}

simple-cli session list
# {"status":"ok","data":{"sessions":[...],"total":1},"meta":{...}}

simple-cli session stop --name agent-workflow
# {"status":"ok","data":{"id":"...","status":"stopped",...},"meta":{...}}
```

Errors write JSON to stderr:

```bash
simple-cli session resume --name no-such-session 2>&1 | cat
# {"status":"error","code":"SESSION_NOT_FOUND","message":"...","hint":"...","meta":{...}}
echo $?  # 3
```

See [docs/ai-agent-guide.md](ai-agent-guide.md) for full schema documentation and code examples.

---

## Environment Variables Reference

| Variable               | Default | Description                              |
| ---------------------- | ------- | ---------------------------------------- |
| `SIMPLE_CLI_OUTPUT`    | `human` | `json` for machine-readable output       |
| `SIMPLE_CLI_LOG_LEVEL` | `info`  | `debug`, `info`, `warn`, `error`         |
| `NO_COLOR`             | (unset) | Any non-empty value disables ANSI colour |
| `SIMPLE_CLI_STATE_DIR` | (auto)  | Override the session state directory     |

See [docs/configuration.md](configuration.md) for the full reference.
