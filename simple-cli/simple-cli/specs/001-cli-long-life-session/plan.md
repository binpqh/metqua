# Implementation Plan: CLI Long-Life Session Application

**Branch**: `001-cli-long-life-session` | **Date**: 2026-03-23 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-cli-long-life-session/spec.md`

---

## Summary

Build `simple-cli` — a cross-platform, long-life session CLI application written in Go 1.22+ using Cobra + Viper. The application maintains persistent sessions across terminal restarts, ships first-class installers for Windows (NSIS + PowerShell), Linux (shell + deb/rpm + profile.d), and macOS (Homebrew + PKG), and registers its PATH automatically on every platform. All commands expose `--output json` for AI agent interoperability. The binary is statically linked, under 20 MB, and starts in < 150 ms.

---

## Technical Context

**Language/Version**: Go 1.22+ (minimum); CI tests on current stable and previous stable release
**Primary Dependencies**:

- `github.com/spf13/cobra` v1.8+ — CLI framework
- `github.com/spf13/viper` v1.18+ — configuration management
- `github.com/google/uuid` — session ID generation
- `github.com/stretchr/testify` — test assertions
- `github.com/testcontainers/testcontainers-go` — sandbox container tests
- `golang.org/x/sys` — platform-specific PATH/registry operations

**Storage**: File-backed session store at `$XDG_STATE_HOME/simple-cli/` (Linux/macOS) or `%APPDATA%\simple-cli\` (Windows); in-memory fallback when filesystem is unavailable
**Testing**: `testing` stdlib + `testify` for assertions + `gomock` for interfaces + Docker via `testcontainers-go` for sandbox tests
**Target Platform**: Linux (amd64, arm64), macOS (amd64/arm64), Windows (amd64) — static binaries
**Project Type**: CLI application (single binary, installable)
**Performance Goals**: Cold startup < 150 ms; binary size < 20 MB post-strip
**Constraints**: `CGO_ENABLED=0` static builds; ≥ 80% unit test coverage; `log/slog` only; no init() in library packages
**Scale/Scope**: Single-user CLI; multi-session concurrency only within one user's context; file-lock concurrency safe

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design._

| #    | Principle                           | Status  | Notes                                                                                                       |
| ---- | ----------------------------------- | ------- | ----------------------------------------------------------------------------------------------------------- |
| I    | Library-First Architecture          | ✅ PASS | `internal/session`, `internal/config`, `internal/security`, `internal/output` all independent packages      |
| II   | Idiomatic Go                        | ✅ PASS | `gofmt`, errors wrapped with `%w`, `context.Context` on all I/O, no `init()` in lib packages                |
| III  | Cross-Platform Installer & ENV PATH | ✅ PASS | NSIS + PS1 (Win), shell + deb/rpm + profile.d (Linux), Homebrew + PKG (macOS)                               |
| IV   | AI Agent Interoperability           | ✅ PASS | `--output json`, stable exit codes 0–5, `SIMPLE_CLI_OUTPUT=json` env var, `--no-color`, `--quiet`           |
| V    | Test-First                          | ✅ PASS | Unit tests alongside implementation, integration tests in `tests/integration/`, sandbox in `tests/sandbox/` |
| VI   | Observability & Structured Logging  | ✅ PASS | `log/slog`, stderr-only logs, redaction helpers in `internal/security/`                                     |
| VII  | Documentation Standards             | ✅ PASS | GoDoc on all exports, `docs/` with 6 required files, Cobra `Short`+`Long`+`Example` on every command        |
| VIII | Simplicity & Lightweight            | ✅ PASS | CGO_ENABLED=0, `-ldflags="-s -w"`, static binary, < 20 MB, < 150 ms startup                                 |
| IX   | Robustness & Error Resilience       | ✅ PASS | Context-aware I/O, graceful SIGTERM shutdown ≤ 5s, file-lock concurrency safety                             |
| X    | Versioning & Maintainability        | ✅ PASS | SemVer, CHANGELOG.md, `golangci-lint` blocks merge, version injected via ldflags                            |

**Post-Phase-1 re-check**: All 10 principles verified against data-model and contracts — no violations found.

---

## Project Structure

### Documentation (this feature)

```text
specs/001-cli-long-life-session/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── commands.md      # Cobra command schema and flag definitions
│   ├── output-schema.md # JSON output schema for --output json mode
│   └── exit-codes.md    # Documented exit code contract
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
simple-cli/
├── cmd/
│   ├── root.go              # Root command, global flags, version injection
│   └── session/
│       ├── session.go       # `session` sub-command group
│       ├── start.go         # `session start`
│       ├── resume.go        # `session resume`
│       ├── list.go          # `session list`
│       ├── stop.go          # `session stop`
│       └── reset.go         # `session reset`
├── internal/
│   ├── config/
│   │   ├── config.go        # Viper-backed config loader; XDG paths
│   │   └── config_test.go
│   ├── session/
│   │   ├── session.go       # Session entity + status enum
│   │   ├── store.go         # SessionStore interface
│   │   ├── filestore.go     # File-backed store (JSON + file lock)
│   │   ├── memstore.go      # In-memory fallback store
│   │   ├── lock.go          # Cross-platform file locking
│   │   └── *_test.go
│   ├── output/
│   │   ├── output.go        # --output json / human formatter
│   │   ├── writer.go        # Separated stdout/stderr writer
│   │   └── *_test.go
│   ├── security/
│   │   ├── redact.go        # Token/password redaction helpers
│   │   └── redact_test.go
│   └── signals/
│       ├── signals.go       # SIGINT/SIGTERM graceful shutdown handler
│       └── signals_test.go
├── pkg/
│   └── version/
│       └── version.go       # Version string (injected via ldflags at build time)
├── scripts/
│   └── install/
│       ├── install.sh       # POSIX install script (Linux + macOS)
│       └── install.ps1      # PowerShell 5.1+ install script (Windows)
├── installer/
│   ├── windows/
│   │   └── setup.nsi        # NSIS installer definition
│   └── macos/
│       └── postinstall      # PKG postinstall PATH registration
├── tests/
│   ├── integration/
│   │   ├── session_test.go  # Black-box session command tests
│   │   └── output_test.go   # --output json contract tests
│   └── sandbox/
│       ├── docker-compose.yml
│       └── docker/
│           ├── ubuntu-22.04/Dockerfile
│           ├── debian-12/Dockerfile
│           ├── alpine-3.19/Dockerfile
│           └── windows-ltsc/Dockerfile
├── docs/
│   ├── installation.md
│   ├── quickstart.md
│   ├── configuration.md
│   ├── architecture.md
│   ├── ai-agent-guide.md
│   └── assets/
├── .goreleaser.yml          # Cross-platform build + package artifacts
├── .golangci.yml            # Linter configuration
├── Makefile                 # Developer workflow: build, test, lint, install-local
├── go.mod
├── go.sum
├── CHANGELOG.md
└── README.md
```

**Structure Decision**: Single-project layout with `cmd/` for Cobra entry points,
`internal/` for domain logic (not importable externally), `pkg/` for stable public API,
`scripts/install/` for install artifacts, and `tests/` for integration and sandbox tests.
Follows standard Go project conventions and satisfies Principle I (Library-First).

---

## Complexity Tracking

No constitution violations requiring justification. All complexity is directly demanded by the stated requirements (cross-platform installers, session persistence, AI agent I/O, sandbox tests).
