# Implementation Plan: Refactor to Generic CLI Template with Daemon Run Command

**Branch**: `002-refactor-template-daemon` | **Date**: 2026-03-24 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/002-refactor-template-daemon/spec.md`

> **Note**: This plan was created as a waiver stub. The `/speckit.tasks` command was invoked
> directly after `/speckit.specify` without running `/speckit.plan`. Architecture decisions
> are documented here for traceability and to allow the prerequisites check to pass.

## Summary

Refactor the `simple-cli` codebase from a session-management CLI into a generic, extensible
CLI template with a single bundled `run` daemon sub-command. The refactor is purely subtractive
(delete session packages) and additive (add `cmd/run.go`). No new dependencies are introduced.

## Technical Context

**Language/Version**: Go 1.22+ (existing module, no change)
**Primary Dependencies**: `github.com/spf13/cobra`, `github.com/spf13/viper` (no change)
**Storage**: None — daemon template stores no persistent state (StateDir removed)
**Testing**: `testing` stdlib + `github.com/stretchr/testify`; integration tests via `exec.Command`
**Target Platform**: Linux, macOS, Windows (cross-platform; CGO_ENABLED=0)
**Project Type**: CLI binary — generic template
**Performance Goals**: Startup < 150 ms cold; binary < 20 MB stripped (no change from existing)
**Constraints**: No new external dependencies (FR-011); BREAKING changes require MAJOR version bump (Principle X)
**Scale/Scope**: Single binary; ~14 files deleted, ~3 files created, ~10 files modified

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Library-First | ✅ | No circular imports; new `cmd/run.go` is isolated |
| II. Idiomatic Go | ✅ | All edits follow gofmt; errors returned, not ignored |
| III. Installer | ✅ | No installer changes needed |
| IV. AI Agent I/O | ✅ | JSON envelope preserved; all commands support `--output json` |
| V. Test-First | ✅ | Tests updated alongside code (T024–T028, T044) |
| VI. Observability | ✅ | slog preserved; log-level flag unchanged |
| VII. Documentation | ✅ | Docs updated in same PR scope (T031–T036) |
| VIII. Simplicity | ✅ | Purely subtractive; binary shrinks |
| IX. Robustness | ✅ | `internal/signals` unchanged; 5 s drain contract preserved |
| X. Versioning | ✅ | BREAKING removal of session → v2.0.0 tag (T043) |

## Project Structure

### Documentation (this feature)

```text
specs/002-refactor-template-daemon/
├── plan.md              ← this file
├── spec.md
├── tasks.md
└── checklists/
    └── requirements.md
```

### Source Code Changes

```text
# DELETED
cmd/session/             (6 files — all session sub-commands)
internal/session/        (13 files — store, lock, filestore, memstore, proxystore + tests)

# MODIFIED
cmd/root.go              — remove session wiring; add newRunCmd(); update metadata
internal/config/config.go — remove StateDir field + defaultStateDir()
internal/config/config_test.go — remove StateDir tests
tests/integration/output_test.go — rewrite; remove session tests
README.md, docs/*.md, CHANGELOG.md — remove session references

# CREATED
cmd/run.go               — daemon sub-command (blocks on ctx.Done())
cmd/example_cmd.go       — minimal example sub-command for template users
cmd/example_cmd_test.go  — unit test for example command
tests/integration/run_test.go         — integration tests for run command
tests/integration/signal_unix_test.go — sendShutdown helper (Unix)
tests/integration/signal_windows_test.go — sendShutdown helper (Windows)
```

## Architecture Decision: No Storage Layer

The original design used `internal/session` (file store + memory store + proxy pattern) to
persist session state across terminal restarts. The new template has no such requirement:
the `run` command is a long-lived process that maintains all state in memory. This eliminates:
- File I/O for session CRUD
- Cross-platform file locking (`internal/session/lock*.go`)
- The proxy pattern needed to late-bind the store after config load

`internal/config.Config.StateDir` is removed because no package reads it after session deletion.

## Complexity Tracking

> No constitution violations requiring COMPLEXITY.md justification.

| Concern | Decision |
|---------|----------|
| Version bump | v2.0.0 git tag after all tasks pass (T043) — satisfies Principle X |
| Cross-platform signal in tests | `syscall.SIGINT` on Unix, `p.Kill()` on Windows via build tags — no new dependency |
| plan.md not generated upfront | Acceptable waiver for a pure refactor with no novel architecture; documented here |
