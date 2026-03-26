<!--
## Sync Impact Report
- **Version change**: 1.0.0 → 2.0.0
- **Modified principles**:
  - Principle I: "Library-First Architecture" — removed "session management" from capability list;
    added "daemon/process lifecycle management" as a first-class capability.
  - Principle IV: "AI Agent Interoperability" — removed session-state persistence bullet;
    added template extensibility guidance.
  - Principle IX: "Robustness & Error Resilience" — replaced "Long-running session commands"
    with "The bundled `run` daemon command" to reflect new project focus.
- **Added sections**: None
- **Removed sections**: None (session concept retired from governance; no dedicated section existed)
- **Templates updated**:
  - ✅ `.specify/memory/constitution.md` — updated (canonical source)
  - ✅ `.github/constitution.md` — updated (this file; mirror copy)
  - ✅ `.specify/templates/plan-template.md` — verified; no session-specific gates; no update needed
  - ✅ `.specify/templates/spec-template.md` — verified; generic; no update needed
  - ✅ `.specify/templates/tasks-template.md` — verified; generic; no update needed
  - ✅ Technology Stack section — updated sub-command pattern example
- **Follow-up TODOs**: None — all fields resolved
-->

# simple-cli Constitution

## Core Principles

### I. Library-First Architecture

Every discrete capability (daemon/process lifecycle management, command dispatch, installer logic,
AI agent I/O) MUST be implemented as an independently importable Go package under `internal/` or
`pkg/`. New domain capabilities added when customising the template follow the same rule.
No cross-package circular imports are permitted. Each package MUST expose a clean public API,
be independently testable with no global state, and carry a package-level doc comment explaining
its single responsibility. Organizational-only packages (packages with no exported symbols) are
prohibited.

### II. Idiomatic Go (NON-NEGOTIABLE)

All code MUST conform to the Go Language Specification and `gofmt` / `goimports` formatting
without exception. The following rules apply at all times:

- Errors MUST be returned, not ignored; `_ = err` is forbidden except in generated code.
- Error values MUST wrap context using `fmt.Errorf("operation: %w", err)` and MUST NOT
  leak stack traces to end users — only to debug/log output.
- Functions MUST have a single, clear responsibility; cyclomatic complexity > 10 requires
  justification in a code comment.
- Exported identifiers MUST carry GoDoc comments; unexported ones SHOULD where non-trivial.
- No `init()` functions except for flag registration in `cmd/` packages.
- `panic` is reserved for programmer errors (invariant violations) and MUST NOT appear in
  library code or on any user-triggered code path.
- `context.Context` MUST be the first parameter of every function that performs I/O,
  long computation, or interacts with external systems.

### III. Cross-Platform Installer & ENV PATH Auto-Registration

The project MUST ship a first-class installer for every supported platform. Installers are
production artifacts equal in quality to the binary itself.

**Windows**

- Primary: NSIS-based `.exe` installer generated via CI.
- Fallback: PowerShell install script (`install.ps1`) usable without admin rights via
  User-scope `PATH` registration in `HKCU\Environment`.
- With elevation: Machine-scope `PATH` in `HKLM\SYSTEM\CurrentControlSet\Control\Session
Manager\Environment`, broadcast `WM_SETTINGCHANGE`.

**Linux**

- Primary: Shell script (`install.sh`) that copies the binary to `/usr/local/bin` (with sudo)
  or `~/.local/bin` (without), then appends to `~/.bashrc`, `~/.zshrc`, and
  `~/.profile` as applicable.
- Packaged: `.deb` and `.rpm` packages generated via `goreleaser`; these MUST declare the
  install path and set `PATH` via `/etc/profile.d/simple-cli.sh`.
- Homebrew tap: `homebrew-simple-cli` tap with a formula for macOS and Linux.

**macOS**

- Primary: Homebrew formula via official tap (preferred distribution path).
- Fallback: Shell script (`install.sh`) identical in logic to the Linux script, respecting
  `/usr/local/bin` on Intel and `/opt/homebrew/bin` on Apple Silicon.
- PKG installer: `.pkg` via `pkgbuild`/`productbuild` for enterprise distribution; uses
  a `postinstall` script to add `/usr/local/bin` to `/etc/paths.d/simple-cli`.
- All installers MUST validate PATH registration by executing `simple-cli --version` in a
  new shell after installation and report success or remediation steps on failure.

### IV. AI Agent Interoperability

The CLI MUST be designed as a first-class participant in AI agent workflows.

- Every command MUST support `--output json` (or `SIMPLE_CLI_OUTPUT=json`) to emit
  machine-readable structured output on stdout; human-readable text is the default.
- Errors MUST always be emitted to stderr; stdout MUST contain only payload data.
- Exit codes MUST be deterministic and documented: `0` success, `1` general error,
  `2` misuse/invalid args, `3` resource not found, `4` permission denied, `5` timeout.
- A `--no-color` / `NO_COLOR=1` flag MUST suppress all ANSI escape codes.
- A `--quiet` flag MUST suppress all informational output, leaving only data on stdout.
- Long-running progress MUST be emitted to stderr as JSON-Lines when
  `SIMPLE_CLI_OUTPUT=json`, enabling agent log parsing without stdout pollution.
- This project is a **customisable template**. Adding a new sub-command constitutes a new
  "capability" for the purposes of Principle I; each new command MUST be independently
  importable, testable, and documented before merging.

### V. Test-First (NON-NEGOTIABLE)

Tests MUST be written before or alongside implementation — never after the fact.

- **Unit tests**: Every exported function and every non-trivial unexported function MUST have
  a corresponding `_test.go` file in the same package. Table-driven tests are preferred.
- **Integration tests**: Placed in `tests/integration/`; MUST exercise the CLI binary as a
  black box (`exec.Command`) against a real or containerized environment.
- **Sandbox / container tests**: A `tests/sandbox/` directory MUST contain Docker Compose
  definitions and test scripts that exercise installer behavior on clean OS images
  (Ubuntu, Debian, Alpine, Windows Server Core, macOS via CI runner).
- Coverage gate: `go test ./...` MUST maintain ≥ 80% line coverage on `internal/` and
  `pkg/` packages; CI enforces this gate and MUST NOT be bypassed.
- No test may rely on global mutable state; each test MUST be runnable in isolation via
  `go test -run TestName`.

### VI. Observability & Structured Logging

- The binary MUST expose a `--log-level` flag (values: `debug`, `info`, `warn`, `error`)
  defaulting to `info`; level MUST also be configurable via `SIMPLE_CLI_LOG_LEVEL`.
- All log output MUST go to stderr, never stdout.
- In JSON output mode (`--output json`), log entries MUST be emitted as JSON-Lines on stderr.
- The `slog` standard library package (Go 1.21+) MUST be used for all logging; third-party
  logging libraries are prohibited unless `slog` proves provably insufficient.
- Sensitive values (tokens, passwords, keys) MUST NEVER appear in log output at any level;
  redaction helpers MUST be provided in `internal/security/`.

### VII. Documentation Standards

- **Inline**: Every exported Go symbol MUST have a GoDoc comment. Complex algorithms MUST
  carry an explanatory comment block citing sources or rationale.
- **Command help**: Every command and flag MUST have a non-empty `Short` and `Long`
  description in Cobra; the `Long` description MUST include at least one usage example.
- **External docs**: A `docs/` directory at repository root MUST contain:
  - `docs/installation.md` — install instructions per OS with ENV PATH guidance.
  - `docs/quickstart.md` — first 5-minute experience guide.
  - `docs/configuration.md` — all flags, env vars, config file schema.
  - `docs/architecture.md` — package layout, dependency graph, design decisions.
  - `docs/ai-agent-guide.md` — machine-readable output spec, exit codes, env vars for agents.
  - `CHANGELOG.md` — kept in Keep A Changelog format.
- Documentation MUST be updated in the same commit/PR as the code it documents; stale docs
  without a linked issue are a constitution violation.

### VIII. Simplicity & Lightweight Binary

- Binary size MUST remain under 20 MB after `ldflags="-s -w"` stripping.
- Startup time (cold, no network) MUST remain under 150 ms on reference hardware
  (2-core VM, 2 GB RAM).
- External dependencies MUST be justified: prefer standard library; add a direct dependency
  only if it saves > 100 lines of well-tested code or covers a security-sensitive surface.
- `go.sum` MUST be committed; `vendor/` is optional but MUST be consistent if used.
- The binary MUST be statically linked (`CGO_ENABLED=0`) for all release builds to ensure
  drop-in deployment.

### IX. Robustness & Error Resilience

- All I/O operations (file, network, subprocess) MUST honor `context.Context` cancellation
  and deadlines; blocking calls without a context are a violation.
- User-facing error messages MUST be actionable: describe what went wrong AND what the user
  can do to fix it.
- The bundled `run` daemon command MUST implement graceful shutdown on SIGINT/SIGTERM:
  drain in-flight work and exit cleanly within 5 seconds (enforced by `internal/signals`).
  Custom commands added to the template MUST follow the same shutdown contract.
- Configuration parsing errors MUST be reported with the offending file path and line number
  where possible.
- The CLI MUST handle read-only filesystems and missing HOME/APPDATA gracefully, falling
  back to in-memory operation with an explicit warning.

### X. Versioning & Maintainability

- Version follows Semantic Versioning 2.0.0 (SemVer). The version string MUST be injected
  at build time via `-ldflags "-X main.Version=$(git describe --tags)"`.
- `CHANGELOG.md` MUST be updated with every release using Keep A Changelog format
  (`Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security` headings).
- All breaking changes MUST be flagged with a `BREAKING:` prefix in commit messages and
  CHANGELOG, and MUST increment the MAJOR version.
- The module path MUST remain stable; renaming the module path is a MAJOR version event.
- CI MUST run `go vet`, `staticcheck`, and `golangci-lint` on every pull request;
  lint failures MUST block merge.
- Dependency updates MUST be handled via Dependabot or Renovate and reviewed weekly.

## Technology Stack & Conventions

**Runtime**: Go 1.22+ (minimum); CI MUST test on the current stable and previous stable release.

**CLI Framework**: `github.com/spf13/cobra` with `github.com/spf13/viper` for configuration.
This project is a **template**: the bundled sub-command is `simple-cli run` (long-running daemon).
Consumers MUST add their own sub-commands following the pattern `simple-cli <verb>` or
`simple-cli <noun> <verb>` and register them in `cmd/root.go`.

**Build**: `goreleaser` for cross-platform binary and package artifacts; `Makefile` for
developer workflows (`make build`, `make test`, `make lint`, `make install-local`).

**Installer toolchain**:

- Windows: NSIS (`.exe`) + PowerShell script (`install.ps1`)
- Linux: Shell script + `goreleaser` `.deb`/`.rpm` + `/etc/profile.d/` snippet
- macOS: Homebrew formula + shell script + `pkgbuild` PKG

**Testing**: `testing` stdlib + `github.com/stretchr/testify` for assertions; `gomock` or
`testify/mock` for interface mocking; Docker + `testcontainers-go` for sandbox tests.

**Linting**: `golangci-lint` with configuration in `.golangci.yml`; enabled linters MUST
include `errcheck`, `govet`, `staticcheck`, `gosec`, `revive`, `gofumpt`.

**Logging**: `log/slog` (stdlib, Go 1.21+).

**CI/CD**: GitHub Actions; release pipeline triggered by semver tags (`v*.*.*`).

**Configuration precedence** (highest to lowest):

1. CLI flags
2. Environment variables (`SIMPLE_CLI_*`)
3. Config file (`$XDG_CONFIG_HOME/simple-cli/config.yaml`)
4. Built-in defaults

## Installer & Distribution Design

**Artifacts produced by CI on every tagged release**:

```
dist/
├── simple-cli_windows_amd64.exe        # raw binary
├── simple-cli_linux_amd64              # raw binary
├── simple-cli_darwin_amd64             # raw binary (Intel)
├── simple-cli_darwin_arm64             # raw binary (Apple Silicon)
├── simple-cli_windows_amd64_setup.exe  # NSIS installer
├── simple-cli_linux_amd64.deb
├── simple-cli_linux_amd64.rpm
├── simple-cli_darwin_universal.pkg
└── checksums.txt                       # SHA-256 of all artifacts
```

All installers MUST:

1. Place the binary in the OS-appropriate system or user bin directory.
2. Register or verify ENV PATH so `simple-cli` is reachable in a new shell.
3. Write an uninstaller / provide `simple-cli self uninstall` instructions.
4. Be idempotent — running the installer twice MUST NOT create duplicate PATH entries.

**Installation scripts** (`scripts/install/`):

- `install.sh` — POSIX-compatible, works on Linux and macOS.
- `install.ps1` — PowerShell 5.1+ compatible, works on Windows 7+ (no WinRM required).

## ENV PATH Registration Strategy

| Platform            | Scope        | Method                                                      | Persistence                  |
| ------------------- | ------------ | ----------------------------------------------------------- | ---------------------------- |
| Windows (admin)     | Machine      | `HKLM\SYSTEM\...\Environment` via `setx /M` or registry API | Broadcast `WM_SETTINGCHANGE` |
| Windows (user)      | User         | `HKCU\Environment` via `setx`                               | Immediate for new shells     |
| Linux (system)      | All users    | `/etc/profile.d/simple-cli.sh` (deb/rpm)                    | Next login shell             |
| Linux (user)        | Current user | Append to `~/.bashrc` and `~/.zshrc`                        | Next shell spawn             |
| macOS (Homebrew)    | Current user | Managed by Homebrew `bin/` symlink                          | Automatic                    |
| macOS (PKG)         | All users    | `/etc/paths.d/simple-cli`                                   | Next login shell             |
| macOS (user script) | Current user | Append to `~/.zshrc` (default shell)                        | Next shell spawn             |

Post-install validation: every installer MUST run `simple-cli --version` in a subprocess
with a clean environment to confirm PATH registration succeeded, and print a human-readable
success or failure message with remediation guidance.

## AI Agent Interaction Patterns

Agents interacting with `simple-cli` MUST follow this contract:

**Invocation pattern**:

```sh
SIMPLE_CLI_OUTPUT=json simple-cli <command> [flags] 2>agent.err.jsonl
```

**Stdout (success)**:

```json
{"status": "ok", "data": { ... }, "meta": {"version": "1.2.3", "duration_ms": 42}}
```

**Stdout (error)**:

```json
{
  "status": "error",
  "code": "RESOURCE_NOT_FOUND",
  "message": "...",
  "hint": "..."
}
```

**Stderr (progress, JSON-Lines)**:

```json
{
  "level": "info",
  "time": "2026-03-23T10:00:00Z",
  "msg": "...",
  "step": 1,
  "total": 5
}
```

Agents MUST check the exit code first, then parse stdout JSON. The `code` field in error
payloads is a stable machine-readable string (SCREAMING_SNAKE_CASE); message text MAY change
between versions but `code` is a versioned contract.

## Testing Strategy

```
tests/
├── unit/           # mirrors src package structure; pure in-process tests
├── integration/    # black-box CLI tests using exec.Command against compiled binary
└── sandbox/
    ├── docker/
    │   ├── ubuntu-22.04/   # Dockerfile + test script
    │   ├── debian-12/
    │   ├── alpine-3.19/
    │   └── windows-ltsc/   # Windows Server Core image
    └── run_sandbox.sh      # orchestrates all sandbox containers
```

**Unit tests**: `go test ./internal/... ./pkg/... ./cmd/...`
**Integration tests**: `go test ./tests/integration/... -tags=integration`
**Sandbox tests**: `make test-sandbox` (requires Docker); validates installer behavior,
PATH registration, and `simple-cli --version` reachability on clean OS images.
**Coverage check**: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`

CI matrix: test on `ubuntu-latest`, `windows-latest`, `macos-latest` for every PR.

## Governance

This constitution supersedes all other development guidelines, practices, and conventions
for the simple-cli project. All contributors MUST read and acknowledge this document before
making their first pull request.

**Amendment procedure**:

1. Open a GitHub Issue proposing the amendment with rationale.
2. Discussion period: minimum 3 business days for input from maintainers.
3. Amendment PR: updates this constitution AND all affected templates/docs in one atomic commit.
4. Version bump follows the rules in Principle X (MAJOR/MINOR/PATCH).
5. PR MUST be approved by at least one maintainer; auto-merge is prohibited for
   constitution changes.

**Compliance review**: Every pull request MUST include a Constitution Check section in its
plan (see `.specify/templates/plan-template.md`) verifying no principle is violated.
Reviewers MUST reject PRs that bypass this check without explicit maintainer sign-off.

**Rationale for complexity**: When a principle requires complexity that would otherwise
violate Principle VIII (Simplicity), the complexity MUST be justified in a `COMPLEXITY.md`
file at repository root and linked from this constitution.

**Runtime guidance**: `.github/agents/speckit.implement.agent.md` provides per-feature
development guidance aligned with this constitution.

**Version**: 2.0.0 | **Ratified**: 2026-03-23 | **Last Amended**: 2026-03-24
