# Feature Specification: OAuth Provider + AI Chat Completion via CLI

**Feature Branch**: `003-oauth-ai-chat-provider`
**Created**: 2026-03-24
**Status**: Implemented — v2.1.0
**Input**: User description: "Source code easily customizable with guidelines to modify; OAuth with GitHub Copilot or any provider; self-hosted API service (OAuth + Chat with Copilot); simple-cli communicates with that API for AI chat completions via CLI because Agent works closely with CLI."

---

## Clarifications

### Session 2026-03-24

- Q: Is the OAuth provider the user's own self-hosted server, or the official GitHub OAuth App / GitHub Copilot API directly? → A: The user's own OAuth 2.0 server — all endpoint URLs (device, token, chat) are configurable; no GitHub-specific code is hardcoded. GitHub Copilot is an internal concern of the user's API service, not of the CLI.
- Q: What JSON contract does the chat API endpoint use? → A: OpenAI-compatible — `POST /v1/chat/completions` with a `messages` array and `stream: true`; SSE response stream with `data: [DONE]` termination.
- Q: Should the CLI auto-persist conversation IDs between `chat` invocations, or treat each call as stateless unless `--conversation` is explicitly passed? → A: Stateless — each call is independent; no conversation ID is ever written to disk; multi-turn is opt-in via explicit `--conversation <id>` flag only.
- Q: What is the maximum wait time for the OAuth 2.0 device authorization flow before the CLI times out? → A: Respect the provider's `expires_in` field; default to 300 seconds if the field is absent; use the provider's `interval` value for polling (default 5 s); exit with code 1 and a clear message on expiry.
- Q: What streaming protocol does the chat API use? → A: Server-Sent Events (SSE) with the OpenAI protocol — `data: {...JSON...}` lines terminated by `data: [DONE]`; the CLI reads the response body line-by-line and flushes `choices[0].delta.content` to stdout immediately.

---

## Overview

This feature transforms `simple-cli` from a bare daemon template into a **provider-agnostic OAuth + AI chat CLI**. The codebase ships three new sub-commands (`auth login`, `auth logout`, `auth status`, `chat`) and a provider extension pattern that lets any developer swap OAuth providers and AI backends by editing a single configuration file and, optionally, a single provider adapter file.

The primary motivating use case is: a developer has a self-hosted API service that handles OAuth authentication and AI chat completion (backed by GitHub Copilot or another model). They want to call that API from the terminal using a fast, structured CLI — especially for AI agent workflows that drive the CLI programmatically.

---

## User Scenarios & Testing

### User Story 1 — Authenticate with an OAuth Provider (Priority: P1)

A developer runs the CLI for the first time and needs to authenticate against their OAuth provider. The CLI initiates the OAuth 2.0 Device Authorization Flow: it displays a short user code and a verification URL, waits while the user completes browser authentication, then stores the resulting access token locally. From this point forward, every subsequent command is automatically authenticated.

**Why this priority**: All other features (chat, token refresh) depend on a stored token. Without working authentication nothing else can function.

**Independent Test**: Run `simple-cli auth login`, open the printed URL in a browser (or mock the device endpoint), observe the CLI exchange the device code for a token, and confirm `simple-cli auth status` shows the authenticated provider and token expiry.

**Acceptance Scenarios**:

1. **Given** the CLI is configured with a valid provider endpoint and client ID, **When** the user runs `simple-cli auth login`, **Then** the CLI prints a device code + verification URL and blocks waiting for browser completion.
2. **Given** the user completes browser authentication, **When** the device polling interval elapses, **Then** the CLI stores the access token and refresh token locally and exits with code 0.
3. **Given** valid credentials are stored, **When** the user runs `simple-cli auth status`, **Then** the CLI prints provider name, token expiry time, and whether the token is still valid.
4. **Given** the user runs `simple-cli auth logout`, **When** the command completes, **Then** stored tokens are deleted and `auth status` reports "not authenticated".
5. **Given** a stored access token has expired, **When** any authenticated command is run, **Then** the CLI automatically attempts token refresh before proceeding; if refresh fails, it instructs the user to run `simple-cli auth login` again.

---

### User Story 2 — Chat with AI via CLI (Priority: P2)

An authenticated developer runs `simple-cli chat "explain context windows"` and receives a streamed AI response in their terminal. In JSON output mode, the response is a stable envelope suitable for AI agent pipelines. Multi-turn conversation is supported within a single session flag.

**Why this priority**: This is the core value proposition of the application once authentication works. It is the command that agents and developers call repeatedly.

**Independent Test**: With a valid token stored, run `simple-cli chat "hello"` pointing to the configured API endpoint and observe the response printed to stdout. Run with `--output json` and pipe to `jq` to validate the envelope.

**Acceptance Scenarios**:

1. **Given** the user is authenticated, **When** they run `simple-cli chat "What is Go?"`, **Then** the CLI calls the configured chat API endpoint and streams the response text to stdout incrementally (human mode) or as a single JSON envelope (JSON mode).
2. **Given** `--output json` is set, **When** the chat command completes, **Then** stdout contains a valid JSON envelope with `status: "ok"`, `data.content` (the AI response text), and the standard `meta` object.
3. **Given** the user passes `--conversation <id>`, **When** the CLI calls the API, **Then** the conversation ID is sent as part of the request, enabling multi-turn context on the server side.
4. **Given** the API returns an error (e.g., rate limited, model unavailable), **When** the response is received, **Then** the CLI exits with a non-zero code and writes a JSON error envelope to stderr (in json mode) or a human-readable error (in human mode).
5. **Given** the access token has expired mid-session, **When** the chat API returns 401, **Then** the CLI automatically refreshes the token and retries the request once before failing.

---

### User Story 3 — Customize and Extend the Template (Priority: P3)

A developer who wants to add a new OAuth provider (e.g., GitLab, Azure AD) or a different AI backend (e.g., OpenAI, Anthropic) can do so by following a documented guide. The pattern requires zero changes to core CLI plumbing — only adding a new adapter file and registering it in config.

**Why this priority**: This is the architectural value of being a template. It is delivered last because US1 and US2 define the patterns; US3 documents and validates them.

**Independent Test**: Follow `docs/customization.md` to add a mock provider adapter. Confirm the CLI picks it up via `simple-cli auth login --provider mock` and `simple-cli chat` routes to the mock backend.

**Acceptance Scenarios**:

1. **Given** the `docs/customization.md` guide exists, **When** a developer follows the "Add a new sub-command" section, **Then** a working sub-command is registered and callable in under 30 minutes.
2. **Given** the provider adapter interface is documented, **When** a developer implements it for a new OAuth provider, **Then** `simple-cli auth login --provider <new>` uses the new provider without modifying any existing file.
3. **Given** the chat backend interface is documented, **When** a developer implements it for a new AI backend, **Then** `simple-cli chat` routes to the new backend when configured, without modifying existing files.
4. **Given** the `example` sub-command exists as a reference, **When** a developer reads it alongside `docs/customization.md`, **Then** the extension pattern is self-evident and consistent with the constitution.

---

### Edge Cases

- What happens when the OAuth provider's device endpoint is unreachable? → CLI exits with a clear "provider unreachable" error and code 1; no partial state is saved.
- What happens when `auth login` is run while already authenticated? → CLI prompts "already authenticated as X; re-authenticate? [y/N]" in human mode or exits with an informative JSON error in JSON mode.
- What happens if the token file is corrupt or tampered with? → CLI detects parse failure, deletes the corrupt file, and instructs the user to run `auth login` again.
- What happens when `chat` is run without authentication? → CLI exits with code 2 (invalid argument / precondition) and prints a clear message directing the user to `auth login`.
- What happens with very long AI responses in streaming mode? → Content is flushed to stdout incrementally; the user can pipe to `less` or redirect to a file without buffering issues.

---

## Requirements

### Functional Requirements

- **FR-001**: CLI MUST provide an `auth` sub-command group with three sub-commands: `login`, `logout`, `status`.
- **FR-002**: `auth login` MUST implement the OAuth 2.0 Device Authorization Flow (RFC 8628), polling the token endpoint at the provider-specified interval until authorized or timed out.
- **FR-003**: `auth login` MUST support a `--provider <name>` flag; if omitted, it uses the default provider from config.
- **FR-004**: Obtained access tokens and refresh tokens MUST be stored in a file inside `ConfigDir()` with permissions `0600` (owner-read/write only) on Unix; equivalent ACL restrictions on Windows.
- **FR-005**: CLI MUST automatically refresh the access token using the stored refresh token when a command receives an authentication error (HTTP 401) from the API, and retry the original request once.
- **FR-006**: `auth logout` MUST delete all locally stored tokens for the active provider and confirm deletion to the user.
- **FR-007**: `auth status` MUST display: provider name, authenticated user identifier (if returned by the provider), token expiry time, and whether the token is currently valid.
- **FR-008**: CLI MUST provide a `chat` sub-command that accepts a message as a positional argument: `simple-cli chat "message"`.
- **FR-009**: `chat` MUST call the configured API endpoint, passing the Bearer access token in the `Authorization` header.
- **FR-010**: In human output mode, `chat` MUST stream response content to stdout using Server-Sent Events (SSE); each `data:` event carrying a `choices[0].delta.content` field MUST be flushed to stdout immediately. Streaming terminates on receipt of `data: [DONE]`. In JSON output mode, `chat` MUST buffer all SSE chunks, concatenate the delta content, and write a single JSON envelope after `data: [DONE]` is received.
- **FR-011**: `chat` MUST support an optional `--conversation <id>` flag. When omitted, the CLI sends no conversation identifier and treats the request as stateless. When provided, the value is forwarded to the API in the request body to enable server-side multi-turn context. The CLI MUST NOT auto-generate or persist conversation IDs to disk under any circumstances.
- **FR-012**: Provider configuration (endpoint URL, client ID, scopes, token endpoint, device endpoint) MUST be definable in the config file under a `providers:` key and overridable via environment variables.
- **FR-013**: The codebase MUST contain a documented provider adapter interface (`internal/provider/`) that any new OAuth provider or AI backend implements to integrate with the CLI.
- **FR-014**: `docs/customization.md` MUST exist and cover: (a) adding a new sub-command, (b) implementing a new OAuth provider adapter, (c) implementing a new AI chat backend adapter, (d) environment variables and config file reference for providers.
- **FR-015**: All new packages (`internal/auth/`, `internal/provider/`, `internal/chat/`) MUST be independently importable and testable with no global state, per Constitution Principle I.
- **FR-016**: CLI MUST emit structured log output to stderr (via `log/slog`) for all OAuth flow steps and chat API calls at `debug` level, allowing developers to trace requests without exposing tokens.
- **FR-017**: The OAuth target MUST be treated as a fully configurable, user-controlled OAuth 2.0 server.
- **FR-018**: The device authorization flow MUST respect the `expires_in` field returned by the provider's device authorization endpoint. If absent, the CLI MUST apply a default maximum wait of 300 seconds. The polling interval MUST use the provider's `interval` field (default: 5 seconds if absent). On expiry, the CLI MUST exit with code 1 and a message instructing the user to retry `auth login`. The CLI MUST NOT hardcode any GitHub-specific OAuth endpoints, client IDs, or Copilot API contracts. All provider endpoints (device authorization URL, token URL, chat completion URL) are supplied exclusively through the config file or environment variables. GitHub Copilot is an internal implementation detail of the user's API service and is invisible to the CLI.

### Key Entities

- **Provider**: Named OAuth configuration (name, device_endpoint, token_endpoint, client_id, scopes). Multiple providers can be registered; one is active at a time.
- **TokenStore**: Locally persisted credential set for a provider (access_token, refresh_token, expiry, provider_name). Keyed by provider name.
- **ChatRequest**: An OpenAI-compatible payload sent to the chat completion endpoint. Required fields: `model` (string, from config default or `--model` flag), `messages` (array of `{role, content}` objects; the current user turn is always `{role: "user", content: "<message>"}`), `stream: true`. Optional fields: `conversation_id` (string, forwarded only when `--conversation` is supplied by the user).
- **ChatResponse**: The assembled content from an OpenAI-compatible SSE stream. Fields consumed by the CLI: `choices[0].delta.content` (streamed text chunks) and `choices[0].finish_reason`. The JSON output envelope exposes `data.content` (full concatenated text) and `data.model`.
- **ProviderAdapter**: Interface that abstracts the differences between OAuth providers (device flow initiation, token exchange, token refresh).
- **ChatBackend**: Interface that abstracts AI completion requests (send message, stream response, handle errors).

---

## Success Criteria

### Measurable Outcomes

- **SC-001**: With the guide in `docs/customization.md`, a developer who has never seen the codebase can add a new sub-command (not auth or chat) in under 30 minutes.
- **SC-002**: The OAuth device flow from `auth login` invocation to token stored on disk completes within 2 minutes of the user completing browser authentication, under normal network conditions.
- **SC-003**: `simple-cli chat "hello"` produces the first streamed character of the AI response within 3 seconds of the HTTP response headers being received from the API.
- **SC-004**: Stored tokens survive process restarts; a fresh terminal session can run `simple-cli chat` without re-authenticating as long as the token has not expired.
- **SC-005**: All new packages in `internal/auth/`, `internal/provider/`, and `internal/chat/` achieve ≥80% unit test coverage (constitution minimum), verified by `make test`.
- **SC-006**: Adding a new OAuth provider requires code changes in exactly 1 file (the new adapter) plus a config entry — no changes to `cmd/`, `internal/auth/`, or `internal/chat/`.
- **SC-007**: The JSON envelope produced by `simple-cli --output json chat "message"` is stable and parseable; `jq '.data.content'` extracts the AI response text reliably.
- **SC-008**: `go vet ./...` and `golangci-lint run ./...` report zero issues after the feature is merged.

---

## Assumptions

- The user's API service accepts HTTP requests with a Bearer token in the `Authorization` header and returns JSON responses. The exact endpoint URL, client ID, device endpoint URL, and token endpoint URL are supplied via config — they are never hardcoded. No provider (including GitHub) receives special treatment in the codebase.
- Streaming uses Server-Sent Events (SSE) with the OpenAI `data: {...}` / `data: [DONE]` protocol exclusively. The chat package reads the HTTP response body line-by-line, parses each `data:` payload as JSON, and extracts `choices[0].delta.content`. Chunked Transfer-Encoding without SSE framing is out of scope for this feature.
- Token storage uses a plain JSON file at `ConfigDir()/tokens.json` with `0600` permissions. OS keychain integration is out of scope for this feature but the `TokenStore` interface is designed to allow a future keychain backend.
- The `--conversation` flag (FR-011) is fully opt-in. Each `chat` invocation is stateless by default. No conversation IDs are generated, stored, or inferred by the CLI. Server-side conversation continuity is entirely the responsibility of the user's API service.
- The `example` sub-command from feature 002 remains in place as a template reference; `docs/customization.md` references it.
