# Tasks: OAuth Provider + AI Chat Completion via CLI

**Feature**: `003-oauth-ai-chat-provider` | **Branch**: `003-oauth-ai-chat-provider`
**Input**: `specs/003-oauth-ai-chat-provider/` — plan.md, spec.md, research.md, data-model.md, contracts/
**Version target**: v2.1.0 (MINOR bump — additive, no breaking changes)

## Format: `[ID] [P?] [Story?] Description — file path`

- **[P]**: Can run in parallel (different file, no blocking dependency)
- **[US1/US2/US3]**: User story this task belongs to
- All paths relative to repository root (`github.com/binpqh/simple-cli`)

---

## Phase 1: Setup

**Purpose**: Create directory scaffolding and gitignore guard before any code is written.

- [x] T001 Create package directories with placeholder `.keep` files: `cmd/auth/`, `internal/auth/`, `internal/chat/`, `internal/provider/`, `internal/tokenstore/`
- [x] T002 [P] Add `tokens.json` entry to `.gitignore` to prevent accidental token commits

**Checkpoint**: Directory structure exists; `go build ./...` still passes.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared types, interfaces, config extension, and FileTokenStore. ALL must be done before any user story implementation begins.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T003 Implement `internal/provider/provider.go`: declare `ProviderAdapter`, `TokenStore`, `ChatBackend` interfaces; declare `DeviceFlowState`, `TokenSet` (with `IsExpired()`), `ChatRequest`, `ChatMessage`, `StreamEvent` structs; declare sentinel errors `ErrDeviceExpired`, `ErrAuthDenied`, `ErrNoRefreshToken`, `ErrProviderNotFound`, `ErrTokenNotFound`
- [x] T004 [P] Extend `internal/config/config.go`: add `ProviderConfig` struct (`ClientID`, `DeviceEndpoint`, `TokenEndpoint`, `ChatEndpoint`, `Scopes []string`, `DefaultModel` — all `mapstructure` tagged); add `DefaultProvider string` and `Providers map[string]ProviderConfig` fields to the main `Config` struct; add `ActiveProvider(name string) (*ProviderConfig, error)` helper that resolves `DefaultProvider` when name is empty
- [x] T005 [P] Implement `internal/tokenstore/tokenstore.go`: `FileTokenStore` struct with `path string` field; implement `TokenStore` interface (`Get`, `Set`, `Delete`); persist as `{"providers": {...}}` JSON; write with `os.WriteFile(path, data, 0600)`; recover from corrupt file by deleting and returning `ErrTokenNotFound`
- [x] T006 [P] Write `internal/tokenstore/tokenstore_test.go`: use `t.TempDir()` as token file path; test `Get` on missing file returns `ErrTokenNotFound`; test `Set`+`Get` round-trip preserves all `TokenSet` fields including `Expiry`; test `Delete` is idempotent; test corrupt JSON in file causes clean recovery; test multiple providers coexist in same file
- [x] T007 [P] Write `internal/provider/provider_test.go`: compile-time interface satisfaction assertions (`var _ provider.TokenStore = (*tokenstore.FileTokenStore)(nil)` etc.); test sentinel error identity (`errors.Is`); test `TokenSet.IsExpired()` with past and future expiry times
- [ ] T008 [P] Update `internal/config/config_test.go`: add test cases for full `ProviderConfig` YAML unmarshal via Viper; multiple providers in one config; `ActiveProvider` returns correct entry; `ActiveProvider` with empty name falls back to `DefaultProvider`; missing `DefaultProvider` key returns error

**Checkpoint**: `go build ./...` and `go test ./internal/provider/... ./internal/tokenstore/... ./internal/config/...` all pass.

---

## Phase 3: User Story 1 — Authenticate with an OAuth Provider (Priority: P1) 🎯 MVP

**Goal**: `auth login` → browser approval → token stored; `auth status` shows state; `auth logout` clears tokens. Automatic token refresh on 401.

**Independent Test**: Run `auth login` against a local `httptest` server that simulates device flow approval → confirm `tokens.json` written → run `auth status` shows valid token → run `auth logout` → `tokens.json` empty.

### Tests for User Story 1

- [x] T009 [P] [US1] Write `internal/auth/auth_test.go`: use `httptest.NewServer` to mock device and token endpoints; test `StartDeviceFlow` happy path returns populated `DeviceFlowState`; test `StartDeviceFlow` on HTTP 500 returns wrapped error; test `PollToken` retries on `authorization_pending`; test `PollToken` returns `ErrDeviceExpired` on `expired_token`; test `PollToken` returns `ErrAuthDenied` on `access_denied`; test `PollToken` happy path returns `TokenSet` with `Expiry` computed from `expires_in`; test `RefreshToken` happy path; test `RefreshToken` returns `ErrNoRefreshToken` when provider omits `refresh_token` grant; test `ctx` cancellation aborts polling

### Implementation for User Story 1

- [x] T010 [US1] Implement `internal/auth/auth.go`: `HTTPProviderAdapter` struct embedding `*config.ProviderConfig` and `*http.Client`; `NewHTTPProviderAdapter(cfg *config.ProviderConfig) *HTTPProviderAdapter`; `StartDeviceFlow(ctx)` POSTs `client_id`+scopes to `DeviceEndpoint`, decodes `DeviceAuthResponse`, returns `DeviceFlowState` with `ExpiresAt = now + expires_in` (default 300s), `Interval` (default 5s); `PollToken(ctx, state)` loops on `time.NewTicker(state.Interval)`, POSTs device_code grant to `TokenEndpoint`, handles all RFC 8628 §3.5 error codes, returns `*provider.TokenSet` on success; `RefreshToken(ctx, rt)` POSTs refresh_token grant and returns new `*provider.TokenSet`; all token fields redacted from `slog.Debug` calls using `internal/security.Redact`
- [x] T011 [P] [US1] Implement `cmd/auth/auth.go`: `newAuthCmd() *cobra.Command` returning a group command with short description; add persistent `--provider` string flag defaulting to `""`; register `newLoginCmd()`, `newLogoutCmd()`, `newStatusCmd()` as sub-commands
- [x] T012 [P] [US1] Implement `cmd/auth/login.go`: `newLoginCmd() *cobra.Command`; resolve provider via `--provider` flag or `cfg.DefaultProvider` using `cfg.ActiveProvider()`; validate with `config.ValidateProviderConfig`; construct `FileTokenStore`; if token already exists prompt `"already authenticated as %s; re-authenticate? [y/N]"` in human mode or return JSON error in json mode; call `auth.NewHTTPProviderAdapter(pc).StartDeviceFlow(ctx)`; print `VerificationURI` and `UserCode`; call `PollToken(ctx, state)` and store result via `tokenstore.Set`; print success human message or JSON envelope per `contracts/auth-output.md`
- [x] T013 [P] [US1] Implement `cmd/auth/logout.go`: `newLogoutCmd() *cobra.Command`; support `--all` bool flag; for each target provider call `tokenstore.Delete`; output per `contracts/auth-output.md`; exit 0 even if token was absent (idempotent)
- [x] T014 [P] [US1] Implement `cmd/auth/status.go`: `newStatusCmd() *cobra.Command`; call `tokenstore.Get(ctx, provider)`; compute `Expired` via `ts.IsExpired()`; compute human-readable time-until-expiry string; output per `contracts/auth-output.md` for both human and json modes; exit 0 regardless of login state (status is informational)
- [x] T015 [US1] Add `ValidateProviderConfig(pc *config.ProviderConfig) error` to `internal/config/config.go`: return descriptive error when `ClientID` is empty; return error when any of `DeviceEndpoint`, `TokenEndpoint`, `ChatEndpoint` is not a valid `https://` URL; used by login and chat commands before making any HTTP calls
- [x] T016 [P] [US1] Update `internal/config/config_test.go` with `ValidateProviderConfig` test cases: valid config passes; empty `client_id` fails; HTTP (not HTTPS) endpoint fails; valid HTTPS endpoints pass
- [x] T017 [US1] Register `newAuthCmd()` in `cmd/root.go` `init()` function: `rootCmd.AddCommand(newAuthCmd())`

**Checkpoint**: `go build ./...` passes; `go test ./internal/auth/... ./internal/tokenstore/... ./cmd/...` passes; `simple-cli auth --help` shows login/logout/status sub-commands.

---

## Phase 4: User Story 2 — Chat with AI via CLI (Priority: P2)

**Goal**: Authenticated user runs `simple-cli chat "message"` and receives streamed AI response; `--output json` produces stable envelope; 401 triggers auto-refresh and single retry.

**Independent Test**: With a valid token in `tokens.json`, run `chat "hello"` against a local `httptest` server that returns 3 SSE `data:` events then `data: [DONE]` — confirm all delta content flushed to stdout; run with `--output json` and verify the envelope matches `contracts/chat-output.md`.

### Tests for User Story 2

- [x] T018 [P] [US2] Write `internal/chat/chat_test.go`: construct `SSEChatBackend` with `httptest.NewServer`; test happy path: 3 SSE chunk events + `[DONE]` → channel yields 3 `StreamEvent{Delta}` then `{Done: true}`; test SSE malformed JSON line is skipped; test HTTP 401 response returns `StreamEvent{Err: ErrUnauthorized}`; test `ctx` cancellation closes channel; test `ChatRequest` wire format (JSON body has `model`, `messages`, `stream: true`); test `conversation_id` only present in body when non-empty; test `Authorization: Bearer <token>` header is sent

### Implementation for User Story 2

- [x] T019 [US2] Implement `internal/chat/chat.go`: `SSEChatBackend` struct with `*http.Client` and `*config.ProviderConfig`; `NewSSEChatBackend(cfg *config.ProviderConfig) *SSEChatBackend`; `Chat(ctx, req, token) (<-chan provider.StreamEvent, error)`: POST JSON body to `ChatEndpoint`, set `Authorization: Bearer <token>` and `Accept: text/event-stream`; on non-2xx return `ErrUnauthorized` (401) or wrapped HTTP error; stream body via `bufio.Scanner`, extract `data: ` prefix lines, skip empty lines, parse JSON delta via `SSEChunk`, emit `StreamEvent{Delta: content}` per chunk, emit `StreamEvent{Done: true}` on `[DONE]`, close channel; redact token from all slog output
- [x] T020 [US2] Add `ErrUnauthorized` sentinel to `internal/provider/provider.go`: `var ErrUnauthorized = errors.New("request unauthorized: token expired or invalid")`
- [x] T021 [US2] Implement `cmd/chat.go`: `newChatCmd() *cobra.Command`; flags: `--provider string`, `--model string` (short `-m`), `--conversation string` (short `-c`), `--system string`; resolve provider, validate with `ValidateProviderConfig`; get token via `FileTokenStore.Get`; if `ts.IsExpired()` attempt refresh via `HTTPProviderAdapter.RefreshToken` and update store; build `ChatRequest` (append system message if `--system` provided, append user message); call `SSEChatBackend.Chat(ctx, req, token)`; human mode: range channel, print each `Delta` immediately with `fmt.Fprint(os.Stdout, e.Delta)`; json mode: collect all deltas into `content` string, emit JSON envelope after `Done`; on `StreamEvent{Err: ErrUnauthorized}` refresh token once and retry entire `Chat` call; on second 401 print error and exit 1
- [x] T022 [US2] Register `newChatCmd()` in `cmd/root.go` `init()` function: `rootCmd.AddCommand(newChatCmd())`
- [x] T023 [P] [US2] Write edge-case test for `cmd/chat.go` auto-refresh path: mock auth endpoint returns 401 on first call, token refresh succeeds, second call succeeds — verify final output contains response content and `tokens.json` is updated with refreshed token
- [x] T024 [US2] Add `--model` flag env override: bind `SIMPLE_CLI_PROVIDER_CHAT_ENDPOINT`, `SIMPLE_CLI_PROVIDER_CLIENT_ID` etc. via Viper `BindEnv` calls in `cmd/chat.go` and `cmd/auth/login.go` per `contracts/provider-config.md` env var table

**Checkpoint**: `go build ./...` passes; `go test ./internal/chat/... ./cmd/...` passes; `simple-cli chat --help` shows all flags; `simple-cli chat "hi"` (with valid config + token) streams a response.

---

## Phase 5: User Story 3 — Customize and Extend the Template (Priority: P3)

**Goal**: `docs/customization.md` enables a developer to add a new sub-command, OAuth provider adapter, or chat backend in under 30 minutes without modifying core files.

**Independent Test**: A developer follows `docs/customization.md` "Add a new OAuth provider" section and creates `internal/auth/mockprovider/mockprovider.go` implementing `provider.ProviderAdapter`. Confirm `var _ provider.ProviderAdapter = (*MockProvider)(nil)` compiles. Run `simple-cli auth login --provider mock` (with mock registered) and verify login flow routes to mock adapter.

### Implementation for User Story 3

- [x] T025 [US3] Create `docs/customization.md`: section (a) — adding a new sub-command (copy `example_cmd.go`, register in `root.go`); section (b) — implementing `provider.ProviderAdapter` for a new OAuth provider (create `internal/auth/<name>/<name>.go`, add compile-time interface check, add to `providerRegistry` map in `cmd/auth/login.go`); section (c) — implementing `provider.ChatBackend` for a new AI backend (create `internal/chat/<name>/<name>.go`, register in `cmd/chat.go`); section (d) — config file reference (`providers:` YAML key, env variable overrides); reference `contracts/interfaces.md` and `contracts/provider-config.md`
- [x] T026 [P] [US3] Update `docs/configuration.md`: add `providers:` config key table with `client_id`, `device_endpoint`, `token_endpoint`, `chat_endpoint`, `scopes`, `default_model`; add `default_provider` key; add env variable override table from `contracts/provider-config.md`; add token file location note
- [x] T027 [P] [US3] Update `docs/ai-agent-guide.md`: add `auth login`, `auth status`, `auth logout` JSON examples with `jq` processing; add `chat --output json` envelope example with `jq '.data.content'` extraction; add note on piping stdin to `chat` for file review workflows
- [x] T028 [P] [US3] Update `docs/architecture.md`: add new package nodes (`internal/provider`, `internal/auth`, `internal/chat`, `internal/tokenstore`) to architecture diagram; show dependency arrows (cmd/auth → internal/auth → internal/provider; cmd/chat → internal/chat → internal/provider; both → internal/tokenstore)
- [x] T029 [US3] Add provider adapter registry pattern to `cmd/auth/login.go`: `providerRegistry map[string]func(*config.ProviderConfig) provider.ProviderAdapter` with `"default"` key → `auth.NewHTTPProviderAdapter`; resolve adapter by checking `--provider` name in registry (falling back to `"default"`); document the registry as the single extension point

**Checkpoint**: `docs/customization.md` is readable standalone; `go build ./...` still passes; all new docs files lint cleanly.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Version bump, coverage gates, linting, and release prep.

- [x] T030 [P] Add `slog.Debug` calls in `cmd/auth/login.go` for each polling attempt (provider, attempt number, interval), `cmd/chat.go` for each SSE chunk count — no token values logged (use `security.Redact` on any token-adjacent struct fields)
- [x] T031 Bump version constant to `v2.1.0` in the version declaration file (locate with `grep -r "v2.0.0" cmd/`)
- [x] T032 [P] Run `go vet ./...` and `staticcheck ./...` (or `golangci-lint run ./...`) — fix all reported issues
- [x] T033 Run `go test -coverprofile=coverage.out ./internal/auth/... ./internal/chat/... ./internal/provider/... ./internal/tokenstore/...` and verify each package meets ≥80% coverage per SC-005; add coverage gaps as focused tests if needed
- [x] T034 [P] Update `CHANGELOG.md` (or create if absent): add `## v2.1.0` section listing new commands (`auth login/logout/status`, `chat`), new packages, new config keys, and link to `docs/customization.md`
- [x] T035 Run `go build -ldflags="-s -w" -o simple-cli .` and confirm binary is <20 MB (SC-008); run `simple-cli --version` to confirm output shows `v2.1.0`
- [x] T036 Validate `quickstart.md` end-to-end: configure a real or mock provider → `auth login` → `auth status` → `chat "hello"` → `auth logout`; confirm all outputs match `contracts/auth-output.md` and `contracts/chat-output.md`

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)      → no dependencies; start immediately
Phase 2 (Foundational) → requires Phase 1 completion; BLOCKS all user stories
Phase 3 (US1)        → requires Phase 2 completion
Phase 4 (US2)        → requires Phase 2 completion; integrates US1 token storage
Phase 5 (US3)        → requires Phase 3 + Phase 4 completion (documents their patterns)
Phase 6 (Polish)     → requires all prior phases complete
```

### User Story Dependencies

| Story           | Depends On                          | Can Start  |
| --------------- | ----------------------------------- | ---------- |
| US1 — Auth      | Phase 2                             | After T008 |
| US2 — Chat      | Phase 2 + US1 (token storage)       | After T017 |
| US3 — Customize | US1 + US2 (for documented patterns) | After T029 |

### Within Each Story

- Test tasks (`*_test.go`) before implementation to identify expected behavior
- `internal/` implementation before `cmd/` wrappers
- Foundation (`internal/config`, `internal/tokenstore`) before auth/chat packages
- Commands before root registration
- Sub-command registration before CLI smoke test

---

## Parallel Execution Examples

### Phase 2 parallel run (after T003)

```
T004  ← internal/config/config.go  (ProviderConfig)
T005  ← internal/tokenstore/tokenstore.go  (FileTokenStore)
T006  ← internal/tokenstore/tokenstore_test.go
T007  ← internal/provider/provider_test.go
T008  ← internal/config/config_test.go update
# All 5 can proceed in parallel once T003 interfaces are defined
```

### Phase 3 parallel run (within US1, after T010)

```
T011  ← cmd/auth/auth.go (group command)
T012  ← cmd/auth/login.go
T013  ← cmd/auth/logout.go
T014  ← cmd/auth/status.go
# All 4 cmd files operate in separate files; no intra-file dependencies
```

### Phase 4 parallel run (within US2, after T019)

```
T021  ← cmd/chat.go implementation
T023  ← cmd/chat test (auto-refresh path)
# T022 (register) must follow T021
```

---

## Implementation Strategy

### MVP Scope (deliver to stakeholders first)

Complete **Phase 1 → Phase 2 → Phase 3 (US1)** as the minimum viable delivery:

- `simple-cli auth login` → stores token
- `simple-cli auth status` → shows token state
- `simple-cli auth logout` → clears token

This is independently deployable and validates the provider config + token store architecture before the chat layer is built.

### Incremental Delivery

1. **MVP**: US1 complete — authentication works end-to-end
2. **Core value**: US2 complete — chat with AI from CLI, streaming output
3. **Template value**: US3 complete — customization guide published

### Task Count Summary

| Phase         | Tasks        | User Story |
| ------------- | ------------ | ---------- |
| Setup         | T001–T002    | —          |
| Foundational  | T003–T008    | —          |
| US1 Auth      | T009–T017    | US1 (P1)   |
| US2 Chat      | T018–T024    | US2 (P2)   |
| US3 Customize | T025–T029    | US3 (P3)   |
| Polish        | T030–T036    | —          |
| **Total**     | **36 tasks** |            |
