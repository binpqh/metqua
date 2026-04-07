# Implementation Plan: OAuth Provider + AI Chat Completion via CLI

**Branch**: `003-oauth-ai-chat-provider` | **Date**: 2026-03-24 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/003-oauth-ai-chat-provider/spec.md`

---

## Summary

Add provider-agnostic OAuth 2.0 Device Flow authentication and OpenAI-compatible streaming AI chat to `simple-cli`. Four new sub-commands (`auth login`, `auth logout`, `auth status`, `chat`) are backed by three new `internal/` packages with clean interfaces, zero GitHub-specific coupling, and full `--output json` support for AI agent workflows.

---

## Technical Context

**Language/Version**: Go 1.25.0 (module `github.com/binpqh/simple-cli`)
**Primary Dependencies**:

- Existing: `github.com/spf13/cobra` v1.10.2, `github.com/spf13/viper` v1.21.0, `github.com/stretchr/testify` v1.11.1
- New: **none** — device flow (RFC 8628) and SSE streaming are implemented with stdlib `net/http`, `bufio`, `encoding/json`. See research.md §1 and §2 for rationale.

**Storage**: `$XDG_CONFIG_HOME/simple-cli/tokens.json` (Unix, mode 0600) / `%APPDATA%\simple-cli\tokens.json` (Windows via `ConfigDir()`). Interface-backed for future OS-keychain swap.

**Testing**: `testing` stdlib + `github.com/stretchr/testify`; HTTP interactions mocked via `net/http/httptest`; no network I/O in unit tests.

**Target Platform**: Linux, macOS, Windows (CGO_ENABLED=0 static binary — unchanged from feature 002).

**Project Type**: CLI application (template variant).

**Performance Goals**:

- First streamed AI character: ≤3 s after HTTP response headers received (SC-003).
- Auth flow latency: ≤2 min after browser completion (SC-002).
- Binary size: remains <20 MB after ldflags strip (SC-008 / Constitution §VIII).

**Constraints**:

- Startup time <150 ms (cold, no network) — unchanged constitution gate.
- Token file permissions: 0600 Unix / restricted ACL Windows.
- No new external Go module dependencies (see research.md §1).
- No GitHub-specific constants anywhere in the codebase (FR-017).

**Scale/Scope**: Single-user CLI; no concurrent request handling required at the binary level.

---

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-checked post-design below._

| Principle                     | Status  | Notes                                                                                                     |
| ----------------------------- | ------- | --------------------------------------------------------------------------------------------------------- |
| I Library-First               | ✅ PASS | `internal/auth`, `internal/chat`, `internal/provider`, `internal/tokenstore` — each single-responsibility |
| II Idiomatic Go               | ✅ PASS | context-first I/O, errors wrapped, no `init()` in new libs                                                |
| III Cross-Platform Installers | ✅ PASS | additive feature; no installer changes needed                                                             |
| IV AI Agent Interoperability  | ✅ PASS | all 4 new commands support `--output json`; stdout payload only, logs to stderr                           |
| V Test-First                  | ✅ PASS | unit tests alongside each new package; httptest mocks for HTTP; ≥80% coverage gate                        |
| VI Observability              | ✅ PASS | slog at `debug` for OAuth steps and chat calls; token values redacted via `internal/security`             |
| VIII Lightweight Binary       | ✅ PASS | no new dependencies; estimated binary growth <200 KB                                                      |
| IX Robustness                 | ✅ PASS | all HTTP calls honour `context.Context`; device flow 300 s timeout (FR-018); 401 → auto-refresh-and-retry |
| X Versioning                  | ✅ PASS | additive feature; SemVer MINOR bump (v2.1.0); no breaking changes                                         |

**Post-design re-check**: `internal/provider` depends on `internal/tokenstore` (via interface only — no concrete import). `internal/auth` depends on `internal/provider` interface. `internal/chat` depends on `internal/tokenstore` (to read token). No circular imports introduced.

**Violations requiring justification**: None.

---

## Project Structure

### Documentation (this feature)

```text
specs/003-oauth-ai-chat-provider/
├── plan.md              ← this file
├── research.md          ← Phase 0 output
├── data-model.md        ← Phase 1 output
├── quickstart.md        ← Phase 1 output
├── contracts/
│   ├── auth-output.md   ← JSON envelope schemas for auth commands
│   ├── chat-output.md   ← JSON envelope schema for chat command
│   ├── provider-config.md ← config.yaml providers: section schema
│   └── interfaces.md    ← Go interface definitions (ProviderAdapter, ChatBackend, TokenStore)
└── tasks.md             ← Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/
├── auth/
│   ├── auth.go          NEW — Cobra group command (newAuthCmd), registers sub-commands
│   ├── login.go         NEW — auth login sub-command (device flow)
│   ├── logout.go        NEW — auth logout sub-command
│   └── status.go        NEW — auth status sub-command
├── chat.go              NEW — chat sub-command (SSE streaming)
├── example_cmd.go       EXISTING (unchanged)
├── root.go              UPDATE — register newAuthCmd() + newChatCmd()
└── run.go               EXISTING (unchanged)

internal/
├── auth/
│   ├── auth.go          NEW — DeviceFlow(), RefreshToken(), BuildAuthHeader()
│   └── auth_test.go     NEW
├── chat/
│   ├── chat.go          NEW — SendMessage(), StreamSSE(), BuildChatRequest()
│   └── chat_test.go     NEW
├── config/
│   ├── config.go        UPDATE — add ProviderConfig struct, Providers map, DefaultProvider
│   └── config_test.go   UPDATE — add provider config tests
├── provider/
│   ├── provider.go      NEW — ProviderAdapter + TokenStore + ChatBackend interfaces
│   └── provider_test.go NEW — interface contract verification tests
├── tokenstore/
│   ├── tokenstore.go    NEW — FileTokenStore: Get/Set/Delete/List
│   └── tokenstore_test.go NEW
├── exitcode/            EXISTING (unchanged; codes 3/4/5 already present)
├── output/              EXISTING (unchanged)
├── security/            EXISTING (unchanged; Redact used for token logging)
└── signals/             EXISTING (unchanged)

docs/
├── customization.md     NEW — extension guide (US3, FR-014)
├── ai-agent-guide.md    UPDATE — add chat + auth command sections
├── configuration.md     UPDATE — add providers: config key reference
└── architecture.md      UPDATE — add new package nodes
```

**Structure Decision**: Single Go module; new packages added under existing `internal/` convention. Command group `cmd/auth/` mirrors the `simple-cli auth <sub>` invocation tree.

---

## Complexity Tracking

No constitution violations requiring justification.
