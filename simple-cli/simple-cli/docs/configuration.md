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
```

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

---

## Provider Configuration

OAuth + AI chat providers are declared under the `providers:` key. The active provider is controlled by `default_provider`.

### Config File Key Reference

| Key under `providers.<name>` | Type       | Required | Description                                          |
| ----------------------------- | ---------- | -------- | ---------------------------------------------------- |
| `client_id`                   | string     | ✅       | OAuth 2.0 client ID registered with the provider    |
| `device_endpoint`             | string     | ✅       | Device authorization URL (must be `https://`)        |
| `token_endpoint`              | string     | ✅       | Token exchange URL (must be `https://`)              |
| `chat_endpoint`               | string     | ✅       | OpenAI-compatible chat completions URL (`https://`)  |
| `scopes`                      | string[]   | ❌       | OAuth scopes requested during device flow            |
| `default_model`               | string     | ❌       | Model used when `--model` flag is omitted            |

### Top-Level Provider Keys

| Config Key          | Env Variable | Description                                         |
| ------------------- | ------------ | --------------------------------------------------- |
| `default_provider`  | —            | Name of the provider selected when `--provider` is omitted |

### Per-Request Environment Variable Overrides

These env vars override the active provider's config keys at runtime — useful for CI and scripting without modifying the config file.

| Env Variable                       | Overrides field        |
| ----------------------------------- | ---------------------- |
| `SIMPLE_CLI_PROVIDER_CLIENT_ID`     | `client_id`            |
| `SIMPLE_CLI_PROVIDER_DEVICE_ENDPOINT` | `device_endpoint`    |
| `SIMPLE_CLI_PROVIDER_TOKEN_ENDPOINT`  | `token_endpoint`     |
| `SIMPLE_CLI_PROVIDER_CHAT_ENDPOINT`   | `chat_endpoint`      |
| `SIMPLE_CLI_PROVIDER_DEFAULT_MODEL`   | `default_model`      |

### Example config.yaml with a provider

```yaml
output: human
log_level: info
default_provider: my-api
providers:
  my-api:
    client_id: "your-client-id"
    device_endpoint: "https://auth.example.com/device"
    token_endpoint: "https://auth.example.com/token"
    chat_endpoint: "https://api.example.com/v1/chat/completions"
    scopes: ["chat", "offline_access"]
    default_model: "gpt-4o"
```

### Token File

Authenticated tokens are stored in `<config-dir>/tokens.json` (permissions `0600`, never committed — it is in `.gitignore`). The token file uses the following structure:

```json
{
  "providers": {
    "my-api": {
      "provider":      "my-api",
      "access_token":  "<redacted>",
      "refresh_token": "<redacted>",
      "expiry":        "2026-04-01T12:00:00Z"
    }
  }
}
```

