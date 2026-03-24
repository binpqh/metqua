# Architecture

**simple-cli** is a cross-platform Go CLI application that manages long-life persistent sessions.

---

## Package Dependency Diagram

```
main
  └─► cmd/root.go               (Cobra root command, slog init, config loading)
        ├─► cmd/session/         (session sub-command group)
        │     ├─► internal/session/   (SessionStore interface + FileStore/MemStore)
        │     ├─► internal/config/    (Config struct, context key)
        │     └─► internal/output/    (Formatter, Writer)
        ├─► internal/config/     (Config, CtxKey, ConfigDir, StateDir)
        ├─► internal/output/     (Formatter interface, HumanFormatter, JSONFormatter)
        ├─► internal/exitcode/   (ExitError, exit code constants)
        ├─► internal/signals/    (graceful shutdown context)
        ├─► internal/security/   (Redact helper)
        └─► pkg/version/         (build-time injectable version info)
```

**Layers** (dependency flows downward — no upward imports):

```
cmd/        — Cobra commands; depend on internal/*; may not be imported by internal/*
internal/   — private packages; may import each other within constraints below
pkg/        — public stable API; no internal/* imports
```

**Cross-cutting constraint**: `internal/session/` has no dependency on `cmd/`, `internal/output/`, or `internal/config/`. It depends only on `internal/session/errors.go` and stdlib.

---

## Design Decisions

### 1. File-backed Session Store (from research.md)

Sessions are stored as `<uuid>.json` files under `$XDG_STATE_HOME/simple-cli/sessions/`, with an `index.json` mapping names to IDs. File-level locking (flock/LockFileEx) ensures cross-process safety without requiring a daemon.

**Alternatives rejected**: SQLite (adds C dependency, CGO_ENABLED!=0), Redis/PostgreSQL (external service, violates lightweight requirement), memory-only (sessions don't survive restarts).

### 2. Cobra + Viper (standard Go CLI stack)

Cobra provides sub-command routing with `PersistentPreRunE` hooks for shared setup. Viper handles the config precedence chain: CLI flags → env vars → config file → built-in defaults.

### 3. Structured Logging with slog (stdlib Go 1.21+)

`log/slog` routes all diagnostic output to stderr. JSON handler is used in `--output json` mode so AI agents can parse log lines as well as command output. No third-party logging library is required.

### 4. Library-First Internal Packages

`internal/session/` is a pure library: no cobra, no slog calls, no config coupling. This makes unit testing straightforward and the package reusable independently of the CLI.

### 5. Zero-CGO Static Binary

`CGO_ENABLED=0` ensures the binary is a static executable that runs on Alpine Linux (no libc), Windows without runtime DLLs, and macOS. File locking is implemented via `golang.org/x/sys` syscalls rather than C bindings.

### 6. Dual Store Pattern (FileStore + MemStore fallback)

The `ProxyStore` in `internal/session/proxystore.go` holds a `*SessionStore` pointer that is set during `PersistentPreRunE`. If the state directory is not writable, `MemStore` is used. This lets the binary run read-only commands (`--help`, `--version`) without initialising the store.

---

## Data Flow: CLI Flag → Config → Formatter → Writer

```
User invokes: simple-cli --output json session list --status active

1. main() → cmd.Execute()
2. rootCmd.PersistentPreRunE():
     flag values → Viper → config.Load() → *Config{Output: "json", ...}
     Config stored in context via context.WithValue(ctx, config.CtxKey{}, cfg)
     FileStore initialised, set into ProxyStore

3. cmd/session/list.go RunE():
     cfg := ctx.Value(config.CtxKey{}).(*config.Config)
     w   := output.NewWriter(cfg.Quiet)          // routes Out→stdout, Err→stderr
     f   := output.NewFormatter("json", w, ...)  // returns JSONFormatter

4. store.List(ctx, &StatusActive) → []*Session

5. f.FormatSuccess("session list", listData{...}, elapsed):
     → json.Marshal(SuccessResponse{status:"ok", data:{sessions:[...], total:1}, meta:{...}})
     → w.Out.Write(jsonBytes + '\n')              // to stdout
     → return nil (success)

6. RunE returns nil → Cobra exits 0
```

---

## Session State Directory Layout

```
$XDG_STATE_HOME/simple-cli/          (Linux/macOS: ~/.local/state/simple-cli)
%APPDATA%\simple-cli\                 (Windows)
├── sessions/
│   ├── <uuid1>.json                  session data
│   ├── <uuid1>.json.lock             advisory file lock
│   ├── <uuid2>.json
│   └── <uuid2>.json.lock
├── index.json                        {name → id} mapping
└── index.json.lock                   advisory lock for index writes
```

---

## Cross-Platform File Locking

| OS      | Mechanism                                                      | File              |
| ------- | -------------------------------------------------------------- | ----------------- |
| Linux   | `flock(2) LOCK_EX LOCK_NB`                                     | `lock_unix.go`    |
| macOS   | `flock(2) LOCK_EX LOCK_NB`                                     | `lock_unix.go`    |
| Windows | `LockFileEx LOCKFILE_EXCLUSIVE_LOCK LOCKFILE_FAIL_IMMEDIATELY` | `lock_windows.go` |

Both paths share the same `FileLock` API (`Lock(ctx context.Context)`, `Unlock()`). The context deadline is respected via a 50ms polling loop that checks `ctx.Done()`.
