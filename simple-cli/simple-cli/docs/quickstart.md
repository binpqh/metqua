# Quickstart

Get up and running with **simple-cli** in under 5 minutes.

---

## Prerequisites

- Go 1.22+ (only needed if building from source)
- Or: download a pre-built binary from [GitHub Releases](https://github.com/binpqh/simple-cli/releases)

---

## Install

### Linux / macOS (shell installer)

```bash
curl -sSL https://github.com/binpqh/simple-cli/releases/latest/download/install.sh | bash
```

### macOS (Homebrew)

```bash
brew install binpqh/tap/simple-cli
```

### Windows (PowerShell)

```powershell
irm https://github.com/binpqh/simple-cli/releases/latest/download/install.ps1 | iex
```

### Windows (NSIS installer)

Download `simple-cli-setup.exe` from the [Releases page](https://github.com/binpqh/simple-cli/releases) and run it. The installer registers `simple-cli` on the system PATH automatically.

---

## Verify Installation

Open a **new** terminal (so PATH changes take effect):

```bash
simple-cli --version
# simple-cli dev (commit: unknown, built: unknown)
```

---

## Start the Daemon

```bash
simple-cli run
# simple-cli started — waiting for shutdown signal (Ctrl+C or SIGTERM)
```

The process blocks until it receives `SIGINT` or `SIGTERM` (e.g., `Ctrl+C`, `kill`, or system shutdown). It exits cleanly within 5 seconds.

---

## Using JSON Output

```bash
simple-cli --output json run
```

On shutdown:

```json
{"status":"ok","data":{"status":"stopped","uptime_ms":5123},"meta":{"version":"1.0.0","duration_ms":5123,"command":"run"}}
```

---

## Example Sub-Command

The template ships with an example sub-command to demonstrate the extension pattern:

```bash
simple-cli example
# Replace this with your logic

simple-cli --output json example
# {"status":"ok","data":{"message":"replace this with your logic"},"meta":{...}}
```

Delete `cmd/example_cmd.go` and add your own commands when customising the template.

---

## AI Agent / JSON Mode

Perfect for scripting or AI agent tool-use:

```bash
export SIMPLE_CLI_OUTPUT=json

simple-cli run
# blocks; on shutdown:
# {"status":"ok","data":{"status":"stopped","uptime_ms":5123},"meta":{...}}
```

See [docs/ai-agent-guide.md](ai-agent-guide.md) for full schema documentation and code examples.

---

## Environment Variables Reference

| Variable               | Default | Description                              |
| ---------------------- | ------- | ---------------------------------------- |
| `SIMPLE_CLI_OUTPUT`    | `human` | `json` for machine-readable output       |
| `SIMPLE_CLI_LOG_LEVEL` | `info`  | `debug`, `info`, `warn`, `error`         |
| `NO_COLOR`             | (unset) | Any non-empty value disables ANSI colour |

See [docs/configuration.md](configuration.md) for the full reference.
