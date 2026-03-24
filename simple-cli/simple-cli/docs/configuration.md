# Configuration Reference

**simple-cli** honours the following precedence chain for all settings:

> **CLI flags** → **Environment variables** → **Config file** → **Built-in defaults**

---

## All Configuration Options

| Flag            | Env Variable           | Config File Key | Type   | Default                                 | Description                                                  |
| --------------- | ---------------------- | --------------- | ------ | --------------------------------------- | ------------------------------------------------------------ |
| `--output`      | `SIMPLE_CLI_OUTPUT`    | `output`        | string | `human`                                 | Output format. `human` for terminal, `json` for agents       |
| `--log-level`   | `SIMPLE_CLI_LOG_LEVEL` | `log_level`     | string | `info`                                  | Log verbosity. `debug`, `info`, `warn`, `error`              |
| `--no-color`    | `NO_COLOR`             | `no_color`      | bool   | `false`                                 | Suppress ANSI colour codes (honours the `NO_COLOR` standard) |
| `--quiet`, `-q` | —                      | `quiet`         | bool   | `false`                                 | Suppress all informational output to stdout                  |
| `--config`      | —                      | —               | string | See [Config File](#config-file)         | Explicit path to config file                                 |
| —               | `SIMPLE_CLI_STATE_DIR` | `state_dir`     | string | See [State Directory](#state-directory) | Directory where session state is persisted                   |

---

## Config File

The config file is a YAML file. Default locations (searched in order):

| OS      | Default path                                                                           |
| ------- | -------------------------------------------------------------------------------------- |
| Linux   | `$XDG_CONFIG_HOME/simple-cli/config.yaml` (default `~/.config/simple-cli/config.yaml`) |
| macOS   | `$XDG_CONFIG_HOME/simple-cli/config.yaml` (default `~/.config/simple-cli/config.yaml`) |
| Windows | `%APPDATA%\simple-cli\config.yaml`                                                     |

Override with `--config /path/to/config.yaml`.

**Example config.yaml**:

```yaml
output: json
log_level: warn
no_color: false
quiet: false
state_dir: /custom/state/dir
```

---

## State Directory

Long-life session data is persisted here. Default locations:

| OS      | Default path                                                       |
| ------- | ------------------------------------------------------------------ |
| Linux   | `$XDG_STATE_HOME/simple-cli` (default `~/.local/state/simple-cli`) |
| macOS   | `$XDG_STATE_HOME/simple-cli` (default `~/.local/state/simple-cli`) |
| Windows | `%APPDATA%\simple-cli`                                             |

Override via `SIMPLE_CLI_STATE_DIR` env var or `state_dir` config file key.

---

## Output Modes

### `--output human` (default)

Plain text suitable for reading in a terminal. Includes ANSI colour codes unless `--no-color` is set.

### `--output json`

JSON Lines written to stdout (success) or stderr (error). Designed for AI agent and scripted consumption. See [docs/ai-agent-guide.md](ai-agent-guide.md) for full schema.

---

## Log Levels

| Level   | What is logged                                             |
| ------- | ---------------------------------------------------------- |
| `debug` | All log entries including file lock polling, store ops     |
| `info`  | Key lifecycle events (session created, resumed, etc.)      |
| `warn`  | Recoverable issues (e.g., falling back to in-memory store) |
| `error` | Unrecoverable errors (never just exit silently)            |

All log output goes to **stderr** regardless of `--output` setting.

---

## Precedence Example

Given the following setup:

- Config file: `log_level: warn`
- Env var: `SIMPLE_CLI_LOG_LEVEL=debug`
- Flag: `--log-level info`

**Result**: `log_level = info` (flag wins over env var over config file).
