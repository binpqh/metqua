# Tasks: CLI Long-Life Session Application

**Feature**: `001-cli-long-life-session`
**Input**: Design documents from `specs/001-cli-long-life-session/`
**Prerequisites**: plan.md ✅ | spec.md ✅ | research.md ✅ | data-model.md ✅ | contracts/ ✅
**Generated**: 2026-03-24

---

## Format: `[ID] [P?] [Story?] Description with file path`

- **[P]**: Parallelizable — operates on different files, no dependency on incomplete sibling tasks
- **[US1]–[US5]**: User story label (required for all story-phase tasks)
- Setup and Foundational tasks carry **no** story label
- Every task includes an exact file path

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the Go module, directory layout, and tooling scaffolding that every subsequent phase depends on.

- [x] T001 Create full project directory structure per plan.md: `cmd/session/`, `internal/config/`, `internal/output/`, `internal/security/`, `internal/session/`, `internal/signals/`, `pkg/version/`, `scripts/install/`, `installer/windows/`, `installer/macos/`, `tests/integration/`, `tests/sandbox/docker/`, `docs/`
- [x] T002 Initialize `go.mod` (module `github.com/binpqh/simple-cli`, Go 1.22) and run `go get` for primary dependencies: `github.com/spf13/cobra`, `github.com/spf13/viper`, `github.com/google/uuid`, `github.com/stretchr/testify`, `golang.org/x/sys` in `go.mod` / `go.sum`
- [x] T003 [P] Create `.golangci.yml` enabling `errcheck`, `gosec`, `revive`, `staticcheck`, `gofmt` linters with per-linter settings and `issues.exclude-rules` for test files in `.golangci.yml`
- [x] T004 [P] Create `Makefile` skeleton defining `build`, `test`, `lint`, `install-local`, and `test-sandbox` phony targets with placeholder bodies in `Makefile`

**Checkpoint**: Module initialised, tooling configured — foundational implementation can begin.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared internal packages and the root Cobra command that every user story depends on. **No user story work can begin until this phase is complete.**

- [x] T005 [P] Create `pkg/version/version.go` with build-time-injectable `Version`, `Commit`, and `BuildDate` variables (default `"dev"`) and a `String() string` method returning the formatted version line
- [x] T006 [P] Create `internal/security/redact.go` with `Redact(key, value string) string` helper that returns `"[REDACTED]"` for sensitive key names (token, password, secret, key) to prevent credential leakage in logs
- [x] T007 [P] Create `internal/signals/signals.go` implementing `NotifyContext(parent context.Context) (context.Context, context.CancelFunc)` wrapping `signal.NotifyContext` for SIGINT/SIGTERM with a 5-second drain timeout via `context.WithTimeout`
- [x] T008 Create `internal/config/config.go` with `Config` struct (`Output`, `LogLevel`, `NoColor`, `Quiet`, `StateDir`), `Load(v *viper.Viper) (*Config, error)` constructor binding Cobra flags and env vars, and `StateDir()` resolving `$XDG_STATE_HOME/simple-cli` (Linux/macOS) or `%APPDATA%\simple-cli` (Windows) and `ConfigDir()` resolving XDG config path in `internal/config/config.go`
- [x] T009 Create `internal/output/writer.go` with `Writer` struct holding `Out io.Writer` (stdout) and `Err io.Writer` (stderr), plus `WriteOut(p []byte)` and `WriteErr(p []byte)` methods — used by formatter to route payload vs diagnostics in `internal/output/writer.go`
- [x] T010 Create `internal/output/output.go` with `Formatter` interface (`FormatSuccess(data any) error`, `FormatError(code, message, hint string) error`), `HumanFormatter` struct, and skeleton `JSONFormatter` struct that both satisfy `Formatter` in `internal/output/output.go`
- [x] T011 Create `cmd/root.go` with Cobra root command (`Use: "simple-cli"`, `Short`, `Long`, `Example`), all persistent global flags (`--output`, `--log-level`, `--no-color`, `--quiet`, `--config`), `--version` flag printing `pkg/version.String()`, `slog` initialization routing to stderr, and `PersistentPreRunE` loading config via `internal/config` in `cmd/root.go`

**Checkpoint**: All shared packages in place and root command wired — user story phases can now begin.

---

## Phase 3: User Story 1 — Install & Run First Command (Priority: P1) 🎯 MVP

**Goal**: Ship cross-platform installers that place `simple-cli` on the user's PATH so `simple-cli --version` works immediately after install on Windows, Linux, and macOS.

**Independent Test**: Run installer on a clean OS image, open a new shell, execute `simple-cli --version` → binary found and returns version string.

### Implementation for User Story 1

- [x] T012 [P] [US1] Write `scripts/install/install.sh` as a POSIX shell installer: detect OS (Linux/macOS), download the correct binary to `~/.local/bin` (or `/usr/local/bin` with sudo), idempotently append `export PATH="$PATH:~/.local/bin"` to `~/.bashrc` and `~/.zshrc` only if the directory is not already present, print success message in `scripts/install/install.sh`
- [x] T013 [P] [US1] Write `scripts/install/install.ps1` as a PowerShell 5.1+ installer: download binary to `$env:LOCALAPPDATA\simple-cli\bin`, check `[Environment]::GetEnvironmentVariable("PATH","User")` for idempotency before calling `[Environment]::SetEnvironmentVariable`, broadcast `WM_SETTINGCHANGE` so running shells refresh without restart, support elevation for machine-scope install in `scripts/install/install.ps1`
- [x] T014 [P] [US1] Create `installer/windows/setup.nsi` NSIS installer script: define `InstallDir`, `WriteRegStr` for machine-scope PATH registration in `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, include uninstaller section that removes the registry PATH entry in `installer/windows/setup.nsi`
- [x] T015 [P] [US1] Create `installer/macos/postinstall` PKG postinstall shell script: write a single-line `/usr/local/bin` (or Homebrew `bin/`) path to `/etc/paths.d/simple-cli`, guard with `[ -f /etc/paths.d/simple-cli ]` idempotency check in `installer/macos/postinstall`
- [x] T016 [US1] Configure `.goreleaser.yml` with build matrix (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`), `CGO_ENABLED=0`, `ldflags: ["-s -w -X main.Version={{.Version}} -X main.Commit={{.Commit}} -X main.BuildDate={{.Date}}"]`, nfpm `.deb` / `.rpm` packages, `.tar.gz` (unix) and `.zip` (windows) archives, SHA-256 checksum file in `.goreleaser.yml`
- [x] T017 [US1] Write `docs/installation.md` covering Homebrew one-liner, Linux `install.sh` with curl, Windows PowerShell one-liner, Windows NSIS GUI installer, manual PATH troubleshooting table per OS, and post-install verification step (`simple-cli --version`) in `docs/installation.md`

**Checkpoint**: Installers complete, `.goreleaser.yml` configured — `simple-cli --version` works on all target platforms.

---

## Phase 4: User Story 2 — Start and Resume a Long-Life Session (Priority: P1)

**Goal**: Users can `session start`, close the terminal, reopen it, and `session resume` with state fully intact. Concurrent writes are safe via file locking.

**Independent Test**: `simple-cli session start`, close terminal, reopen terminal, `simple-cli session resume` → session ID matches, state is intact.

### Implementation for User Story 2

- [x] T018 [P] [US2] Create `internal/session/session.go`
- [x] T019 [P] [US2] Create `internal/session/store.go`
- [x] T020 [P] [US2] Create `internal/session/errors.go`
- [x] T021 [P] [US2] Create `internal/session/lock.go`
- [x] T022 [US2] Create `internal/session/filestore.go`
- [x] T023 [US2] Create `internal/session/memstore.go`
- [x] T024 [P] [US2] Create `cmd/session/session.go`
- [x] T025 [US2] Create `cmd/session/start.go`
- [x] T026 [US2] Create `cmd/session/resume.go`
- [x] T027 [US2] Create `cmd/session/list.go`
- [x] T028 [P] [US2] Create `cmd/session/stop.go`
- [x] T029 [P] [US2] Create `cmd/session/reset.go`
- [x] T030 [US2] Register the `session` command group in `cmd/root.go`

**Checkpoint**: Full session lifecycle (start → list → resume → stop → reset) works and state persists across terminal restarts.

---

## Phase 5: User Story 3 — AI Agent Machine-Readable Output (Priority: P2)

**Goal**: Every command supports `--output json` (or `SIMPLE_CLI_OUTPUT=json`) producing stable JSON envelopes on stdout; all errors appear as JSON on stderr; exit codes are deterministic.

**Independent Test**: `simple-cli session list --output json` returns valid JSON array on stdout; `simple-cli no-such-cmd --output json` returns JSON error on stderr with exit code ≠ 0.

### Implementation for User Story 3

- [x] T031 [P] [US3] Implement `SuccessResponse`, `ErrorResponse`, `Meta`, `JSONFormatter.FormatSuccess/FormatError` in `internal/output/output.go`
- [x] T032 [US3] Wire `JSONFormatter` into all five session commands
- [x] T033 [P] [US3] Implement `slog` handler switching in `cmd/root.go`
- [x] T034 [US3] Implement env var bindings (`SIMPLE_CLI_OUTPUT`, `SIMPLE_CLI_LOG_LEVEL`, `NO_COLOR`) in `internal/config/config.go` and `cmd/root.go`
- [x] T035 [P] [US3] Implement ANSI stripping in `internal/output/output.go` `HumanFormatter`
- [x] T036 [P] [US3] Implement `--quiet` suppression in `internal/output/writer.go`
- [x] T037 [US3] Write `docs/ai-agent-guide.md`

**Checkpoint**: All commands produce schema-valid JSON with stable exit codes; AI agents can parse stdout without grepping stderr.

---

## Phase 6: User Story 4 — Developer Local Build & Test Workflow (Priority: P2)

**Goal**: Contributors run `make build`, `make test`, `make lint` on all three platforms with zero manual steps. `make test` enforces ≥80% coverage. `make install-local` dog-foods the binary immediately.

**Independent Test**: On a fresh clone with Go 1.22+ and `golangci-lint` installed, `make build` produces `dist/simple-cli`, `make test` passes with ≥80% coverage, `make lint` reports zero issues.

### Implementation for User Story 4

- [x] T038 [P] [US4] Finalize `Makefile` with complete tested bodies: `build` → `CGO_ENABLED=0 go build -ldflags="-s -w -X ..." -o dist/simple-cli ./...`; `test` → `go test ./... -coverprofile=coverage.out` then `go tool cover -func=coverage.out | grep total` with `awk` gate failing if coverage < 80%; `lint` → `golangci-lint run ./...`; `install-local` → `cp dist/simple-cli $(go env GOPATH)/bin/` in `Makefile`
- [x] T039 [P] [US4] Finalise `.goreleaser.yml` release pipeline: add `changelog.use: github`, GitHub Release `extra_files` listing installer artifacts, Homebrew tap formula stub (`brews:` block with `install` and `test` do blocks) in `.goreleaser.yml`
- [x] T040 [US4] Write unit tests achieving ≥80% line coverage for `internal/session/`: `internal/session/session_test.go` (SessionStatus JSON round-trip), `internal/session/filestore_test.go` (CRUD operations, concurrent lock safety with `t.Parallel()`), `internal/session/memstore_test.go` (concurrent read/write with goroutines), `internal/session/lock_test.go` (lock/unlock, double-lock blocks)
- [x] T041 [P] [US4] Write unit tests in `internal/config/config_test.go` covering: flag > env > file > default precedence, XDG path resolution on Linux (mock `XDG_STATE_HOME`), `%APPDATA%` path on Windows via env override, config file parse-error reporting with file path in error message
- [x] T042 [P] [US4] Write unit tests in `internal/output/output_test.go` (JSON envelope correct fields, no-color stripping, quiet suppresses stdout) and `internal/output/writer_test.go` (WriteOut goes to Out, WriteErr goes to Err)
- [x] T043 [P] [US4] Write unit tests in `internal/security/redact_test.go` (Redact returns `[REDACTED]` for sensitive keys, passes through safe keys) and `internal/signals/signals_test.go` (context cancelled on interrupt, 5-second drain timeout fires)
- [x] T044 [US4] Write integration tests in `tests/integration/session_test.go` covering the full black-box session lifecycle: start → list (verify entry) → resume (verify status active) → stop (verify status stopped) → reset (verify new ID, fresh state) using the compiled binary via `exec.Command`
- [x] T045 [US4] Write integration tests in `tests/integration/output_test.go` validating: `--output json` stdout parses as valid JSON with required envelope fields; error conditions produce `{"status":"error",...}` on stderr with matching exit code per `contracts/exit-codes.md`; `SIMPLE_CLI_OUTPUT=json` env var behaves identically to `--output json` flag
- [x] T046 [P] [US4] Write `docs/architecture.md` describing: numbered package dependency diagram, design decisions from `research.md` (file-backed store, Cobra+Viper, slog, goreleaser), data flow from CLI flag → Config → Formatter → Writer in `docs/architecture.md`
- [x] T047 [P] [US4] Write `docs/configuration.md` with a complete reference table of all flags, env vars, config file keys, types, defaults, and descriptions in `docs/configuration.md`

**Checkpoint**: `make build && make test && make lint` all pass; ≥80% coverage enforced; `make install-local` works.

---

## Phase 7: User Story 5 — Sandbox Installer Testing (Priority: P3)

**Goal**: CI spins up clean Docker containers for Ubuntu, Debian, Alpine, and Windows Server Core, runs the installer, and validates `simple-cli --version` exits 0 inside each container.

**Independent Test**: `docker compose -f tests/sandbox/docker-compose.yml up --abort-on-container-exit` exits 0 with all container tests passing.

### Implementation for User Story 5

- [x] T048 [US5] Create `tests/sandbox/docker-compose.yml` with four services (`ubuntu`, `debian`, `alpine`, `windows`) each referencing its Dockerfile context, binding the `dist/` binary as a volume, and using `simple-cli --version` as the container CMD; set `x-sandbox-common` extension for shared `build.args` in `tests/sandbox/docker-compose.yml`
- [x] T049 [P] [US5] Create `tests/sandbox/docker/ubuntu-22.04/Dockerfile` with `FROM ubuntu:22.04`, `RUN apt-get install -y curl`, `COPY scripts/install/install.sh /install.sh`, `RUN bash /install.sh` then `CMD ["simple-cli","--version"]` in `tests/sandbox/docker/ubuntu-22.04/Dockerfile`
- [x] T050 [P] [US5] Create `tests/sandbox/docker/debian-12/Dockerfile` with `FROM debian:12-slim`, curl install, `install.sh` run, and `simple-cli --version` CMD in `tests/sandbox/docker/debian-12/Dockerfile`
- [x] T051 [P] [US5] Create `tests/sandbox/docker/alpine-3.19/Dockerfile` with `FROM alpine:3.19`, `RUN apk add file bash`, `COPY dist/simple-cli /usr/local/bin/`, `RUN file /usr/local/bin/simple-cli | grep -q "statically linked"`, `CMD ["simple-cli","--version"]` in `tests/sandbox/docker/alpine-3.19/Dockerfile`
- [x] T052 [P] [US5] Create `tests/sandbox/docker/windows-ltsc/Dockerfile` with `FROM mcr.microsoft.com/windows/servercore:ltsc2022`, `COPY dist/simple-cli.exe C:\simple-cli.exe`, `COPY scripts/install/install.ps1 C:\install.ps1`, `RUN powershell -File C:\install.ps1` then `CMD ["simple-cli","--version"]` in `tests/sandbox/docker/windows-ltsc/Dockerfile`
- [x] T053 [US5] Wire sandbox tests into `Makefile` `test-sandbox` target: `docker compose -f tests/sandbox/docker-compose.yml up --build --abort-on-container-exit --exit-code-from ubuntu`; document Windows container requirement (requires Windows CI runner) with a `$(SKIP_WINDOWS_SANDBOX)` guard variable in `Makefile`

**Checkpoint**: `make test-sandbox` passes on Linux CI for ubuntu/debian/alpine; Windows service validated on Windows CI runner.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Finalise user-facing documentation and project-level files; validate the complete quickstart end-to-end.

- [x] T054 [P] Write `docs/quickstart.md` covering: prerequisites, install step for each platform, `simple-cli --version` verification, `session start`, `session list`, terminal-close and `session resume`, `session stop`, `session reset --force`, and AI agent `--output json` mode — exactly mirroring (and superseding) the draft at `specs/001-cli-long-life-session/quickstart.md` in `docs/quickstart.md`
- [x] T055 [P] Write `README.md` with: project headline, OS/Go badges, one-line install commands per platform, 3-command quick-start (`session start` → `session resume` → `session stop`), feature highlights (session persistence, AI agent JSON output, cross-platform installers), contribution guide pointing to `CONTRIBUTING.md` (to be created) in `README.md`
- [x] T056 [P] Create `CHANGELOG.md` with initial `## [1.0.0] — Unreleased` section listing all features from `spec.md` user stories as changelog entries in `CHANGELOG.md`
- [x] T057 Run End-to-End quickstart.md validation against the built binary: build with `make build`, then execute every command from `docs/quickstart.md` in sequence in a subprocess and assert the expected exit code and output pattern — record result in `tests/integration/quickstart_validation_test.go`

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)
  └─► Phase 2 (Foundational)   — BLOCKS all user stories
        ├─► Phase 3 (US1)      — independent of US2; can start immediately after Phase 2
        ├─► Phase 4 (US2)      — independent of US1; can start immediately after Phase 2
        │     └─► Phase 5 (US3) — depends on US2 session commands existing
        │           └─► Phase 6 (US4) — depends on complete implementation to write tests
        ├─► Phase 6 (US4 – Makefile/goreleaser tasks) — [P] items can start after Phase 2
        └─► Phase 7 (US5)      — depends on US1 installer scripts being complete
              └─► Phase 8 (Polish) — depends on all stories complete
```

### User Story Dependencies

| User Story | Phase | Depends On         | Notes                                                                    |
| ---------- | ----- | ------------------ | ------------------------------------------------------------------------ |
| US1 (P1)   | 3     | Phase 2 complete   | Independent of US2; touches only installer files and goreleaser config   |
| US2 (P1)   | 4     | Phase 2 complete   | Independent of US1; touches only `internal/session/` and `cmd/session/`  |
| US3 (P2)   | 5     | US2 complete       | JSON output integrates with session commands; can't start before T030    |
| US4 (P2)   | 6     | US2 + US3 complete | Tests cover all packages; final Makefile and goreleaser in this phase    |
| US5 (P3)   | 7     | US1 complete       | Sandbox tests require installer scripts and a compiled binary in `dist/` |

### Within Each User Story

- Session entity and interface before implementations: T018, T019, T020 → T022, T023
- Lock utility before FileStore: T021 → T022
- FileStore and MemStore before commands: T022, T023 → T025–T029
- Session command group before sub-commands: T024 → T025–T029
- All session commands before JSON integration: T025–T030 → T031, T032
- Output envelope types before integration: T031 → T032

### Parallel Opportunities

All tasks marked `[P]` within the same phase can be dispatched simultaneously:

| Phase  | Parallel batch                                                                                           |
| ------ | -------------------------------------------------------------------------------------------------------- |
| Setup  | T003, T004 in parallel after T001+T002                                                                   |
| Found. | T005, T006, T007 in parallel; T009, T010 in parallel after T008                                          |
| US1    | T012, T013, T014, T015 all in parallel; T016 after T002; T017 independent                                |
| US2    | T018, T019, T020, T021 in parallel; T022, T023 in parallel after T021; T028, T029 in parallel after T022 |
| US3    | T031, T033, T035, T036 in parallel after T030; T034 independent after T008                               |
| US4    | T038, T039, T041, T042, T043, T046, T047 all in parallel; T040 after T022/T023; T044, T045 after T030    |
| US5    | T049, T050, T051, T052 in parallel after T048                                                            |
| Polish | T054, T055, T056 in parallel                                                                             |

---

## Parallel Execution Examples

### US1 — Install & Run First Command

```
# Four installer artifacts are completely file-independent — run together:
T012  scripts/install/install.sh        (POSIX script)
T013  scripts/install/install.ps1       (Windows PowerShell)
T014  installer/windows/setup.nsi       (NSIS installer)
T015  installer/macos/postinstall       (PKG postinstall)

# Then sequentially:
T016  .goreleaser.yml                   (depends on knowing all targets from above)
T017  docs/installation.md             (depends on knowing all install methods)
```

### US2 — Start and Resume a Long-Life Session

```
# Batch 1 — pure data types (no file conflicts):
T018  internal/session/session.go
T019  internal/session/store.go
T020  internal/session/errors.go
T021  internal/session/lock.go

# Batch 2 — implementations (depend on T018–T021):
T022  internal/session/filestore.go
T023  internal/session/memstore.go
T024  cmd/session/session.go

# Batch 3 — stop and reset are file-independent:
T028  cmd/session/stop.go
T029  cmd/session/reset.go

# Sequential (depend on store implementations):
T025 → T026 → T027 → T030
```

### US4 — Developer Build & Test Workflow

```
# Parallel infrastructure (different files):
T038  Makefile            (build tooling)
T039  .goreleaser.yml     (release tooling)
T041  internal/config/config_test.go
T042  internal/output/*_test.go
T043  internal/security/redact_test.go + internal/signals/signals_test.go
T046  docs/architecture.md
T047  docs/configuration.md

# After all implementation complete:
T040  internal/session/*_test.go       (depends on T022, T023)
T044  tests/integration/session_test.go      (depends on T030)
T045  tests/integration/output_test.go       (depends on T032)
```

---

## Implementation Strategy

### MVP Scope (User Stories 1 + 2 Only)

1. Complete **Phase 1** (Setup) — ~1 session
2. Complete **Phase 2** (Foundational) — ~1 session
3. Complete **Phase 3** (US1) — installer + PATH + goreleaser
4. Complete **Phase 4** (US2) — session persistence + commands
5. **STOP and VALIDATE**: Run independent test for US1 and US2 manually
6. Deliver: binary installs, `--version` works, `session start/resume/stop/reset/list` all work with file persistence

### Incremental Delivery

| Increment | Phases  | Deliverable                                                 |
| --------- | ------- | ----------------------------------------------------------- |
| 1 — MVP   | 1+2+3+4 | Installable CLI with working session lifecycle              |
| 2 — Agent | +5      | JSON output, stable exit codes, AI agent integration        |
| 3 — DevEx | +6      | Passing `make test` (≥80%), `make lint`, full documentation |
| 4 — CI    | +7      | Docker sandbox tests for all installer platforms            |
| 5 — Final | +8      | Complete docs, README, CHANGELOG, E2E quickstart validation |

### Parallel Team Strategy

With a team of 3 after Phase 2 is complete:

- **Developer A**: US1 (installers, goreleaser, docs/installation.md)
- **Developer B**: US2 (session package, session commands)
- **Developer C**: Begins US4 Makefile/goreleaser finalization + docs stubs

Once US2 is complete:

- **Developer A or B** picks up US3 (JSON output integration)
- **Developer C** begins US4 tests as soon as implementation files exist

---

## Summary

| Metric                            | Value                                   |
| --------------------------------- | --------------------------------------- |
| Total tasks                       | 57                                      |
| Phase 1 — Setup                   | 4 tasks (T001–T004)                     |
| Phase 2 — Foundational            | 7 tasks (T005–T011)                     |
| Phase 3 — US1 (Install & Version) | 6 tasks (T012–T017)                     |
| Phase 4 — US2 (Session)           | 13 tasks (T018–T030)                    |
| Phase 5 — US3 (AI Output)         | 7 tasks (T031–T037)                     |
| Phase 6 — US4 (Dev Workflow)      | 10 tasks (T038–T047)                    |
| Phase 7 — US5 (Sandbox)           | 6 tasks (T048–T053)                     |
| Phase 8 — Polish                  | 4 tasks (T054–T057)                     |
| Parallelizable tasks `[P]`        | 32 of 57 (56%)                          |
| Suggested MVP scope               | Phases 1–4 (US1 + US2) = 30 tasks       |
| Format validated                  | ✅ All 57 tasks follow checklist format |
