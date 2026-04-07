# Feature Specification: CLI Long-Life Session Application

**Feature Branch**: `001-cli-long-life-session`  
**Created**: 2026-03-23  
**Status**: Draft  
**Input**: Build a template for CLI application long-life session with cross-platform installer, Go build system, AI agent interactivity, documentation, and testing.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Install & Run First Command (Priority: P1)

A developer downloads the `simple-cli` binary or runs an install script on their OS
(Windows, Linux, or macOS). After running the installer, `simple-cli` is on their `PATH`
without requiring a shell restart, and they can immediately execute `simple-cli --version`.

**Why this priority**: Without a working installer and PATH registration, no other feature
is usable. This is the absolute entry point for every user.

**Independent Test**: Run installer on a clean OS image, open a new shell, execute
`simple-cli --version` → binary is found and returns the version string.

**Acceptance Scenarios**:

1. **Given** a clean Windows 10/11 machine, **When** `install.ps1` is executed as a user (no elevation), **Then** `simple-cli` is added to the user-scope PATH and `simple-cli --version` succeeds in a new PowerShell session.
2. **Given** a clean Ubuntu 22.04 machine, **When** `install.sh` is executed, **Then** binary is placed in `~/.local/bin`, `~/.bashrc` is updated, and `simple-cli --version` succeeds in a new bash session.
3. **Given** a macOS 13+ machine with Homebrew, **When** `brew install simple-cli`, **Then** binary is on PATH and `simple-cli --version` succeeds.
4. **Given** any installer, **When** `simple-cli` is already installed and PATH already contains it, **Then** the installer detects idempotency and does not duplicate PATH entries.

---

### User Story 2 — Start and Resume a Long-Life Session (Priority: P1)

A user starts a persistent session (`simple-cli session start`), performs work across
multiple commands, closes the terminal, reopens it, and seamlessly resumes the session
(`simple-cli session resume`). Session state persists across shell restarts.

**Why this priority**: The "long-life session" is the core value proposition of this CLI.
Without it, the application has no distinguishing feature.

**Independent Test**: `simple-cli session start`, set a key, close terminal, reopen terminal,
`simple-cli session resume`, retrieve the key → value is intact.

**Acceptance Scenarios**:

1. **Given** no active session, **When** `simple-cli session start`, **Then** a new session is created, a session ID is printed, and session state file is written to `$XDG_STATE_HOME/simple-cli/` (Linux/macOS) or `%APPDATA%\simple-cli\` (Windows).
2. **Given** an active session, **When** `simple-cli session resume [--id <id>]`, **Then** session state is loaded and subsequent commands operate within that session context.
3. **Given** an active session, **When** SIGTERM/Ctrl-C is received, **Then** session state is flushed to disk and the process exits cleanly within 5 seconds.
4. **Given** two concurrent terminal sessions, **When** both attempt to write session state, **Then** file locking prevents corruption and the second write waits or returns a clear conflict error.

---

### User Story 3 — AI Agent Machine-Readable Output (Priority: P2)

An AI agent (LLM tool-use or automated script) invokes `simple-cli` with `--output json`
and parses the structured JSON response to drive further automation. All errors appear on
stderr as JSON-Lines; all payload data appears on stdout.

**Why this priority**: AI agent interoperability is a first-class goal and enables the
automation use-cases that justify the "long-life session" design.

**Independent Test**: `simple-cli session list --output json` returns valid JSON array on
stdout; any error from `simple-cli does-not-exist --output json` returns JSON on stderr with
exit code ≠ 0.

**Acceptance Scenarios**:

1. **Given** any valid command, **When** `--output json` is passed, **Then** stdout contains a valid JSON object/array with consistent schema.
2. **Given** any error condition, **When** `--output json` is passed, **Then** stderr contains `{"level":"error","msg":"...","code":<int>}` JSON-Line and exit code matches documented exit code table.
3. **Given** `SIMPLE_CLI_OUTPUT=json` env var, **When** any command is run without `--output`, **Then** output mode is JSON as if `--output json` was passed.
4. **Given** `--quiet` flag, **When** any command, **Then** only payload data (or nothing) appears on stdout; all informational messages suppressed.
5. **Given** `--no-color` or `NO_COLOR=1`, **When** any command in human mode, **Then** zero ANSI escape sequences in output.

---

### User Story 4 — Developer Local Build & Test Workflow (Priority: P2)

A contributor clones the repository, runs `make build`, `make test`, and `make lint`
successfully on all three platforms, and can install locally via `make install-local`
for immediate dogfooding.

**Why this priority**: Without a reliable developer workflow, contributions stall and the
project cannot be maintained at high quality.

**Independent Test**: On a fresh clone with Go 1.22+ and `golangci-lint` installed,
`make test` passes with ≥ 80% coverage and `make lint` reports zero issues.

**Acceptance Scenarios**:

1. **Given** Go 1.22+ installed, **When** `make build`, **Then** static binary produced in `dist/` with `CGO_ENABLED=0`, size < 20 MB.
2. **Given** `make test`, **Then** unit + integration tests pass, coverage ≥ 80% on `internal/` and `pkg/`.
3. **Given** `make lint`, **Then** `golangci-lint` reports zero issues including `errcheck`, `gosec`, `revive`.
4. **Given** `make install-local`, **Then** binary is copied to a local `bin/` on PATH for dogfooding.

---

### User Story 5 — Sandbox Installer Testing (Priority: P3)

A CI pipeline spins up clean Docker containers for Ubuntu, Debian, Alpine, and Windows
Server Core, executes the respective installer, and validates that `simple-cli --version`
runs successfully inside the container.

**Why this priority**: Sandbox tests catch regression in installer logic across OS
differences without requiring physical machines.

**Independent Test**: `docker compose -f tests/sandbox/docker-compose.yml up --abort-on-container-exit`
exits 0 with all container tests passing.

**Acceptance Scenarios**:

1. **Given** Ubuntu 22.04 container, **When** `install.sh` runs, **Then** `simple-cli --version` succeeds.
2. **Given** Alpine 3.19 container, **When** `install.sh` runs, **Then** binary identified as statically linked and `simple-cli --version` succeeds.
3. **Given** Windows Server Core container on CI runner, **When** `install.ps1` runs, **Then** PATH updated and `simple-cli.exe --version` succeeds.

---

### Edge Cases

- What happens when `PATH` already contains the install directory? (idempotent — no duplicate entry appended)
- What happens when the session state file is corrupt / partially-written? (error reported with file path; session can be reset with `simple-cli session reset`)
- What happens when HOME / APPDATA is unset or read-only? (falls back to in-memory session with a clear warning on stderr)
- What happens when `--output json` is combined with a command that produces streaming output? (JSON-Lines on stderr; summary JSON on stdout at end)
- What happens when binary size exceeds 20 MB? (CI gate fails; build must strip symbols)

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST build a statically-linked Go binary (CGO_ENABLED=0) targeting linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 via `goreleaser`.
- **FR-002**: Installer MUST auto-register `simple-cli` on the system PATH for Windows (user scope + machine scope with elevation), Linux (bashrc/zshrc/profile + /etc/profile.d), and macOS (Homebrew + /etc/paths.d).
- **FR-003**: System MUST provide `simple-cli session start | resume | list | stop | reset` commands for managing persistent long-life sessions.
- **FR-004**: Session state MUST be persisted to `$XDG_STATE_HOME/simple-cli/` (Linux/macOS) or `%APPDATA%\simple-cli\` (Windows) with file locking against concurrent writes.
- **FR-005**: Every command MUST support `--output json` (or `SIMPLE_CLI_OUTPUT=json`) for machine-readable structured output on stdout.
- **FR-006**: Errors MUST always be emitted to stderr; stdout MUST contain only payload data.
- **FR-007**: Exit codes MUST be deterministic: `0` success, `1` general error, `2` misuse/invalid args, `3` not found, `4` permission denied, `5` timeout.
- **FR-008**: System MUST support `--log-level` (debug/info/warn/error) and `SIMPLE_CLI_LOG_LEVEL` env var; all logging via `log/slog` package.
- **FR-009**: Long-running commands MUST implement graceful shutdown on SIGINT/SIGTERM, flushing state and exiting within 5 seconds.
- **FR-010**: Documentation MUST include `docs/installation.md`, `docs/quickstart.md`, `docs/configuration.md`, `docs/architecture.md`, `docs/ai-agent-guide.md`, and `CHANGELOG.md`.
- **FR-011**: Unit tests MUST achieve ≥ 80% line coverage on `internal/` and `pkg/`; coverage gate enforced in CI.
- **FR-012**: Sandbox tests MUST validate installer behavior on Ubuntu, Debian, Alpine, and Windows Server Core via Docker Compose in `tests/sandbox/`.
- **FR-013**: Binary MUST remain under 20 MB after `-ldflags="-s -w"` stripping; startup time < 150 ms.
- **FR-014**: Version string MUST be injected at build time: `-ldflags "-X main.Version=$(git describe --tags)"`.
- **FR-015**: CI MUST run `go vet`, `staticcheck`, and `golangci-lint` on every PR; failures block merge.

### Key Entities

- **Session**: Represents a named, persistent user session. Fields: `id` (UUID), `name` (string), `createdAt` (time), `updatedAt` (time), `state` (map[string]any), `status` (enum: active/paused/stopped).
- **SessionStore**: Abstraction over session persistence (file-backed default; in-memory fallback). Provides thread-safe CRUD with file locking.
- **Command**: A Cobra command node. Each command carries `Short`, `Long`, and at least one usage `Example`. Global flags: `--output`, `--log-level`, `--no-color`, `--quiet`.
- **Installer**: Platform-specific install artifact (NSIS `.exe`, PowerShell `.ps1`, shell `.sh`, Homebrew formula, PKG). Responsible for binary placement and PATH registration.
- **Config**: Viper-backed configuration. Precedence: CLI flags > env vars > config file > defaults. Config file path: `$XDG_CONFIG_HOME/simple-cli/config.yaml`.

### Non-Functional Requirements

- **NFR-001**: Binary must be statically linked (CGO_ENABLED=0) for drop-in deployment.
- **NFR-002**: Binary size < 20 MB post-strip; cold startup < 150 ms.
- **NFR-003**: All I/O operations must honor context.Context cancellation.
- **NFR-004**: Sensitive values (tokens, passwords) MUST NEVER appear in log output; redaction helpers in `internal/security/`.
- **NFR-005**: Configuration parsing errors MUST report offending file path and line number.
- **NFR-006**: All log output uses `log/slog`; third-party logging libraries are prohibited.
