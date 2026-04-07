# Exit Code Contract

**Version**: 1.0.0 | **Branch**: `001-cli-long-life-session` | **Date**: 2026-03-23

Exit codes are a **stable, versioned contract**. Any change to the meaning of an
exit code is a MAJOR version bump. New exit codes may be added in MINOR releases.

---

## Exit Code Table

| Code | Name                  | Condition                                                        | Example                                               |
| ---- | --------------------- | ---------------------------------------------------------------- | ----------------------------------------------------- |
| `0`  | Success               | Command completed successfully                                   | `session start` created a new session                 |
| `1`  | General Error         | Unclassified internal error                                      | Unexpected panic (should not happen; indicates a bug) |
| `2`  | Misuse / Invalid Args | Missing required flags, unexpected arguments, flag type mismatch | `session start --name ""`                             |
| `3`  | Not Found             | Requested resource does not exist                                | `session resume --name no-such-session`               |
| `4`  | Permission Denied     | Filesystem or OS permission error                                | State directory not writable                          |
| `5`  | Timeout               | Context deadline exceeded or lock acquisition timeout            | Concurrent lock held too long                         |

---

## Checking Exit Codes in Scripts

### Bash

```bash
simple-cli session resume --name my-session --output json
STATUS=$?

case $STATUS in
  0) echo "Success" ;;
  2) echo "Invalid arguments — check flags" ;;
  3) echo "Session not found — use 'session list' to see available sessions" ;;
  4) echo "Permission denied — check state directory permissions" ;;
  5) echo "Timeout — another process may be holding the session lock" ;;
  *) echo "Unexpected error (code $STATUS)" ;;
esac
```

### PowerShell

```powershell
& simple-cli session resume --name my-session --output json
switch ($LASTEXITCODE) {
    0 { Write-Host "Success" }
    2 { Write-Host "Invalid arguments" }
    3 { Write-Host "Session not found" }
    4 { Write-Host "Permission denied" }
    5 { Write-Host "Timeout" }
    default { Write-Host "Unexpected error: $LASTEXITCODE" }
}
```

### Python (AI agent tool-use)

```python
import subprocess, json, sys

result = subprocess.run(
    ["simple-cli", "session", "resume", "--name", "my-session", "--output", "json"],
    capture_output=True, text=True
)

if result.returncode == 0:
    payload = json.loads(result.stdout)
    session_id = payload["data"]["id"]
elif result.returncode == 3:
    # Session not found — create it
    ...
else:
    error = json.loads(result.stdout) if result.stdout else {}
    raise RuntimeError(f"CLI error {result.returncode}: {error.get('code', 'UNKNOWN')}")
```

---

## Error Code to Exit Code Mapping

| Error Code (JSON `code`)    | Exit Code |
| --------------------------- | --------- |
| `SESSION_NOT_FOUND`         | 3         |
| `SESSION_NAME_CONFLICT`     | 1         |
| `SESSION_LOCK_TIMEOUT`      | 5         |
| `STORE_READ_ONLY`           | 4         |
| `INVALID_ARGUMENT`          | 2         |
| `INTERNAL_ERROR`            | 1         |
| `CONTEXT_CANCELED`          | 1         |
| `CONTEXT_DEADLINE_EXCEEDED` | 5         |
