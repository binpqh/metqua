# Research: CLI Long-Life Session Application

**Branch**: `001-cli-long-life-session` | **Date**: 2026-03-23
**Status**: Complete — all NEEDS CLARIFICATION items resolved

---

## 1. Go Project Layout

**Decision**: Standard Go project layout with `cmd/`, `internal/`, `pkg/`, `scripts/`, `tests/`, `docs/`.
**Rationale**: The [golang-standards/project-layout](https://github.com/golang-standards/project-layout) pattern is widely recognized and enforced by `golangci-lint`. `internal/` prevents external import of domain logic; `pkg/` exposes only the stable public `version` package. `cmd/` separates CLI entry points from business logic, enabling clean unit testing of logic without CLI wiring.
**Alternatives considered**: Flat layout (all packages at root) — rejected because it prevents `internal/` encapsulation and makes dependency graph harder to audit.

---

## 2. CLI Framework: Cobra + Viper

**Decision**: `github.com/spf13/cobra` + `github.com/spf13/viper`.
**Rationale**: Cobra is the most widely adopted Go CLI framework (used by kubectl, Docker CLI, GitHub CLI). It provides sub-command hierarchy (`simple-cli session start|resume|list|stop|reset`), built-in help generation, persistent flags, and shell completion. Viper integrates natively with Cobra to bind flags/env vars/config file with a single precedence chain. No other framework reaches this combination of adoption, Go idiom compliance, and AI agent tooling integration via structured help text.
**Alternatives considered**:

- `urfave/cli` — less idiomatic for deep sub-command trees; weaker Viper compatibility.
- `kingpin` — abandoned / unmaintained; not suitable for long-lived projects.
- Hand-rolled `flag` package — violates Principle VIII (significantly more code for same functionality).

---

## 3. Session Persistence: File-Backed JSON + File Locking

**Decision**: Sessions stored as individual JSON files in a per-user state directory; file locks via `flock` (Linux/macOS) and `LockFileEx` (Windows) through `golang.org/x/sys`.
**Rationale**:

- Keeps the binary statically linked (`CGO_ENABLED=0`) — SQLite's `cgo` requirement is disqualifying.
- JSON files are human-inspectable (agents can read them directly per Principle IV).
- Per-session files avoid lock contention — each session has its own lock.
- XDG Base Directory Spec (`$XDG_STATE_HOME/simple-cli/`) is the correct location for mutable runtime state on Linux; `%APPDATA%` is the Windows equivalent.
- In-memory fallback satisfies Principle IX (handle read-only filesystem gracefully).

**File format**:

```json
{
  "id": "uuid-v4",
  "name": "my-session",
  "status": "active",
  "created_at": "2026-03-23T10:00:00Z",
  "updated_at": "2026-03-23T10:05:00Z",
  "state": {}
}
```

**Alternatives considered**:

- SQLite via `modernc.org/sqlite` (pure Go) — valid but adds ~3 MB to binary and requires more complex migration logic for a simple key-value store.
- BoltDB — unmaintained (last release 2017); `bbolt` fork is maintained but adds binary size for marginal benefit over JSON.
- Redis/external store — violates Principle VIII (heavyweight, requires infrastructure).

---

## 4. Cross-Platform PATH Registration

**Decision**: Platform-detected at install-time; idempotent write with existence check before appending.

### Windows (`install.ps1`)

```powershell
# User scope (no elevation required)
$current = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($current -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$current;$installDir", "User")
    # Broadcast WM_SETTINGCHANGE so running shells pick up the change
}
```

Machine scope (elevation): same via `[EnvironmentVariableTarget]::Machine`.

### Linux (`install.sh`)

```bash
# Append only if not already in the file
grep -qF "$INSTALL_DIR" ~/.bashrc || echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> ~/.bashrc
grep -qF "$INSTALL_DIR" ~/.zshrc  2>/dev/null || echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> ~/.zshrc
```

System scope (deb/rpm): `/etc/profile.d/simple-cli.sh` (idempotent by file existence).

### macOS

- Homebrew: PATH managed automatically via Homebrew `bin/` symlink.
- PKG: `/etc/paths.d/simple-cli` file (single line, idempotent by content check).
- Shell script: same as Linux, with detection of Apple Silicon path (`/opt/homebrew/bin`).

**Rationale**: Each method uses the OS-native persistence mechanism. Idempotency is enforced by string-contains check before writing. Post-install validation runs `simple-cli --version` in a sub-process with a clean `PATH` override to confirm registration.

**Alternatives considered**: `pathman` tool — adds a binary dependency; simpler to implement directly per OS.

---

## 5. AI Agent Output Format

**Decision**: `--output json` flag (or `SIMPLE_CLI_OUTPUT=json`) produces structured JSON on stdout; `--output human` (default) is human-readable text; all log output always goes to stderr.

**JSON schema** (success):

```json
{
  "status": "ok",
  "data": { ... },
  "meta": { "version": "1.2.3", "duration_ms": 42, "command": "session start" }
}
```

**JSON schema** (error):

```json
{
  "status": "error",
  "code": "RESOURCE_NOT_FOUND",
  "message": "session 'abc' not found",
  "hint": "Use 'simple-cli session list' to see available sessions.",
  "meta": { "version": "1.2.3", "duration_ms": 5, "command": "session resume" }
}
```

**Exit codes** (stable contract):

| Code | Meaning                             |
| ---- | ----------------------------------- |
| 0    | Success                             |
| 1    | General error                       |
| 2    | Misuse / invalid args               |
| 3    | Resource not found                  |
| 4    | Permission denied                   |
| 5    | Timeout / context deadline exceeded |

**Rationale**: Separating `code` (stable SCREAMING_SNAKE_CASE) from `message` (human-readable, may change) ensures agent scripts do not break on message wording changes. The `meta` envelope provides traceability without requiring log parsing. stderr-only logs allow agents to pipe stdout cleanly.

---

## 6. Graceful Shutdown

**Decision**: `os/signal.NotifyContext` with SIGINT + SIGTERM; 5-second drain timeout enforced via `context.WithTimeout`.

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
// pass ctx to all I/O operations
// on cancellation: flush session state, release locks, exit 0
```

**Rationale**: `NotifyContext` is the idiomatic Go 1.16+ approach. The 5-second drain matches Principle IX. On Windows, `os.Interrupt` maps to Ctrl-C (CTRL_C_EVENT); SIGTERM is also handled via `syscall.SIGTERM` (supported since Go 1.12 on Windows).

---

## 7. Logging: log/slog

**Decision**: `log/slog` (stdlib, Go 1.21+) with a custom handler that:

- Routes all output to stderr.
- Switches to JSON handler when `SIMPLE_CLI_OUTPUT=json` or `--output json`.
- Respects `--log-level` / `SIMPLE_CLI_LOG_LEVEL`.
- Redacts sensitive keys via `internal/security.Redact()` before logging.

**Rationale**: `slog` is now part of the standard library, eliminating an external dependency. It supports structured key-value logging and JSON output natively. Using it exclusively (Principle VI) keeps the dependency count low and the binary small.

**Alternatives considered**: `zerolog` — ~0.2 ms faster but marginal for a CLI; adds binary size and external dependency for negligible gain. `zap` — same reasoning. Both rejected per Principle VIII.

---

## 8. Build Toolchain: goreleaser

**Decision**: `goreleaser` with `.goreleaser.yml` for cross-compilation and packaging; `Makefile` for developer workflow.

**Key goreleaser targets**:

- `GOARCH=amd64,arm64` for linux, darwin; `GOARCH=amd64` for windows
- `CGO_ENABLED=0` always
- `ldflags: "-s -w -X main.Version={{ .Version }}"`
- Archives: `.tar.gz` for unix, `.zip` for windows
- Packages: `.deb`, `.rpm` via `nfpm`
- Checksums: SHA-256

**Makefile targets**:

```makefile
build        # CGO_ENABLED=0 go build ./... → dist/
test         # go test ./... -coverprofile=coverage.out
lint         # golangci-lint run
install-local# cp dist/simple-cli $(shell go env GOPATH)/bin/
test-sandbox # docker compose -f tests/sandbox/docker-compose.yml up
```

**Rationale**: goreleaser is the standard for Go release pipelines; nfpm handles .deb/.rpm without external system tools. The Makefile provides a consistent developer interface across all OSes.

---

## 9. Testing Strategy

**Unit tests**: `go test ./internal/... ./pkg/... ./cmd/...` — table-driven, no global state, each test isolated. Interfaces (e.g., `SessionStore`) enable mock injection without file system.

**Integration tests** (`tests/integration/`, build tag `integration`):

- Compile binary once via `TestMain`.
- Drive via `exec.Command` to test real CLI behavior.
- Verify JSON output schemas, exit codes, file state location.

**Sandbox tests** (`tests/sandbox/`, `make test-sandbox`):

- Docker Compose with one service per OS image.
- Each service: copy binary → run install script → validate `simple-cli --version`.
- Fail-fast on first container failure; CI reports which OS failed.

**Coverage gate**: `go tool cover -func=coverage.out` parsed by CI; fails if `internal/` + `pkg/` total < 80%.

---

## 10. Resolved Clarifications

| Item                      | Resolution                                                                                            |
| ------------------------- | ----------------------------------------------------------------------------------------------------- |
| Session ID format         | UUID v4 (via `github.com/google/uuid`)                                                                |
| Session state schema      | `map[string]any` — free-form extensible data store                                                    |
| Config file location      | `$XDG_CONFIG_HOME/simple-cli/config.yaml` (Linux/macOS); `%APPDATA%\simple-cli\config.yaml` (Windows) |
| Go minimum version        | 1.22 (slog: 1.21+; range-over-func: 1.22+)                                                            |
| Concurrent session writes | Per-session file lock; in-process mutex for in-memory fallback                                        |
| Windows installer type    | NSIS `.exe` primary; `install.ps1` as no-dependency fallback                                          |
| macOS ARM path            | `/opt/homebrew/bin` (Apple Silicon) vs `/usr/local/bin` (Intel) — auto-detected via `uname -m`        |
| AI agent `code` field     | SCREAMING_SNAKE_CASE strings; versioned contract in `contracts/output-schema.md`                      |
