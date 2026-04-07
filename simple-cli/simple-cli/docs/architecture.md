# Architecture

**simple-cli** is a cross-platform Go CLI template that stays alive as a long-running daemon process.

---

## Package Dependency Diagram

```
main
  └─► cmd/root.go               (Cobra root command, slog init, config loading)
        ├─► cmd/run.go               (daemon sub-command — blocks on SIGINT/SIGTERM)
        ├─► cmd/example_cmd.go       (template example sub-command, safe to delete)
        ├─► cmd/auth/                (auth login/logout/status sub-commands)
        │     └─► internal/auth/         (HTTPProviderAdapter: device flow, poll, refresh)
        ├─► cmd/chat.go              (chat sub-command: SSE streaming, auto-refresh)
        │     └─► internal/chat/         (SSEChatBackend: OpenAI-compatible SSE)
        ├─► internal/config/         (Config struct, ProviderConfig, context key, ConfigDir)
        ├─► internal/provider/       (ProviderAdapter, TokenStore, ChatBackend interfaces + types)
        ├─► internal/tokenstore/     (FileTokenStore: atomic JSON persistence, 0600)
        ├─► internal/output/         (Formatter interface, HumanFormatter, JSONFormatter)
        ├─► internal/exitcode/       (ExitError, exit code constants)
        ├─► internal/signals/        (graceful shutdown context with 5 s drain)
        ├─► internal/security/       (Redact helper)
        └─► pkg/version/             (build-time injectable version info)
```

**Layers** (dependency flows downward — no upward imports):

```
cmd/        — Cobra commands; depend on internal/*; may not be imported by internal/*
internal/   — private packages; may import each other within constraints below
pkg/        — public stable API; no internal/* imports
```

**Cross-cutting constraint**: `internal/` packages have no dependency on `cmd/`.

---

## Design Decisions

### 1. Daemon Process via SIGINT/SIGTERM

`simple-cli run` blocks on a context cancelled by `signals.NotifyContext`, which catches `SIGINT` and `SIGTERM`. A 5-second drain window lets in-flight work complete before `os.Exit`. On Windows, `cmd.Process.Kill()` is used by callers since child processes cannot receive SIGINT.

**Alternatives rejected**: polling loop (wastes CPU), goroutine with ticker (adds unnecessary complexity), OS service manager API (platform-specific, violates cross-platform principle).

### 2. Template Extension Pattern

To add a sub-command, create `cmd/mycommand.go` with a `newMyCmd()` constructor and register it in `cmd/root.go`. The template ships with `cmd/example_cmd.go` as a minimal reference. Delete it when customising.

### 3. Cobra + Viper (standard Go CLI stack)

Cobra provides sub-command routing with `PersistentPreRunE` hooks for shared setup. Viper handles the config precedence chain: CLI flags → env vars → config file → built-in defaults.

### 4. Structured Logging with slog (stdlib Go 1.21+)

`log/slog` routes all diagnostic output to stderr. JSON handler is used in `--output json` mode so AI agents can parse log lines as well as command output. No third-party logging library is required.

### 5. Zero-CGO Static Binary

`CGO_ENABLED=0` ensures the binary is a static executable that runs on Alpine Linux (no libc), Windows without runtime DLLs, and macOS.

### 6. Extensible Sub-Command Pattern

Each sub-command is a `func newXxxCmd() *cobra.Command` in its own file under `cmd/`. The constructor is called in `cmd/root.go init()`. This pattern isolates command logic, simplifies testing, and makes the codebase easy to navigate when customising the template.

---

## Data Flow: CLI Flag → Config → Formatter → Writer

```
User invokes: simple-cli --output json run

1. main() → cmd.Execute()
2. rootCmd.PersistentPreRunE():
     flag values → Viper → config.Load() → *Config{Output: "json", ...}
     Config stored in context via context.WithValue(ctx, config.CtxKey{}, cfg)

3. cmd/run.go RunE():
     ctx, stop := signals.NotifyContext(cmd.Context())   // blocks on SIGINT/SIGTERM
     cfg := ctx.Value(config.CtxKey{}).(*config.Config)
     w   := output.NewWriter(cfg.Quiet)                 // routes Out→stdout, Err→stderr
     f   := output.NewFormatter("json", w, ...)          // returns JSONFormatter

4. <-ctx.Done()   (blocks until shutdown signal)

5. f.FormatSuccess("run", {"status":"stopped","uptime_ms":N}, elapsed):
     → json.Marshal(SuccessResponse{status:"ok", data:{...}, meta:{...}})
     → w.Out.Write(jsonBytes + '\n')  // to stdout
     → return nil (success)

6. RunE returns nil → Cobra exits 0
```

