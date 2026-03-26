# AI Agent Integration Guide

**simple-cli** is designed for programmatic use by AI agents, automation scripts, and CI pipelines.
Every command supports `--output json` (or `SIMPLE_CLI_OUTPUT=json`) to produce structured,
stable JSON on stdout while keeping logs on stderr.

---

## Quick Start for Agents

```bash
# Preferred: set once in the agent's environment
export SIMPLE_CLI_OUTPUT=json
export SIMPLE_CLI_LOG_LEVEL=warn   # suppress info logs on stderr

# Invoke your sub-commands
simple-cli example
simple-cli run   # blocks until shutdown
```

---

## Envelope Schemas

### Success Envelope

Every successful command writes exactly one JSON line to **stdout**:

```json
{
  "status": "ok",
  "data": "<command-specific object or array>",
  "meta": {
    "version": "2.0.0",
    "duration_ms": 12,
    "command": "run"
  }
}
```

### Error Envelope

Every failed command writes exactly one JSON line to **stderr** and exits with a non-zero code:

```json
{
  "status": "error",
  "code": "INVALID_ARGUMENT",
  "message": "invalid output format 'xml'",
  "hint": "Use 'human' or 'json'.",
  "meta": {
    "version": "2.0.0",
    "duration_ms": 1,
    "command": "run"
  }
}
```

| Field              | Type                        | Stable | Description                             |
| ------------------ | --------------------------- | ------ | --------------------------------------- |
| `status`           | `"ok"` \| `"error"`         | ✅     | Top-level discriminator                 |
| `data`             | object \| array             | ✅     | Command payload (success only)          |
| `code`             | SCREAMING_SNAKE_CASE string | ✅     | Machine-readable error discriminator    |
| `message`          | string                      | ❌     | Human-readable description (may change) |
| `hint`             | string (optional)           | ❌     | Remediation suggestion (may change)     |
| `meta.version`     | semver string               | ✅     | CLI version that produced the response  |
| `meta.duration_ms` | integer                     | ✅     | Wall-clock ms for the command           |
| `meta.command`     | string                      | ✅     | Canonical command name                  |

---

## Exit Code Table

| Code | Name              | Condition                                    |
| ---- | ----------------- | -------------------------------------------- |
| `0`  | Success           | Command completed successfully               |
| `1`  | General Error     | Unclassified internal error                  |
| `2`  | Invalid Argument  | Missing/invalid flags or argument validation |

| Error Code (`code` field)   | Exit Code |
| --------------------------- | --------- |
| `INVALID_ARGUMENT`          | 2         |
| `INTERNAL_ERROR`            | 1         |
| `CONTEXT_CANCELED`          | 1         |

---

## Invocation Examples

### auth login / status / logout

All auth commands support `--output json`. Use `jq` to parse the results:

```bash
export SIMPLE_CLI_OUTPUT=json

# Login: opens browser for device flow; blocks until approved
simple-cli auth login --provider my-api

# Check login state
simple-cli auth status --provider my-api | jq '{provider: .data.provider, expired: .data.expired}'
# Example output:
# { "provider": "my-api", "expired": false }

# Logout a single provider
simple-cli auth logout --provider my-api

# Logout all providers
simple-cli auth logout --all
```

**auth status JSON envelope** (`--output json`):
```json
{
  "status": "ok",
  "data": {
    "provider": "my-api",
    "authenticated": true,
    "expired": false,
    "expires_in": "23h59m"
  },
  "meta": { "version": "2.1.0", "duration_ms": 3, "command": "auth status" }
}
```

### chat

`chat` streams the response to stdout. Use `--output json` to get a stable JSON envelope after streaming completes:

```bash
# Human-readable streaming (default)
simple-cli chat "Explain Go interfaces in 2 sentences"

# JSON envelope after stream completes
simple-cli chat --output json "Explain Go interfaces" | jq '.data.content'

# Pipe a file for review
cat mycode.go | simple-cli chat "Review this Go code for logic errors"

# Multi-turn conversation tracking
simple-cli chat --conversation "session-1" "What is 2+2?"
simple-cli chat --conversation "session-1" --system "You are a math tutor" "Now multiply by 3"

# Override model for one request
simple-cli chat --model gpt-4o-mini "Quick question: what's 10^6?"
```

**chat JSON envelope** (`--output json`):
```json
{
  "status": "ok",
  "data": {
    "provider": "my-api",
    "model": "gpt-4o",
    "content": "Go interfaces define a set of method signatures...",
    "streaming": true
  },
  "meta": { "version": "2.1.0", "duration_ms": 812, "command": "chat" }
}
```

Extract just the content:
```bash
simple-cli chat --output json "Hello" | jq -r '.data.content'
```

### Bash

```bash
#!/usr/bin/env bash
set -euo pipefail

export SIMPLE_CLI_OUTPUT=json

# Run a quick command and parse JSON
response=$(simple-cli example)
status=$(echo "$response" | jq -r '.status')
echo "Status: $status"

# Start daemon and capture its exit status
simple-cli run &
daemon_pid=$!
wait $daemon_pid
echo "Daemon exited with code $?"
```

### PowerShell

```powershell
$env:SIMPLE_CLI_OUTPUT = "json"

# Run example and parse response
$response = simple-cli example | ConvertFrom-Json
Write-Host "Status: $($response.status)"

# Start daemon, wait for it
$p = Start-Process simple-cli -ArgumentList "run" -PassThru -NoNewWindow
$p.WaitForExit()
Write-Host "Daemon exited: $($p.ExitCode)"
```

### Python

```python
import json
import subprocess
from typing import Any


def run_cli(*args: str) -> tuple[int, dict[str, Any], dict[str, Any]]:
    """Run simple-cli with JSON output. Returns (exit_code, stdout_json, stderr_json)."""
    result = subprocess.run(
        ["simple-cli", *args],
        capture_output=True,
        text=True,
        env={**__import__("os").environ, "SIMPLE_CLI_OUTPUT": "json"},
    )
    stdout = json.loads(result.stdout) if result.stdout.strip() else {}
    stderr = json.loads(result.stderr) if result.stderr.strip() else {}
    return result.returncode, stdout, stderr


# Run example command
code, data, _ = run_cli("example")
assert code == 0, f"Unexpected exit code: {code}"
print("message:", data["data"]["message"])
```

---

## Environment Variables

| Variable               | Equivalent Flag | Values                           | Default  |
| ---------------------- | --------------- | -------------------------------- | -------- |
| `SIMPLE_CLI_OUTPUT`    | `--output`      | `human`, `json`                  | `human`  |
| `SIMPLE_CLI_LOG_LEVEL` | `--log-level`   | `debug`, `info`, `warn`, `error` | `info`   |
| `NO_COLOR`             | `--no-color`    | any non-empty value              | (unset)  |

---

## Streaming / Progress (stderr JSON-Lines)

When `--log-level debug` and `--output json` are both active, diagnostic log
entries are written as JSON-Lines to **stderr** using Go's `slog.NewJSONHandler`:

```json
{"time":"2026-03-23T10:00:01Z","level":"DEBUG","msg":"acquiring session lock","id":"550e8400"}
{"time":"2026-03-23T10:00:01Z","level":"INFO","msg":"session created","id":"550e8400","name":"my-workflow"}
```

Agents that only need the final result should redirect stderr to `/dev/null`
(or `$null` on Windows) and parse only stdout.

---

## Stability Guarantees

- **Stable** (no breaking changes without major version bump):
  - All `code` values in error envelopes
  - Top-level envelope field names (`status`, `data`, `code`, `meta.*`)
  - Exit code meanings
  - `data` field names and types in command responses

- **Not stable** (may change in patch/minor releases):
  - `message` and `hint` text content
  - Log message text on stderr
  - Output formatting in human mode
