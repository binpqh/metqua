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

simple-cli session start --name my-workflow
simple-cli session list
simple-cli session resume --name my-workflow
simple-cli session stop --name my-workflow
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
    "version": "1.0.0",
    "duration_ms": 12,
    "command": "session start"
  }
}
```

### Error Envelope

Every failed command writes exactly one JSON line to **stderr** and exits with a non-zero code:

```json
{
  "status": "error",
  "code": "SESSION_NOT_FOUND",
  "message": "session 'my-project' not found",
  "hint": "Use 'simple-cli session list' to see available sessions.",
  "meta": {
    "version": "1.0.0",
    "duration_ms": 5,
    "command": "session resume"
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
| `3`  | Not Found         | Requested session does not exist             |
| `4`  | Permission Denied | State directory not writable                 |
| `5`  | Timeout           | Lock acquisition or context deadline timeout |

| Error Code (`code` field)   | Exit Code |
| --------------------------- | --------- |
| `SESSION_NOT_FOUND`         | 3         |
| `SESSION_NAME_CONFLICT`     | 1         |
| `SESSION_LOCK_TIMEOUT`      | 5         |
| `STORE_READ_ONLY`           | 4         |
| `INVALID_ARGUMENT`          | 2         |
| `INTERNAL_ERROR`            | 1         |
| `CONTEXT_CANCELED`          | 1         |
| `CONTEXT_DEADLINE_EXCEEDED` | 5         |

---

## Invocation Examples

### Bash

```bash
#!/usr/bin/env bash
set -euo pipefail

export SIMPLE_CLI_OUTPUT=json

# Start a session and capture the ID
response=$(simple-cli session start --name my-workflow)
session_id=$(echo "$response" | jq -r '.data.id')
echo "Started session: $session_id"

# Check exit code and parse error
if ! simple-cli session resume --name no-such-session 2>err.json; then
  code=$(jq -r '.code' err.json)
  case $code in
    SESSION_NOT_FOUND) echo "Session does not exist — creating it..." ;;
    SESSION_LOCK_TIMEOUT) echo "Locked by another process, retry later" ;;
    *) echo "Unexpected error: $code"; exit 1 ;;
  esac
fi

# List sessions, filter active ones
simple-cli session list | jq '.data[] | select(.status == "active")'
```

### PowerShell

```powershell
$env:SIMPLE_CLI_OUTPUT = "json"

# Start session
$response = simple-cli session start --name my-workflow | ConvertFrom-Json
$sessionId = $response.data.id
Write-Host "Session ID: $sessionId"

# Resume with error handling
& simple-cli session resume --name my-workflow 2>$null
switch ($LASTEXITCODE) {
    0 { Write-Host "Resumed successfully" }
    3 { Write-Host "Session not found" }
    5 { Write-Host "Timed out — retry" }
    default { Write-Host "Error: $LASTEXITCODE" }
}

# List all sessions
$sessions = (simple-cli session list | ConvertFrom-Json).data
$sessions | Where-Object { $_.status -eq "active" }
```

### Python

```python
import json
import subprocess
import sys
from typing import Any


def run_cli(*args: str) -> tuple[int, dict[str, Any], dict[str, Any]]:
    """Run simple-cli with JSON output. Returns (exit_code, stdout_json, stderr_json)."""
    result = subprocess.run(
        ["simple-cli", *args, "--output", "json"],
        capture_output=True,
        text=True,
    )
    stdout = json.loads(result.stdout) if result.stdout.strip() else {}
    stderr = json.loads(result.stderr) if result.stderr.strip() else {}
    return result.returncode, stdout, stderr


# Start session
code, data, _ = run_cli("session", "start", "--name", "my-workflow")
assert code == 0, f"Unexpected exit code: {code}"
session_id = data["data"]["id"]

# Resume with error handling
code, data, err = run_cli("session", "resume", "--name", "my-workflow")
if code != 0:
    error_code = err.get("code", "UNKNOWN")
    if error_code == "SESSION_NOT_FOUND":
        print("Session not found, creating...")
    elif error_code == "SESSION_LOCK_TIMEOUT":
        print("Locked — retry")
        sys.exit(5)
    else:
        raise RuntimeError(f"Unexpected error: {error_code}")

# List active sessions
code, data, _ = run_cli("session", "list")
active = [s for s in data.get("data", []) if s["status"] == "active"]
print(f"Active sessions: {len(active)}")
```

---

## Environment Variables

| Variable               | Equivalent Flag | Values                           | Default                      |
| ---------------------- | --------------- | -------------------------------- | ---------------------------- |
| `SIMPLE_CLI_OUTPUT`    | `--output`      | `human`, `json`                  | `human`                      |
| `SIMPLE_CLI_LOG_LEVEL` | `--log-level`   | `debug`, `info`, `warn`, `error` | `info`                       |
| `NO_COLOR`             | `--no-color`    | any non-empty value              | (unset)                      |
| `SIMPLE_CLI_STATE_DIR` | (config key)    | absolute path                    | `$XDG_STATE_HOME/simple-cli` |

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
