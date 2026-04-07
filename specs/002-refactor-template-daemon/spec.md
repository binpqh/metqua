# Feature Specification: Refactor to Generic CLI Template with Daemon Run Command

**Feature Branch**: `002-refactor-template-daemon`
**Created**: 2026-03-24
**Status**: Draft
**Input**: Refactor source code from session manager to generic CLI template with daemon run command
**Constitution**: v2.0.0

## User Scenarios & Testing _(mandatory)_

### User Story 1 — Developer clones and runs the template out of the box (Priority: P1)

A developer wants a cross-platform CLI starting point. They clone the repository, build it, and immediately run `simple-cli run`. The process starts, logs that it is alive, and stays running until the developer presses Ctrl+C (or the machine shuts down), at which point it exits cleanly within 5 seconds.

**Why this priority**: This is the entire point of the refactor. Without a working scaffold that stays alive as a daemon, nothing else has value.

**Independent Test**: Build the binary and run `simple-cli run`. The process must not exit on its own. Sending SIGINT must cause a clean shutdown message and exit 0.

**Acceptance Scenarios**:

1. **Given** the binary is built, **When** `simple-cli run` is executed, **Then** the process starts, emits a startup log entry, and blocks.
2. **Given** the process is running, **When** Ctrl+C (SIGINT) is sent, **Then** the process logs "shutdown signal received", exits 0, and does so within 5 seconds.
3. **Given** the process is running, **When** `simple-cli --output json run` is used, **Then** on shutdown, the process emits a JSON envelope `{"status":"ok","data":{"status":"stopped","uptime_ms":...},...}` to stdout.
4. **Given** the binary is built, **When** `simple-cli --version` is run, **Then** the output contains a semver string and exit 0.
5. **Given** the binary is built, **When** `simple-cli --help` is run, **Then** the help text describes a template, not a session manager, and lists `run` as the primary sub-command.

---

### User Story 2 — Developer customises the template for their own domain (Priority: P2)

A developer forks or clones the repository and wants to add their own business logic sub-command (e.g., `simple-cli serve`). They add a new file in `cmd/`, register it in `cmd/root.go`, and the new command works with the existing `--output json`, `--log-level`, and `--quiet` flags automatically.

**Why this priority**: The core promise of a template is extensibility. Without clear, working extension points, the template provides no advantage over starting from scratch.

**Independent Test**: Add a minimal no-op sub-command, register it, build, and verify `simple-cli mycommand --output json` emits a valid JSON envelope and exit 0.

**Acceptance Scenarios**:

1. **Given** a new `cmd/mycommand.go` is created following the existing pattern, **When** it is registered via `rootCmd.AddCommand(...)`, **Then** it inherits global flags (`--output`, `--log-level`, `--no-color`, `--quiet`) automatically.
2. **Given** a new command emits output via the shared `output.Formatter`, **When** `--output json` is passed, **Then** the output conforms to the standard JSON envelope schema.
3. **Given** the project builds with no new dependencies, **When** `go build ./...` is run, **Then** it completes without errors and the binary is under 20 MB.

---

### User Story 3 — AI agent or script consumes the CLI output reliably (Priority: P3)

An AI agent or CI script needs to invoke `simple-cli` and parse its output deterministically. It sets `SIMPLE_CLI_OUTPUT=json`, calls the `run` command, sends SIGINT after a delay, and parses the shutdown JSON envelope.

**Why this priority**: Machine-readable output is a first-class requirement per the constitution. Human users can read help text; agents need the JSON contract.

**Independent Test**: Run `simple-cli --output json run` in a subprocess, send SIGINT after 100 ms, capture stdout, and assert the JSON envelope structure is valid.

**Acceptance Scenarios**:

1. **Given** `SIMPLE_CLI_OUTPUT=json` is set, **When** any command exits successfully, **Then** stdout is a single-line JSON object `{"status":"ok","data":{...},"meta":{"version":"...","duration_ms":...,"command":"..."}}`.
2. **Given** a command fails, **When** the process exits non-zero, **Then** stderr contains `{"status":"error","code":"SCREAMING_SNAKE_CASE","message":"...","meta":{...}}` and stdout is empty.
3. **Given** exit codes are tested, **Then** they match the documented contract: 0 success, 1 general error, 2 misuse, 3 not found, 4 permission denied, 5 timeout.

---

### Edge Cases

- **SIGTERM during startup** (before fully initialised): the process must still exit cleanly within 5 seconds and not panic.
- **Double signal** (Ctrl+C pressed twice rapidly): drainTimeout forces exit; process must not hang or exit before draining.
- **`--output` value is invalid** (not `human` or `json`): process exits 2 with a clear human-readable error before executing any command logic.
- **`--log-level` is invalid**: same as above — config validation catches it at startup.
- **Binary run without any sub-command**: Cobra's default help is shown; exit 0.
- **Unknown sub-command**: Cobra returns exit 1 and "unknown command" message to stderr.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: All session-management source code (`cmd/session/`, `internal/session/`) MUST be removed from the codebase.
- **FR-002**: `cmd/root.go` MUST be updated: remove session imports, remove `sessionStore` variable, remove session store initialisation from `PersistentPreRunE`, and remove the `state_dir` env-var binding.
- **FR-003**: A new `cmd/run.go` file MUST be created providing a `run` sub-command that blocks on a signal-aware context until SIGINT/SIGTERM, then exits cleanly.
- **FR-004**: `internal/config.Config` MUST remove the `StateDir` field and the `defaultStateDir()` helper because daemon operation requires no persistent file-based state.
- **FR-005**: The `run` command MUST emit a JSON shutdown envelope (status, uptime_ms) when `--output json` is active, and a human-readable shutdown line otherwise.
- **FR-006**: Session integration tests MUST be handled as follows: `tests/integration/session_test.go` MUST be deleted; `tests/integration/quickstart_validation_test.go` MUST be deleted; `tests/integration/output_test.go` MUST be rewritten to test `--version` output and `--output json` flag envelope (no session references); a new `tests/integration/run_test.go` MUST be created to test the `run` command lifecycle (start → block → SIGINT → clean exit). Cross-platform signal delivery: `syscall.SIGINT` on Unix, `cmd.Process.Kill()` on Windows, differentiated via build tags.
- **FR-007**: All unit tests referencing `StateDir` in `internal/config/config_test.go` MUST be removed.
- **FR-008**: `README.md`, `docs/quickstart.md`, `docs/configuration.md`, `docs/architecture.md`, `docs/ai-agent-guide.md`, and `CHANGELOG.md` MUST be updated to remove session references and reflect the template/daemon identity.
- **FR-009**: `rootCmd.Short`, `rootCmd.Long`, and `rootCmd.Example` MUST describe a generic CLI template, not a session manager.
- **FR-010**: The `internal/signals` package MUST remain unchanged; it already implements the 5-second drain contract required by Principle IX.
- **FR-011**: No new external dependencies MUST be introduced; the refactor is purely subtractive/reorganisational.
- **FR-012**: `go build ./...` and `go test ./...` MUST pass with zero errors after the refactor.
- **FR-013**: `golangci-lint run` MUST pass with zero lint errors after the refactor.
- **FR-014**: The `Makefile` MUST be audited for session-specific targets (e.g., test targets that run deleted packages or reference `internal/session`) and any such targets MUST be removed or updated to reference the new package structure. CI workflow files MUST also be checked for references to deleted packages.

### Key Entities

- **`cmd/run.go`**: New file — the daemon sub-command. Owns the process lifecycle loop (`<-ctx.Done()`), emits startup log and shutdown result.
- **`cmd/root.go`**: Modified — root Cobra command wiring; now registers only `run` as the bundled sub-command.
- **`internal/config.Config`**: Modified — `Output`, `LogLevel`, `NoColor`, `Quiet` fields only; `StateDir` removed.
- **`internal/signals`**: Unchanged — provides `NotifyContext` (SIGINT/SIGTERM + 5 s drain).
- **`internal/output`**: Unchanged — `Formatter` and `Writer` provide the JSON/human output layer.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: `go build ./...` completes with zero errors on Linux, macOS, and Windows CI runners.
- **SC-002**: `go test ./...` passes with zero failures; `internal/` and `pkg/` maintain ≥ 80% line coverage.
- **SC-003**: `golangci-lint run` reports zero issues with the existing `.golangci.yml` configuration.
- **SC-004**: Running `simple-cli run` keeps the process alive; sending SIGINT causes clean exit within 5 seconds — verified by an automated integration test. Cross-platform signal delivery: `syscall.SIGINT` on Unix, `cmd.Process.Kill()` on Windows, differentiated by build tags in the integration test file.
- **SC-005**: `simple-cli --help` output contains no mention of "session" — verified by a string-match assertion in tests or CI.
- **SC-006**: Binary size remains under 20 MB (stripped) on all platforms.
- **SC-007**: No file under `cmd/session/` or `internal/session/` exists in the repository after the refactor.
- **SC-008**: All docs pages (`README.md`, `docs/*.md`, `CHANGELOG.md`) contain no unremediated references to session management.

## Assumptions

- The existing `internal/signals`, `internal/output`, `internal/exitcode`, `internal/security`, and `pkg/version` packages require no changes — they are already generic.
- The Makefile MUST be audited (FR-014); no session-specific targets are expected but this must be verified. Installer scripts (`scripts/install/`) and goreleaser config do not reference session logic and are out of scope.
- The 5-second drain timeout enforced by `internal/signals` is sufficient for the empty daemon loop; no additional drain logic is required.
- The `specs/001-cli-long-life-session/` directory is retained for historical reference and is not deleted.
