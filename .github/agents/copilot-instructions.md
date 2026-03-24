# simple-cli Development Guidelines

Auto-generated from feature plan `001-cli-long-life-session`. Last updated: 2026-03-23

## Active Technologies

| Category | Technology | Notes |
|----------|-----------|-------|
| Language | Go 1.22+ | Min version; CI on current + previous stable |
| CLI Framework | `github.com/spf13/cobra` v1.8+ | Sub-commands: `simple-cli <noun> <verb>` |
| Config | `github.com/spf13/viper` v1.18+ | Flags > env vars > config file > defaults |
| Storage | File-backed JSON + file locks | `$XDG_STATE_HOME/simple-cli/` or `%APPDATA%\simple-cli\` |
| Logging | `log/slog` (stdlib, Go 1.21+) | stderr only; JSON-Lines in `--output json` mode |
| Testing | `testify` + `gomock` + `testcontainers-go` | Unit + integration + sandbox |
| Build | `goreleaser` + `Makefile` | `CGO_ENABLED=0`; static binaries |
| Linting | `golangci-lint` with `.golangci.yml` | `errcheck`, `gosec`, `revive`, `gofumpt` |
| CI/CD | GitHub Actions | `ubuntu-latest`, `windows-latest`, `macos-latest` matrix |

## Project Structure

```text
simple-cli/
├── cmd/                  # Cobra entry points (root.go + session/ sub-commands)
├── internal/
│   ├── config/           # Viper-backed config loader
│   ├── session/          # Session entity, SessionStore interface, filestore, memstore, lock
│   ├── output/           # --output json / human formatter
│   ├── security/         # Token/password redaction helpers
│   └── signals/          # SIGINT/SIGTERM graceful shutdown
├── pkg/version/          # Version string (ldflags-injected)
├── scripts/install/      # install.sh (POSIX) + install.ps1 (PowerShell)
├── installer/            # NSIS setup.nsi (Windows) + PKG postinstall (macOS)
├── tests/
│   ├── integration/      # Black-box CLI tests via exec.Command
│   └── sandbox/          # Docker Compose: Ubuntu, Debian, Alpine, Windows LTSC
├── docs/                 # installation, quickstart, configuration, architecture, ai-agent-guide
├── .goreleaser.yml       # Cross-platform build + deb/rpm/pkg artifacts
├── .golangci.yml         # Linter config
└── Makefile              # build, test, lint, install-local, test-sandbox
```

## Commands

```makefile
make build          # CGO_ENABLED=0 go build ./... → dist/
make test           # go test ./... -coverprofile=coverage.out  (≥80% gate)
make lint           # golangci-lint run
make install-local  # cp dist/simple-cli $(GOPATH)/bin/
make test-sandbox   # docker compose -f tests/sandbox/docker-compose.yml up
```

```sh
# CLI commands
simple-cli session start [--name <name>]
simple-cli session resume [--name <name> | --id <id>]
simple-cli session list [--status active|paused|stopped]
simple-cli session stop [--name <name> | --id <id>]
simple-cli session reset [--name <name> | --id <id>] [--force]
```

## Code Style (Go)

- `gofmt` + `goimports` always; enforced by CI
- Errors: `fmt.Errorf("operation: %w", err)` — never ignore, never leak stack traces to users
- `context.Context` as first parameter on all I/O functions
- Table-driven tests preferred; each test independently runnable (`go test -run TestName`)
- No `init()` except flag registration in `cmd/`; no `panic` in library code
- `log/slog` only — no third-party logging libraries
- All exported symbols have GoDoc comments; complex algorithms carry rationale comments
- Cobra `Short` + `Long` + at least one `Example` on every command

## AI Agent Output Contract

- `--output json` → stdout: `{"status":"ok","data":{...},"meta":{"version":"...","duration_ms":42}}`
- Errors → stderr JSON-Line: `{"level":"error","msg":"...","code":"SESSION_NOT_FOUND"}`
- Exit codes: `0` success, `1` general, `2` invalid args, `3` not found, `4` permission denied, `5` timeout
- `SIMPLE_CLI_OUTPUT=json` env var equivalent to `--output json`

## Recent Changes

- **001-cli-long-life-session**: Initial feature — long-life session management, cross-platform installers, AI agent I/O, sandbox testing

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
