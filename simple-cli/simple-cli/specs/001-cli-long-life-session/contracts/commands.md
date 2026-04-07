# Command Schema Contract

**Version**: 1.0.0 | **Branch**: `001-cli-long-life-session` | **Date**: 2026-03-23

This document defines the stable public interface for all `simple-cli` commands.
Breaking changes to any item here constitute a MAJOR version bump (SemVer).

---

## Global Flags (persistent across all commands)

| Flag          | Short | Type     | Default    | Env Override           | Description                                     |
| ------------- | ----- | -------- | ---------- | ---------------------- | ----------------------------------------------- |
| `--output`    | `-o`  | `string` | `human`    | `SIMPLE_CLI_OUTPUT`    | Output format: `human` or `json`                |
| `--log-level` | —     | `string` | `info`     | `SIMPLE_CLI_LOG_LEVEL` | Log verbosity: `debug`, `info`, `warn`, `error` |
| `--no-color`  | —     | `bool`   | `false`    | `NO_COLOR`             | Suppress ANSI escape codes                      |
| `--quiet`     | `-q`  | `bool`   | `false`    | —                      | Suppress all informational output               |
| `--config`    | —     | `string` | OS default | `SIMPLE_CLI_CONFIG`    | Path to config file                             |
| `--version`   | `-v`  | —        | —          | —                      | Print version and exit                          |
| `--help`      | `-h`  | —        | —          | —                      | Show help                                       |

---

## Root Command

```
simple-cli [flags]
```

**Short**: A cross-platform CLI for managing long-life sessions
**Example**:

```sh
simple-cli --version
simple-cli --output json session list
```

---

## `session` Sub-command Group

```
simple-cli session <verb> [flags]
```

**Short**: Manage long-life sessions

---

### `session start`

```
simple-cli session start [--name <name>] [flags]
```

**Short**: Start a new session
**Long**: Creates a new persistent session. If `--name` is not provided, a name is generated from an adjective-noun pair (e.g., `bold-river`).

| Flag     | Short | Type     | Default        | Description                                                             |
| -------- | ----- | -------- | -------------- | ----------------------------------------------------------------------- |
| `--name` | `-n`  | `string` | auto-generated | Human-readable session name (1–64 chars, `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`) |

**Example**:

```sh
simple-cli session start
simple-cli session start --name my-project
simple-cli session start --name my-project --output json
```

**Success output (human)**:

```
Session 'my-project' started (id: 550e8400-e29b-41d4-a716-446655440000)
```

**Success output (json)** — see [output-schema.md](output-schema.md#session-start).

---

### `session resume`

```
simple-cli session resume [--name <name> | --id <id>] [flags]
```

**Short**: Resume an existing session
**Long**: Loads an existing session by name or ID and sets it as the active context. At least one of `--name` or `--id` must be provided.

| Flag     | Short | Type     | Default | Description  |
| -------- | ----- | -------- | ------- | ------------ |
| `--name` | `-n`  | `string` | —       | Session name |
| `--id`   | —     | `string` | —       | Session UUID |

**Example**:

```sh
simple-cli session resume --name my-project
simple-cli session resume --id 550e8400-e29b-41d4-a716-446655440000
```

---

### `session list`

```
simple-cli session list [--status <status>] [flags]
```

**Short**: List all sessions
**Long**: Lists all sessions, optionally filtered by status. Columns: ID (truncated), Name, Status, Created, Updated.

| Flag       | Short | Type     | Default | Description                                     |
| ---------- | ----- | -------- | ------- | ----------------------------------------------- |
| `--status` | `-s`  | `string` | all     | Filter by status: `active`, `paused`, `stopped` |

**Example**:

```sh
simple-cli session list
simple-cli session list --status active --output json
```

---

### `session stop`

```
simple-cli session stop [--name <name> | --id <id>] [flags]
```

**Short**: Stop a session
**Long**: Sets session status to `stopped`. The session state is retained on disk for inspection but the session is no longer active.

| Flag     | Short | Type     | Default | Description  |
| -------- | ----- | -------- | ------- | ------------ |
| `--name` | `-n`  | `string` | —       | Session name |
| `--id`   | —     | `string` | —       | Session UUID |

**Example**:

```sh
simple-cli session stop --name my-project
```

---

### `session reset`

```
simple-cli session reset [--name <name> | --id <id>] [--force] [flags]
```

**Short**: Reset (delete + recreate) a session
**Long**: Deletes the session and creates a new one with the same name and a fresh state. Requires `--force` to skip the confirmation prompt in human mode. In `--output json` mode, `--force` is assumed.

| Flag      | Short | Type     | Default | Description              |
| --------- | ----- | -------- | ------- | ------------------------ |
| `--name`  | `-n`  | `string` | —       | Session name             |
| `--id`    | —     | `string` | —       | Session UUID             |
| `--force` | `-f`  | `bool`   | `false` | Skip confirmation prompt |

**Example**:

```sh
simple-cli session reset --name my-project --force
```

---

## Error Codes (stable contract)

See [exit-codes.md](exit-codes.md) for full exit code table.

| Code String             | Exit Code | Condition                                  |
| ----------------------- | --------- | ------------------------------------------ |
| `SESSION_NOT_FOUND`     | 3         | No session with the given name or ID       |
| `SESSION_NAME_CONFLICT` | 1         | Session with this name already exists      |
| `SESSION_LOCK_TIMEOUT`  | 5         | Could not acquire file lock within timeout |
| `STORE_READ_ONLY`       | 4         | State directory is read-only               |
| `INVALID_ARGUMENT`      | 2         | Flag value fails validation                |
| `INTERNAL_ERROR`        | 1         | Unexpected internal error                  |
