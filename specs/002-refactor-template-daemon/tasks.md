# Tasks: Refactor to Generic CLI Template with Daemon Run Command

**Input**: `specs/002-refactor-template-daemon/spec.md` (plan.md was not generated — architecture waived; see stub plan.md)
**Spec**: spec.md (required — user stories, FRs, SCs)
**Constitution**: v2.0.0
**Branch**: `002-refactor-template-daemon`

**Note**: No new external dependencies are introduced (FR-011). Every task modifies
existing files or creates files using only the packages already in `go.mod`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (touches a different file from other [P] tasks in the same phase)
- **[US1]**: User Story 1 — daemon runs and shuts down cleanly
- **[US2]**: User Story 2 — template is customisable
- **[US3]**: User Story 3 — AI agent output contract

---

## Phase 1: Removal — Delete Session Source Code

**Purpose**: Remove all session-domain Go files. Tasks are fully independent (different
directories). This phase MUST be completed before any Phase 2 edits to avoid dangling imports.

**Checkpoint**: `cmd/session/` and `internal/session/` directories no longer exist.

- [ ] T001 [P] Delete file `cmd/session/session.go` (parent command registration for `simple-cli session`)
- [ ] T002 [P] Delete file `cmd/session/start.go` (session start sub-command)
- [ ] T003 [P] Delete file `cmd/session/stop.go` (session stop sub-command)
- [ ] T004 [P] Delete file `cmd/session/resume.go` (session resume sub-command)
- [ ] T005 [P] Delete file `cmd/session/list.go` (session list sub-command)
- [ ] T006 [P] Delete file `cmd/session/reset.go` (session reset sub-command)
- [ ] T007 [P] Delete file `internal/session/session.go` (Session struct and NewSession constructor)
- [ ] T008 [P] Delete file `internal/session/store.go` (SessionStore interface definition)
- [ ] T009 [P] Delete file `internal/session/errors.go` (ErrSessionNotFound, ErrSessionExists, etc.)
- [ ] T010 [P] Delete file `internal/session/lock.go` (cross-platform lock interface)
- [ ] T011 [P] Delete file `internal/session/lock_unix.go` (flock-based lock implementation)
- [ ] T012 [P] Delete file `internal/session/lock_windows.go` (LockFileEx-based lock implementation)
- [ ] T013 [P] Delete file `internal/session/filestore.go` (JSON file-backed SessionStore)
- [ ] T014 [P] Delete file `internal/session/memstore.go` (in-memory SessionStore)
- [ ] T015 [P] Delete file `internal/session/proxystore.go` (proxy that delegates to a \*SessionStore pointer)
- [ ] T016 [P] Delete file `internal/session/session_test.go` (unit tests for Session struct)
- [ ] T017 [P] Delete file `internal/session/filestore_test.go` (unit tests for filestore)
- [ ] T018 [P] Delete file `internal/session/memstore_test.go` (unit tests for memstore)
- [ ] T019 [P] Delete file `internal/session/lock_test.go` (unit tests for lock)

---

## Phase 2: Foundational — Core Source Code Refactor

**Purpose**: Update the root command wiring and config package. Phase 1 MUST be complete
first (imports to deleted packages would cause compile errors).

**⚠️ CRITICAL**: T020 and T021 MUST be complete before `go build ./...` can succeed.

- [ ] T020 [US1] [US2] Edit `cmd/root.go` — remove the two session import lines:
      `sessioncmd "github.com/your-org/simple-cli/cmd/session"` and
      `"github.com/your-org/simple-cli/internal/session"`. Remove the package-level
      `sessionStore session.SessionStore` variable. Remove the session store initialisation
      block inside `PersistentPreRunE` (`if sessionStore == nil { ... }`). Remove the
      `_ = v.BindEnv("state_dir", "SIMPLE_CLI_STATE_DIR")` line. Remove the `proxy` variable
      and the `rootCmd.AddCommand(sessioncmd.NewSessionCmd(proxy))` call. Add
      `rootCmd.AddCommand(newRunCmd())` in its place.

- [ ] T021 [US1] [US2] *(depends on T020)* Edit `cmd/root.go` — update `rootCmd` metadata:
  - `Use`: keep `"simple-cli"`
  - `Short`: `"A cross-platform CLI template"`
  - `Long`: replace session-focused paragraph with:

    ```
    simple-cli is a cross-platform CLI template that stays alive as a
    long-running process until the device shuts down.

    Customise it by adding sub-commands in cmd/ and implementing your logic
    inside internal/.

    Use --output json for machine-readable output suitable for AI agent workflows.
    ```

  - `Example`: replace session examples with:

    ```
      # Start the long-running process
      simple-cli run

      # Run with JSON output (for AI agents / scripts)
      simple-cli --output json run
    ```

- [ ] T022 [P] [US1] [US2] Edit `internal/config/config.go` — remove `StateDir string` field
      from the `Config` struct (the `mapstructure:"state_dir"` tagged field). Remove the
      `if cfg.StateDir == ""` block that calls `defaultStateDir()`. Remove the entire
      `defaultStateDir()` function (the function that returns `filepath.Join(...)` paths for
      Windows/XDG/home). Remove `"runtime"` from the imports if it is no longer used by
      `ConfigDir()` (verify: `ConfigDir()` uses `runtime.GOOS` — keep `runtime` import if so).

- [ ] T023 [P] [US1] [US2] Create new file `cmd/run.go` — the daemon sub-command. The file
      must be in `package cmd`. Implement `newRunCmd() *cobra.Command` returning a `*cobra.Command`
      with:
  - `Use: "run"`
  - `Short: "Start the long-running process (stays alive until device shutdown)"`
  - `Long`: multi-line description explaining that the process blocks until SIGINT/SIGTERM
    and that application logic belongs below the `<-ctx.Done()` line
  - `Example`: show `simple-cli run` and `simple-cli --output json run`
  - `RunE` body:
    1. Record `start := time.Now()`
    2. Call `signals.NotifyContext(cmd.Context())` to get a cancellable ctx and stop func; defer stop
    3. Retrieve `cfg` from context: `cfg := ctx.Value(config.CtxKey{}).(*config.Config)`
    4. Construct `w := output.NewWriter(cfg.Quiet)` and `f := output.NewFormatter(cfg.Output, w, cfg.NoColor)`
    5. Log `slog.Info("simple-cli started", "pid", os.Getpid())`
    6. Add a `// TODO: add your application logic here` comment
    7. `<-ctx.Done()` — blocks until signal
    8. Log `slog.Info("shutdown signal received, exiting cleanly")`
    9. `return f.FormatSuccess("run", map[string]any{"status": "stopped", "uptime_ms": time.Since(start).Milliseconds()}, time.Since(start))`
  - Imports: `"log/slog"`, `"os"`, `"time"`, `"github.com/spf13/cobra"`,
    `"github.com/your-org/simple-cli/internal/config"`,
    `"github.com/your-org/simple-cli/internal/output"`,
    `"github.com/your-org/simple-cli/internal/signals"`

**Checkpoint**: `go build ./...` MUST succeed before proceeding to Phase 3.

---

## Phase 3: User Story 1 — Daemon Command Tests 🎯 MVP

**Goal**: Verify `simple-cli run` blocks and shuts down cleanly with a clean exit code.
**Independent Test**: Build binary, run `simple-cli run`, send signal, assert exit 0 + JSON envelope.

- [ ] T024 [US1] Edit `internal/config/config_test.go` — remove the four StateDir test functions:
      `TestStateDirFromEnvXDG`, `TestStateDirFallbackHomeDir`, `TestStateDirFromEnvAPPDATA`,
      `TestStateDirOverride`. Also remove `cfg.StateDir` assertion from `TestLoadDefaults`.
      Remove `"runtime"` from the import block if no other test in the file uses `runtime.GOOS`.
      All other tests (`TestLoadDefaults`, `TestLoadFromViper`, `TestLoadInvalidOutput`,
      `TestLoadInvalidLogLevel`, `TestConfigDirXDG`, `TestConfigDirNotEmpty`,
      `TestLoadFromConfigFile`) MUST remain unchanged.

- [ ] T025 [US1] Delete file `tests/integration/session_test.go` entirely (contains
      `TestSessionLifecycle`, `TestSessionNotFound`, and the `TestMain`+`run` helpers that
      reference `SIMPLE_CLI_STATE_DIR`).

- [ ] T026 [US1] Delete file `tests/integration/quickstart_validation_test.go` entirely
      (contains `TestQuickstartValidation` which exercises `session start`, `session list`,
      `session resume`, `session stop`, `session reset`).

- [ ] T027 [US1] [US3] Rewrite `tests/integration/output_test.go` — keep the `//go:build integration`
      build tag and the `package integration` declaration. Replace all test bodies:
  - Keep the `binaryPath` package-level var and `TestMain` + `run` + `runEnvJSON` helpers
    (remove `SIMPLE_CLI_STATE_DIR` from `run`'s env injection — no state dir needed).
  - `TestVersionOutput`: run `simple-cli --version`, assert exit 0 and stdout contains a
    semver pattern (`v?[0-9]+\.[0-9]+\.[0-9]+`).
  - `TestHelpNoSessionMention`: run `simple-cli --help`, assert stdout does NOT contain
    the word "session".
  - `TestJSONOutputFlag`: start `simple-cli --output json run` as subprocess, send shutdown
    signal after 150 ms via `sendShutdown` (defined in T028 signal helper files), capture
    stdout, assert valid JSON object with `status == "ok"` and exit 0.
  - `TestEnvVarOutputJSON`: same but set `SIMPLE_CLI_OUTPUT=json` via env var instead of flag.
  - Remove `TestErrorExitCodeJSONEnvelope` (it relied on `session stop` with a missing session).
  - Add `TestInvalidOutputFlag`: run `simple-cli --output xml version`, assert exit 2.

- [ ] T028 [US1] [US3] Create new file `tests/integration/run_test.go` — `//go:build integration`,
      `package integration`. Implement:
  - `TestRunBlocksAndExitsOnSignal`: start `simple-cli run` as a subprocess via `exec.Command`;
    after 150 ms send SIGINT (Unix: `cmd.Process.Signal(syscall.SIGINT)`, Windows:
    `cmd.Process.Kill()`); assert exit code 0 and that the process terminates within 5 seconds.
  - `TestRunJSONEnvelope`: same as above but with `--output json`; capture stdout; parse JSON;
    assert `status == "ok"`, `data.status == "stopped"`, `data.uptime_ms >= 0`,
    `meta.command == "run"`.
  - `TestRunHumanOutput`: same without `--output json`; assert stdout is non-empty and
    contains "stopped" (HumanFormatter writes the result map to stdout via WriteOut),
    and stderr contains "shutdown" (from slog).
  - Use build-tag files for signal delivery:
    - Create `tests/integration/signal_unix_test.go` (`//go:build integration && !windows`) with
      `func sendShutdown(p *os.Process) error { return p.Signal(syscall.SIGINT) }`.
    - Create `tests/integration/signal_windows_test.go` (`//go:build integration && windows`) with
      `func sendShutdown(p *os.Process) error { return p.Kill() }`.

---

## Phase 4: User Story 2 — Extension-Point Scaffold & Tests

**Goal**: Verify the template scaffold is clean and ready for customisation.
**Independent Test**: `go build ./...` with a new stub command registered in `cmd/root.go`.

- [ ] T029 [P] [US2] Edit `cmd/root.go` — add an inline doc comment block above the
      `rootCmd.AddCommand(newRunCmd())` line:

  ```go
  // To add a new sub-command, create cmd/mycommand.go with a newMyCmd() constructor
  // following the same pattern as cmd/run.go, then register it here:
  //   rootCmd.AddCommand(newMyCmd())
  ```

  This is the only change for this task.

- [ ] T030 [P] [US2] Edit `cmd/run.go` — expand the TODO comment to a multi-line guide:
  ```go
  // TODO: Add your application logic here. This context is cancelled on SIGINT/SIGTERM.
  // Example:
  //   go myWorker(ctx)   // starts a goroutine that respects ctx.Done()
  //   <-ctx.Done()       // blocks until shutdown signal
  ```

- [ ] T044 [P] [US2] Create `cmd/example_cmd.go` — a minimal, fully documented example
  sub-command demonstrating the extension pattern. In `package cmd`, implement
  `newExampleCmd() *cobra.Command` with:
  - `Use: "example"`, `Short: "Example sub-command — safe to delete when customising"`
  - `RunE` body: retrieve cfg from context, construct formatter, call
    `f.FormatSuccess("example", map[string]any{"message": "replace this with your logic"}, time.Since(start))`
  - GoDoc comment: `// newExampleCmd returns a minimal sub-command demonstrating how to add
  //   commands to the template. Delete this file and add your own in cmd/.`
  Register it in `cmd/root.go` `init()` alongside `newRunCmd()`:
  `rootCmd.AddCommand(newExampleCmd())`
  Also create `cmd/example_cmd_test.go` in `package cmd` with a single table-driven unit
  test `TestExampleCmdRunE` that constructs the command, injects a test context with a
  `*config.Config`, calls `cmd.RunE(cmd, nil)`, and asserts error is nil.

---

## Phase 5: User Story 3 — Documentation Updates

**Goal**: All docs describe a template/daemon, not a session manager.
**Independent Test**: `grep -r "session" docs/ README.md CHANGELOG.md` returns zero matches.

Tasks in this phase are all independent (different files) and can run in parallel.

- [ ] T031 [P] [US3] Edit `README.md` — replace the feature overview, quick-start commands,
      and any usage examples that reference `session start`, `session list`, `session stop`,
      `session resume`, `session reset`. Replace with `simple-cli run` usage. Update the
      "What is this?" section to describe a generic cross-platform CLI template. Remove the
      session-management feature bullet points. Retain the JSON output, installer, and
      AI-agent-interoperability sections.

- [ ] T032 [P] [US3] Edit `docs/quickstart.md` — replace the 5-step session walkthrough
      (`session start` → `session list` → `session resume` → `session stop` → `session reset`)
      with a 3-step daemon walkthrough:
  1. `simple-cli --version` (verify install)
  2. `simple-cli run` (start daemon, observe log)
  3. Ctrl+C (observe clean shutdown)
     Add a "Customising the template" section pointing to `cmd/run.go`.

- [ ] T033 [P] [US3] Edit `docs/configuration.md` — remove the `state_dir` row from the
      configuration reference table. Remove any `SIMPLE_CLI_STATE_DIR` environment variable
      documentation. Update the example config YAML to remove the `state_dir:` key. Add a
      brief note that the daemon template stores no persistent state by default.

- [ ] T034 [P] [US3] Edit `docs/architecture.md` — remove the session store architecture
      section (the diagram or prose describing `internal/session`, `FileStore`, `MemStore`,
      `ProxyStore`, lock files). Update the package layout tree to remove `internal/session/`
      and `cmd/session/`. Add `cmd/run.go` to the tree. Update the dependency graph section
      to remove session-store arrows.

- [ ] T035 [P] [US3] Edit `docs/ai-agent-guide.md` — remove all `session` command examples.
      Replace with `run` command interaction patterns: start subprocess, wait for startup log,
      send SIGINT, parse shutdown JSON envelope. Ensure the exit code table and JSON envelope
      schema sections are still present (they are generic and need no removal, only session
      examples within them).

- [ ] T036 [P] [US3] Edit `CHANGELOG.md` — add an `## [Unreleased]` section (or update
      the existing one) with:

  ```markdown
  ### Removed

  - BREAKING: `session` sub-commands (`start`, `stop`, `resume`, `list`, `reset`) removed.
  - BREAKING: `internal/session` package removed; `StateDir` config field removed.
  - `SIMPLE_CLI_STATE_DIR` environment variable no longer recognised.

  ### Added

  - `run` sub-command: long-running daemon process, stays alive until SIGINT/SIGTERM.
  - `cmd/run.go`: extension point for custom application logic.

  ### Changed

  - Root command description updated to reflect generic CLI template identity.
  - `docs/quickstart.md`, `docs/configuration.md`, `docs/architecture.md`,
    `docs/ai-agent-guide.md` updated to remove session references.
  ```

---

## Phase 6: Polish — Makefile Audit & Final Verification

**Purpose**: Confirm Makefile is clean, binary builds, and all tests pass.

- [ ] T037 [P] Audit `Makefile` — check all targets for references to deleted packages
      (`session`, `internal/session`, `tmp_sess_cover.out`, `cmd/session`). If any are found,
      remove or update them. If `Makefile` has a dedicated session coverage target (e.g., the
      `tmp_sess_cover.out` artifact visible in the repo root), remove that target and the
      output file reference. No structural changes to existing targets needed if clean.

- [ ] T038 [P] Delete artefact files in repo root that are session-specific remnants:
      `tmp_sess_cover.out` (seen in workspace listing). Also verify `tmp_cover.out` does not
      contain session package paths; if it does, delete it (it is a generated file).

- [ ] T039 Run `go build ./...` and fix any remaining compile errors. Expected: zero errors.
      Common failure points: lingering import in a file not covered by previous tasks, missing
      `os` import in `cmd/run.go`, wrong `output.NewWriter` / `output.NewFormatter` call
  signature. After a successful build, assert binary size: `(Get-Item dist\simple-cli.exe).length`
  on Windows or `wc -c dist/simple-cli` on Unix must be < 20 971 520 bytes (SC-006).

- [ ] T040 Run `go test ./internal/... ./pkg/... -coverprofile=coverage.out -covermode=atomic`
  and verify: zero test failures, coverage ≥ 80% on `internal/` and `pkg/` packages.

- [ ] T041 Run `go test -tags integration ./tests/integration/...` (requires built binary at
      `dist/simple-cli`). Verify: `TestRunBlocksAndExitsOnSignal`, `TestRunJSONEnvelope`,
      `TestVersionOutput`, `TestHelpNoSessionMention`, `TestInvalidOutputFlag` all pass.

- [ ] T042 Run `golangci-lint run ./...` and fix any reported issues. Expected: zero issues.
  Common: unused import after session removal, missing GoDoc on `newRunCmd`.

- [ ] T043 **Version bump — Principle X compliance**: Update `CHANGELOG.md` to replace
  `## [Unreleased]` with `## [2.0.0] - 2026-03-24`. Run `git tag v2.0.0` on the branch
  after all other tasks pass. Verify `simple-cli --version` outputs `v2.0.0` when built
  with `make build` (ldflags inject `git describe --tags`). Satisfies the BREAKING change
  version-increment requirement of Constitution Principle X.

---

## Dependencies

```
Phase 1 (T001–T019) ──────────┐
                               ▼
Phase 2 (T020–T023) ───────────┤ (T020 and T021 must complete before T023 uses newRunCmd)
                               ▼
Phase 3 (T024–T028) ─────┐    │
Phase 4 (T029–T030, T044) ──┤    │  (all depend on Phase 2 compile success)
Phase 5 (T031–T036) ─────┘    │
                               ▼
Phase 6 (T037–T043) ─────────── (verification; all prior phases must be complete)
```

Within Phase 2:

- T022 (`internal/config/config.go`) and T023 (`cmd/run.go`) are independent of each other [P]
- T020 and T021 both edit `cmd/root.go` — run sequentially (same file); T021 depends on T020

Within Phase 3:

- T024 (`config_test.go`), T025 (delete session_test.go), T026 (delete quickstart_validation_test.go),
  T027 (rewrite output_test.go), T028 (create run_test.go) are all independent [P]

---

## Parallel Execution Examples

**Phase 1** — all 19 deletion tasks can run simultaneously (different files):

```
T001 T002 T003 T004 T005 T006 T007 T008 T009 T010
T011 T012 T013 T014 T015 T016 T017 T018 T019
```

**Phase 2** — T020+T021 sequentially (same file), T022+T023 in parallel:

```
T020 → T021 (root.go edits)
T022 ║ T023 (config.go + run.go in parallel)
```

**Phase 3 + 4 + 5** — all 15 tasks are independent (different files):

```
T024 T025 T026 T027 T028 T029 T030 T031 T032 T033 T034 T035 T036 T044
```

**Phase 6** — T037+T038 in parallel, then T039→T040→T041→T042→T043 sequentially:

```
T037 ║ T038
T039 → T040 → T041 → T042 → T043
```

---

## Implementation Strategy

**MVP scope** (Phase 1 + Phase 2 + T039): Produces a buildable template binary with `simple-cli run`.
**Full delivery**: All phases — cleanly tested, documented, lint-passing.

Start with T001–T019 (all deletions), then T020–T022 (root + config edits), then T023 (create run.go),
then verify `go build ./...` passes (T039 preview), then proceed to test + doc updates.

---

## Summary

| Phase            | Tasks                        | Parallelisable               | Blocks      |
| ---------------- | ---------------------------- | ---------------------------- | ----------- |
| 1: Removal       | T001–T019 (19 tasks)         | All 19 in parallel           | Phase 2     |
| 2: Core refactor | T020–T023 (4 tasks)          | T022 ∥ T023                  | Phase 3/4/5 |
| 3: US1 tests     | T024–T028 (5 tasks)          | All 5 in parallel            | Phase 6     |
| 4: US2 scaffold  | T029–T030, T044 (3 tasks)    | All 3 in parallel            | Phase 6     |
| 5: US3 docs      | T031–T036 (6 tasks)          | All 6 in parallel            | Phase 6     |
| 6: Verification  | T037–T043 (7 tasks)          | T037 ∥ T038; rest sequential | —           |
| **Total**        | **44 tasks**                 | **36 parallelisable**        |             |

**Suggested MVP**: Phase 1 + Phase 2 + T039 = 24 tasks → `simple-cli run` compiles and runs.
