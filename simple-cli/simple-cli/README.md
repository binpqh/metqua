# simple-cli

[![Go](https://img.shields.io/badge/Go-1.22%2B-blue?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com/your-org/simple-cli/releases)

A cross-platform CLI for managing **long-life sessions** that survive terminal restarts.
Designed for developers and AI agent workflows.

---

## Quick Install

```bash
# Linux / macOS
curl -sSL https://github.com/your-org/simple-cli/releases/latest/download/install.sh | bash

# macOS (Homebrew)
brew install your-org/tap/simple-cli

# Windows (PowerShell)
irm https://github.com/your-org/simple-cli/releases/latest/download/install.ps1 | iex
```

Verify: `simple-cli --version`

---

## 3-Command Quick Start

```bash
# 1. Start a session
simple-cli session start --name my-project

# 2. Close your terminal, reopen it, then resume
simple-cli session resume --name my-project

# 3. Stop when done
simple-cli session stop --name my-project
```

---

## Features

| Feature                       | Description                                                                  |
| ----------------------------- | ---------------------------------------------------------------------------- |
| **Session persistence**       | Sessions survive terminal restarts via file-backed storage                   |
| **Cross-platform**            | Linux, macOS, Windows — single static binary, no runtime dependencies        |
| **AI agent JSON output**      | `--output json` (or `SIMPLE_CLI_OUTPUT=json`) produces stable JSON envelopes |
| **Structured exit codes**     | Deterministic exit codes for scripting and error handling                    |
| **Cross-platform installers** | Shell, PowerShell, NSIS, PKG — PATH registered automatically                 |
| **Session lifecycle**         | `start`, `resume`, `list`, `stop`, `reset`                                   |
| **Concurrent-safe**           | File locking (flock / LockFileEx) prevents data corruption                   |

---

## Commands

```
simple-cli session start   [--name <name>]
simple-cli session resume  [--name <name> | --id <id>]
simple-cli session list    [--status active|paused|stopped]
simple-cli session stop    [--name <name> | --id <id>]
simple-cli session reset   [--name <name> | --id <id>] [--force]
```

Global flags: `--output {human|json}`, `--log-level {debug|info|warn|error}`, `--no-color`, `--quiet`

---

## AI Agent Integration

```bash
export SIMPLE_CLI_OUTPUT=json
simple-cli session start --name workflow
# {"status":"ok","data":{"id":"...","name":"workflow","status":"active",...},...}
```

Exit codes: `0` success, `2` invalid args, `3` not found, `4` permission denied, `5` timeout.

See [docs/ai-agent-guide.md](docs/ai-agent-guide.md) for full examples in Bash, PowerShell, and Python.

---

## Development

```bash
# Build
make build

# Test (≥80% coverage)
make test

# Lint
make lint

# Install locally
make install-local
```

See [docs/architecture.md](docs/architecture.md) and [docs/configuration.md](docs/configuration.md).

---

## Contributing

1. Fork the repo and create a feature branch
2. Run `make test` and `make lint` before submitting
3. See [CONTRIBUTING.md](CONTRIBUTING.md) for full guidelines

---

## License

MIT — see [LICENSE](LICENSE).
