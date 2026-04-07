# JSON Output Schema Contract

**Version**: 1.0.0 | **Branch**: `001-cli-long-life-session` | **Date**: 2026-03-23

All schemas below are the stable contract for `--output json` mode.
The `code` field in error responses and all top-level field names are versioned;
`message` and `hint` text MAY change between patch releases.

---

## Envelope Schema

All responses wrap command-specific data in a common envelope.

### Success Envelope

```json
{
  "status": "ok",
  "data": <command-specific object or array>,
  "meta": {
    "version": "1.2.3",
    "duration_ms": 42,
    "command": "session start"
  }
}
```

### Error Envelope

```json
{
  "status": "error",
  "code": "SESSION_NOT_FOUND",
  "message": "session 'my-project' not found",
  "hint": "Use 'simple-cli session list' to see available sessions.",
  "meta": {
    "version": "1.2.3",
    "duration_ms": 5,
    "command": "session resume"
  }
}
```

Fields:

| Field              | Type                        | Stable                   | Description                            |
| ------------------ | --------------------------- | ------------------------ | -------------------------------------- |
| `status`           | `"ok"` \| `"error"`         | ✅                       | Top-level discriminator                |
| `data`             | object \| array             | ✅ shape; ✅ field names | Command payload (success only)         |
| `code`             | SCREAMING_SNAKE_CASE string | ✅                       | Machine-readable error discriminator   |
| `message`          | string                      | ❌ (may change)          | Human-readable error description       |
| `hint`             | string?                     | ❌ (may change)          | Remediation suggestion                 |
| `meta.version`     | semver string               | ✅                       | CLI version that produced the response |
| `meta.duration_ms` | integer                     | ✅                       | Wall-clock ms for the command          |
| `meta.command`     | string                      | ✅                       | Canonical command name                 |

---

## Command-Specific Data Schemas

### `session start`

```json
{
  "status": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-project",
    "status": "active",
    "created_at": "2026-03-23T10:00:00Z",
    "updated_at": "2026-03-23T10:00:00Z",
    "state": {}
  },
  "meta": { "version": "1.0.0", "duration_ms": 12, "command": "session start" }
}
```

### `session resume`

```json
{
  "status": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-project",
    "status": "active",
    "created_at": "2026-03-23T10:00:00Z",
    "updated_at": "2026-03-23T10:05:00Z",
    "state": { "last_task": "build" }
  },
  "meta": { "version": "1.0.0", "duration_ms": 8, "command": "session resume" }
}
```

### `session list`

```json
{
  "status": "ok",
  "data": {
    "sessions": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "my-project",
        "status": "active",
        "created_at": "2026-03-23T10:00:00Z",
        "updated_at": "2026-03-23T10:05:00Z"
      }
    ],
    "total": 1
  },
  "meta": { "version": "1.0.0", "duration_ms": 4, "command": "session list" }
}
```

Note: `state` is omitted from list output for performance. Use `session resume` to
load full state.

### `session stop`

```json
{
  "status": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-project",
    "status": "stopped",
    "updated_at": "2026-03-23T10:10:00Z"
  },
  "meta": { "version": "1.0.0", "duration_ms": 6, "command": "session stop" }
}
```

### `session reset`

```json
{
  "status": "ok",
  "data": {
    "old_id": "550e8400-e29b-41d4-a716-446655440000",
    "id": "7f1e9a2b-dead-beef-cafe-123456789abc",
    "name": "my-project",
    "status": "active",
    "created_at": "2026-03-23T10:15:00Z",
    "updated_at": "2026-03-23T10:15:00Z",
    "state": {}
  },
  "meta": { "version": "1.0.0", "duration_ms": 15, "command": "session reset" }
}
```

---

## Stderr Progress Format (JSON-Lines)

When `--output json` and a long-running operation is in progress, progress events
are emitted to **stderr** as newline-delimited JSON:

```json
{"level":"info","time":"2026-03-23T10:00:01Z","msg":"Acquiring session lock","step":1,"total":3}
{"level":"info","time":"2026-03-23T10:00:01Z","msg":"Writing session state","step":2,"total":3}
{"level":"info","time":"2026-03-23T10:00:01Z","msg":"Updating index","step":3,"total":3}
```

Fields:

| Field   | Type    | Description                         |
| ------- | ------- | ----------------------------------- |
| `level` | string  | `debug`, `info`, `warn`, `error`    |
| `time`  | RFC3339 | Timestamp                           |
| `msg`   | string  | Human-readable step description     |
| `step`  | int?    | Current step number (if applicable) |
| `total` | int?    | Total steps (if applicable)         |

---

## AI Agent Invocation Pattern

```sh
# Recommended invocation for AI agents
SIMPLE_CLI_OUTPUT=json simple-cli session start --name my-session 2>progress.jsonl

# Check exit code first, then parse stdout
if [ $? -eq 0 ]; then
  SESSION_ID=$(cat result.json | jq -r '.data.id')
fi
```

See [exit-codes.md](exit-codes.md) for the full exit code table.
