# Research: OAuth Provider + AI Chat Completion via CLI

**Branch**: `003-oauth-ai-chat-provider` | **Date**: 2026-03-24
**Purpose**: Resolve all NEEDS CLARIFICATION items and technical unknowns identified in plan.md Technical Context before Phase 1 design begins.

---

## §1 — OAuth 2.0 Device Flow: Roll Our Own vs `golang.org/x/oauth2`

### Decision

**Roll our own** using stdlib `net/http` + `encoding/json`. No new go module dependency.

### Rationale

The OAuth 2.0 Device Authorization Flow (RFC 8628) is two HTTP calls plus a polling loop:

1. `POST {device_endpoint}` → `{device_code, user_code, verification_uri, interval, expires_in}`
2. Display `user_code` + `verification_uri` to user
3. Loop: `POST {token_endpoint}` every `interval` seconds → poll until `authorization_pending` clears or `expires_in` elapses → returns `{access_token, refresh_token, token_type, expires_in}`

This is ~80 lines of `net/http` code. Constitution Principle VIII requires a direct dependency only when it "saves >100 lines of well-tested code or covers a security-sensitive surface." The device flow is well-understood, easily tested with `httptest.NewServer`, and needs precise `expires_in`/`interval` control that `golang.org/x/oauth2/deviceauth` abstracts away (making it harder to honour FR-018: respect provider `expires_in`).

### Alternatives Considered

| Option                                          | Verdict     | Reason Rejected                                                                                                                                  |
| ----------------------------------------------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `golang.org/x/oauth2/deviceauth`                | ❌ Rejected | Adds transitive deps (`golang.org/x/net`, `golang.org/x/text`); abstracts away `expires_in`/`interval` control needed by FR-018; saves <80 lines |
| `github.com/cli/oauth` (GitHub's CLI OAuth lib) | ❌ Rejected | GitHub-specific assumptions; FR-017 forbids GitHub coupling                                                                                      |
| Stdlib `net/http`                               | ✅ Chosen   | Zero new deps; full control of polling; easily mocked with httptest; satisfies FR-018                                                            |

---

## §2 — SSE Streaming in Go (stdlib only)

### Decision

Use `bufio.Scanner` with a custom `SplitFunc` to split on `\n`, reading each line prefixed with `data: ` as an SSE event. Flush to stdout after each non-empty delta content.

### Pattern

```go
// Pseudo-code — full implementation in internal/chat/chat.go
scanner := bufio.NewScanner(resp.Body)
for scanner.Scan() {
    line := scanner.Text()
    if !strings.HasPrefix(line, "data: ") { continue }
    payload := strings.TrimPrefix(line, "data: ")
    if payload == "[DONE]" { break }
    // parse JSON delta
    var chunk sseChunk
    if err := json.Unmarshal([]byte(payload), &chunk); err != nil { continue }
    if delta := chunk.Choices[0].Delta.Content; delta != "" {
        fmt.Fprint(w, delta)  // flush immediately
    }
}
```

### Rationale

- No external dep (`bufio` + `encoding/json` are stdlib)
- `bufio.Scanner` default token size is 64 KB per line; sufficient for AI response chunks
- `resp.Body` is a streaming reader; content appears incrementally as the server sends it
- For large responses: Go's `http.Transport` streams by default; no full-buffer issue

### Alternatives Considered

| Option                   | Verdict     | Reason                                                                       |
| ------------------------ | ----------- | ---------------------------------------------------------------------------- |
| `github.com/r3labs/sse`  | ❌ Rejected | External dep; adds reconnect logic we don't need; overkill for one-shot chat |
| Manual byte-by-byte read | ❌ Rejected | Fragile, slower, harder to test                                              |
| `bufio.Scanner` (chosen) | ✅ Chosen   | Stdlib, simple, testable with `strings.NewReader` in tests                   |

---

## §3 — Token Storage Security: Cross-Platform 0600 Permissions

### Decision

`os.WriteFile(path, data, 0600)` on all platforms. Add a `// TODO(security): Windows DACL hardening` comment. Windows ACL via `icacls` or `golang.org/x/sys/windows` is deferred.

### Rationale

- On Unix: `0600` is enforced by the OS; owner-only read/write.
- On Windows: Go's `os.WriteFile` with `0600` creates the file with default ACL (inherits from directory), which already restricts access to the file owner on a standard NTFS volume. The Go stdlib does not translate Unix mode bits to Windows DACLs; true Windows hardening would require `SetFileSecurity` via `golang.org/x/sys/windows`.
- Constitution Principle VIII: no new deps unless security-sensitive. The token file on Windows is protected by user-profile directory ACL (only the logged-in user's processes can read `%APPDATA%`), which is sufficient for a developer tool.
- The `TokenStore` interface allows a future keychain backend that would bypass file storage entirely.

### File Layout

```
ConfigDir()/
└── tokens.json   # {"providers": {"my-provider": {"access_token":"...", "refresh_token":"...", "expiry":"..."}}}
```

### Key Security Rules

- Access tokens and refresh tokens MUST be redacted in all `slog` output (use `internal/security.Redact`)
- `tokens.json` MUST NOT be committed to git (add to `.gitignore` if not already)
- Token struct MUST NOT implement `fmt.Stringer` to prevent accidental logging
- If `tokens.json` fails JSON parse, it is deleted and the user is directed to `auth login`

---

## §4 — Provider Config Integration with Viper

### Decision

Add `ProviderConfig` struct and `Providers map[string]ProviderConfig` + `DefaultProvider string` to `internal/config/config.go`. Unmarshal via Viper's existing `mapstructure` pipeline.

### Config File Schema

```yaml
# $XDG_CONFIG_HOME/simple-cli/config.yaml
default_provider: my-api

providers:
  my-api:
    client_id: "abc123"
    device_endpoint: "https://api.example.com/oauth/device/code"
    token_endpoint: "https://api.example.com/oauth/token"
    chat_endpoint: "https://api.example.com/v1/chat/completions"
    scopes: ["chat", "read"]
    default_model: "gpt-4o"
```

### Environment Variable Overrides

| Variable                              | Overrides                           |
| ------------------------------------- | ----------------------------------- |
| `SIMPLE_CLI_DEFAULT_PROVIDER`         | `default_provider`                  |
| `SIMPLE_CLI_PROVIDER_CLIENT_ID`       | active provider's `client_id`       |
| `SIMPLE_CLI_PROVIDER_DEVICE_ENDPOINT` | active provider's `device_endpoint` |
| `SIMPLE_CLI_PROVIDER_TOKEN_ENDPOINT`  | active provider's `token_endpoint`  |
| `SIMPLE_CLI_PROVIDER_CHAT_ENDPOINT`   | active provider's `chat_endpoint`   |

Note: per-provider env overrides apply to the _active_ (selected) provider only. Multi-provider config requires the config file.

### Viper Integration Note

Viper unmarshals nested maps via `mapstructure`. The key `providers` becomes `map[string]ProviderConfig`. Because Viper keys are case-insensitive, provider names in config are normalized to lowercase.

---

## §5 — Provider Adapter Interface Pattern

### Decision

Define three interfaces in `internal/provider/provider.go`:

```go
// ProviderAdapter abstracts OAuth 2.0 Device Flow operations for a single provider.
type ProviderAdapter interface {
    StartDeviceFlow(ctx context.Context) (*DeviceFlowState, error)
    PollToken(ctx context.Context, state *DeviceFlowState) (*TokenSet, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenSet, error)
}

// TokenStore persists and retrieves token sets keyed by provider name.
type TokenStore interface {
    Get(ctx context.Context, provider string) (*TokenSet, error)
    Set(ctx context.Context, provider string, t *TokenSet) error
    Delete(ctx context.Context, provider string) error
}

// ChatBackend sends a chat completion request and streams the response.
type ChatBackend interface {
    Chat(ctx context.Context, req *ChatRequest, token string) (<-chan string, error)
    // Chan yields content deltas; closed with nil on [DONE], error value on failure.
}
```

The default implementation of `ProviderAdapter` (in `internal/auth/`) uses `net/http` and reads config from `*config.ProviderConfig`. `TokenStore` default implementation (in `internal/tokenstore/`) writes `tokens.json`. `ChatBackend` default (in `internal/chat/`) uses SSE streaming.

### Extension Pattern for Customisers

To add a new provider:

1. Create `internal/auth/myprovider/myprovider.go` implementing `ProviderAdapter`
2. Register via config: add `my-provider:` section to `config.yaml`
3. Register the constructor in `cmd/auth/login.go`'s `providerRegistry` map

Zero changes to `internal/auth/`, `internal/chat/`, or `internal/tokenstore/`.

### NEEDS CLARIFICATION Resolution

All five clarification answers are incorporated:

- Q1 → no hardcoded providers; all endpoints from config
- Q2 → ChatRequest uses OpenAI messages array + `stream: true`
- Q3 → conversation_id is opt-in, never auto-generated
- Q4 → `expires_in` respected, 300 s default, `interval` respected, 5 s default
- Q5 → SSE with `data: [DONE]`; `bufio.Scanner` line-by-line parse
