# Changelog

All notable changes to **simple-cli** are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

#### US1 — Install & Run First Command

- POSIX shell installer (`scripts/install/install.sh`) supporting Linux and macOS
  with idempotent PATH registration in `.bashrc`, `.zshrc`, and `.profile`
- PowerShell installer (`scripts/install/install.ps1`) for Windows with user-scope
  PATH registration via `[Environment]::SetEnvironmentVariable` and `WM_SETTINGCHANGE` broadcast
- NSIS Windows installer (`installer/windows/setup.nsi`) with machine-scope registry PATH entry
- macOS PKG postinstall script (`installer/macos/postinstall`) writing `/etc/paths.d/simple-cli`
- GoReleaser pipeline (`.goreleaser.yml`) producing static binaries for
  `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
  with `.deb`, `.rpm`, `.tar.gz`, and `.zip` archives
- Installation guide (`docs/installation.md`) with per-platform instructions and PATH troubleshooting

#### US2 — Start and Resume a Long-Life Session

- `session start [--name <name>]` — creates a new persistent session; auto-generates name if omitted
- `session resume [--name <name> | --id <id>]` — resumes a session after terminal restart
- `session list [--status <status>]` — lists sessions with optional status filter
- `session stop [--name <name> | --id <id>]` — marks a session as stopped
- `session reset [--name <name> | --id <id>] [--force]` — deletes and recreates a session with fresh state
- File-backed session store (`internal/session/FileStore`) with per-session file locking
  using `flock(2)` on Linux/macOS and `LockFileEx` on Windows
- In-memory fallback store (`internal/session/MemStore`) for read-only environments
- Session state persisted in `$XDG_STATE_HOME/simple-cli/` (Linux/macOS) or `%APPDATA%\simple-cli\` (Windows)

#### US3 — AI Agent Machine-Readable Output

- `--output json` flag (and `SIMPLE_CLI_OUTPUT` env var) for stable JSON envelope output
- JSON success envelope: `{"status":"ok","data":{...},"meta":{"version":"...","duration_ms":N,"command":"..."}}`
- JSON error envelope on stderr: `{"status":"error","code":"...","message":"...","hint":"...","meta":{...}}`
- Deterministic exit codes: 0 (success), 2 (invalid args), 3 (not found), 4 (permission denied), 5 (timeout)
- Structured JSON logging on stderr via `log/slog` when `--output json`
- `--no-color` flag and `NO_COLOR` env var support for ANSI-free output
- `--quiet` flag to suppress informational stdout writes
- AI agent integration guide (`docs/ai-agent-guide.md`) with Bash, PowerShell, and Python examples

#### US4 — Developer Local Build & Test Workflow

- `make build` — static `CGO_ENABLED=0` binary in `dist/`
- `make test` — unit tests with ≥80% coverage gate
- `make lint` — `golangci-lint` with `errcheck`, `gosec`, `revive`, `staticcheck`
- `make install-local` — copies binary to `$GOPATH/bin`
- Unit tests for all `internal/` packages (>80% coverage on session, output, security, signals, config)
- Integration tests (`tests/integration/`) covering the full session lifecycle black-box
- Architecture docs (`docs/architecture.md`)
- Configuration reference (`docs/configuration.md`)

#### US5 — Sandbox Installer Testing

- Docker Compose sandbox (`tests/sandbox/docker-compose.yml`) with Ubuntu 22.04, Debian 12, Alpine 3.19 services
- Dockerfile for each Linux target verifying the static binary runs inside the container
- Windows Server Core Dockerfile (requires Windows CI runner)
- `make test-sandbox` target with `SKIP_WINDOWS_SANDBOX` guard
