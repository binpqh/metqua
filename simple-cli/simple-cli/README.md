# simple-cli

[![Go](https://img.shields.io/badge/Go-1.22%2B-blue?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com/binpqh/simple-cli/releases)

A **cross-platform CLI template** that stays alive as a long-running process until the device shuts down.
Designed to be customised into any CLI application, with built-in support for AI agent workflows.

---

## Quick Install

```bash
# Linux / macOS
curl -sSL https://github.com/binpqh/simple-cli/releases/latest/download/install.sh | bash

# macOS (Homebrew)
brew install binpqh/tap/simple-cli

# Windows (PowerShell)
irm https://github.com/binpqh/simple-cli/releases/latest/download/install.ps1 | iex
```

Verify: `simple-cli --version`

---

## Quick Start

```bash
# Start the long-running daemon (blocks until shutdown signal)
simple-cli run

# Run with JSON output for AI agents / scripts
simple-cli --output json run
# {"status":"ok","data":{"status":"stopped","uptime_ms":5123},"meta":{...}}
```

Stop with `Ctrl+C` or `SIGTERM`. The process exits cleanly within 5 seconds.

---

## Features

| Feature                       | Description                                                                  |
| ----------------------------- | ---------------------------------------------------------------------------- |
| **Daemon lifecycle**          | `run` command blocks until SIGINT/SIGTERM — survives until device shutdown   |
| **Template extensibility**    | Add sub-commands in `cmd/`, implement logic in `internal/`                   |
| **Cross-platform**            | Linux, macOS, Windows — single static binary, no runtime dependencies        |
| **AI agent JSON output**      | `--output json` (or `SIMPLE_CLI_OUTPUT=json`) produces stable JSON envelopes |
| **Structured exit codes**     | Deterministic exit codes for scripting and error handling                    |
| **Cross-platform installers** | Shell, PowerShell, NSIS, PKG — PATH registered automatically                 |

---

## Commands

```
simple-cli run      # Start the long-running daemon process
simple-cli example  # Example sub-command (safe to delete when customising)
```

Global flags: `--output {human|json}`, `--log-level {debug|info|warn|error}`, `--no-color`, `--quiet`

---

## AI Agent Integration

```bash
export SIMPLE_CLI_OUTPUT=json
simple-cli run
# blocks; on shutdown:
# {"status":"ok","data":{"status":"stopped","uptime_ms":12345},"meta":{"command":"run",...}}
```

Exit codes: `0` success, `1` general error, `2` invalid args.

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
