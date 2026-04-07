# Data Model: OAuth Provider + AI Chat Completion

**Branch**: `003-oauth-ai-chat-provider` | **Spec**: FR-001–FR-018, §SC
**Input**: research.md §§1–5

---

## Entity Relationship Overview

```
Config
└── Providers map[string]ProviderConfig   (1 config : N providers)
└── DefaultProvider string

ProviderConfig                             (describes 1 OAuth + Chat server)
    ClientID, DeviceEndpoint, TokenEndpoint, ChatEndpoint, Scopes, DefaultModel

TokenSet                                   (stored per provider in tokens.json)
    Provider, AccessToken, RefreshToken, Expiry, UserID

DeviceFlowState                            (ephemeral, in-memory during auth login)
    DeviceCode, UserCode, VerificationURI, ExpiresIn, Interval

ChatRequest ──contains──> []ChatMessage    (sent to ChatEndpoint)
    Model, Messages, Stream, ConversationID?

SSEChunk                                   (received per SSE event during streaming)
    ID, Choices[0].Delta.Content, FinishReason
```

---

## 1. Config Layer

### `ProviderConfig` — Go struct (new)
Added to `internal/config/config.go`.

```go
// ProviderConfig holds all settings for one OAuth 2.0 + Chat provider.
type ProviderConfig struct {
    ClientID        string   `mapstructure:"client_id"`
    DeviceEndpoint  string   `mapstructure:"device_endpoint"`
    TokenEndpoint   string   `mapstructure:"token_endpoint"`
    ChatEndpoint    string   `mapstructure:"chat_endpoint"`
    Scopes          []string `mapstructure:"scopes"`
    DefaultModel    string   `mapstructure:"default_model"`
}
```

### `Config` struct additions (patch to existing)
```go
// Added fields in Config:
DefaultProvider string                       `mapstructure:"default_provider"`
Providers       map[string]ProviderConfig    `mapstructure:"providers"`
```

### Validation Rules
- `ClientID` MUST be non-empty when `auth login` is invoked
- `DeviceEndpoint` MUST be a valid HTTPS URL
- `TokenEndpoint` MUST be a valid HTTPS URL
- `ChatEndpoint` MUST be a valid HTTPS URL
- `Scopes` MAY be empty (some providers do not require explicit scopes)
- `DefaultModel` MAY be empty; `--model` flag overrides it at runtime
- If `DefaultProvider` is set, a matching key MUST exist in `Providers`

---

## 2. Token Storage

### `TokenSet` — Go struct (new, `internal/tokenstore/types.go`)
```go
// TokenSet holds the token material for one provider.
// WARNING: Never log fields; Stringer intentionally omitted.
type TokenSet struct {
    Provider     string    `json:"provider"`
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token,omitempty"`
    Expiry       time.Time `json:"expiry"`
    UserID       string    `json:"user_id,omitempty"`
    TokenType    string    `json:"token_type,omitempty"`
}

// IsExpired returns true when the token is within the 30-second buffer window.
func (t *TokenSet) IsExpired() bool {
    return time.Now().Add(30 * time.Second).After(t.Expiry)
}
```

### `tokenFile` — on-disk layout
```json
{
  "providers": {
    "my-api": {
      "provider": "my-api",
      "access_token": "eyJhbGci...",
      "refresh_token": "def50200...",
      "expiry": "2026-03-25T10:00:00Z",
      "token_type": "Bearer",
      "user_id": "user-123"
    }
  }
}
```

File mode: `0600` (owner-read/write only). Path: `ConfigDir()/tokens.json`.

### `TokenStore` interface (`internal/provider/provider.go`)
```go
type TokenStore interface {
    Get(ctx context.Context, provider string) (*TokenSet, error)
    Set(ctx context.Context, provider string, t *TokenSet) error
    Delete(ctx context.Context, provider string) error
}
```

### State Transitions: TokenSet lifecycle
```
[absent] ──(auth login success)──> [valid]
[valid]  ──(auth logout)──────────> [absent]
[valid]  ──(expiry + refresh)─────> [valid]  (access_token rotated)
[valid]  ──(expiry, no refresh)───> [expired]
[expired]──(auth login)───────────> [valid]
[expired]──(chat attempt)─────────> ERROR: "token expired, run auth login"
```

---

## 3. Auth Layer

### `DeviceFlowState` — ephemeral (scoped to `auth login` command)
```go
// DeviceFlowState holds the device code and display values during polling.
// It is never persisted to disk.
type DeviceFlowState struct {
    DeviceCode      string
    UserCode        string
    VerificationURI string
    ExpiresAt       time.Time   // computed from expires_in
    Interval        time.Duration
}
```

### `DeviceAuthResponse` — JSON deserialization target
```go
type DeviceAuthResponse struct {
    DeviceCode              string `json:"device_code"`
    UserCode                string `json:"user_code"`
    VerificationURI         string `json:"verification_uri"`
    VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
    ExpiresIn               int    `json:"expires_in"`
    Interval                int    `json:"interval"`
}
```

### `TokenResponse` — JSON deserialization target
```go
type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token,omitempty"`
    ExpiresIn    int    `json:"expires_in"`
    TokenType    string `json:"token_type"`
    Scope        string `json:"scope,omitempty"`
    UserID       string `json:"user_id,omitempty"`
    // Error fields (RFC 8628 §3.5)
    Error            string `json:"error,omitempty"`
    ErrorDescription string `json:"error_description,omitempty"`
}
```

### `ProviderAdapter` interface (`internal/provider/provider.go`)
```go
type ProviderAdapter interface {
    // StartDeviceFlow initiates RFC 8628 device authorization and returns
    // the device code state required for polling.
    StartDeviceFlow(ctx context.Context) (*DeviceFlowState, error)

    // PollToken polls the token endpoint until either the user approves the
    // request, the device code expires, or ctx is cancelled.
    PollToken(ctx context.Context, state *DeviceFlowState) (*TokenSet, error)

    // RefreshToken exchanges the given refresh token for a new TokenSet.
    // Returns ErrNoRefreshToken if the provider does not support refresh.
    RefreshToken(ctx context.Context, refreshToken string) (*TokenSet, error)
}
```

### Sentinel Errors (`internal/auth/auth.go`)
```go
var (
    ErrDeviceExpired    = errors.New("device code expired; run auth login again")
    ErrAuthDenied       = errors.New("authorization denied by user")
    ErrNoRefreshToken   = errors.New("provider does not issue refresh tokens")
    ErrProviderNotFound = errors.New("provider not configured; check config file")
)
```

---

## 4. Chat Layer

### `ChatMessage` (component of ChatRequest)
```go
type ChatMessage struct {
    Role    string `json:"role"`    // "user" | "assistant" | "system"
    Content string `json:"content"`
}
```

### `ChatRequest` — sent to `POST {chat_endpoint}`
```go
type ChatRequest struct {
    Model          string        `json:"model"`
    Messages       []ChatMessage `json:"messages"`
    Stream         bool          `json:"stream"`
    ConversationID string        `json:"conversation_id,omitempty"` // opt-in via --conversation
}
```

### `SSEChunk` — one parsed SSE `data:` line
```go
type SSEDelta struct {
    Content string `json:"content"`
    Role    string `json:"role,omitempty"`
}

type SSEChoice struct {
    Delta        SSEDelta `json:"delta"`
    FinishReason string   `json:"finish_reason,omitempty"`
    Index        int      `json:"index"`
}

type SSEChunk struct {
    ID      string      `json:"id"`
    Object  string      `json:"object"`
    Choices []SSEChoice `json:"choices"`
}
```

### `ChatBackend` interface (`internal/provider/provider.go`)
```go
// ChatBackend streams chat completion deltas from an OpenAI-compatible endpoint.
type ChatBackend interface {
    // Chat sends req to the provider's chat endpoint, authenticated with token.
    // It returns a channel that yields content deltas until closed.
    // An error value in the channel of type error signals a fatal stream error.
    Chat(ctx context.Context, req *ChatRequest, token string) (<-chan StreamEvent, error)
}

// StreamEvent is one item emitted by ChatBackend.Chat.
type StreamEvent struct {
    Delta string // non-empty when content is available
    Done  bool   // true on [DONE] signal
    Err   error  // non-nil on fatal error
}
```

---

## 5. Output Envelopes

The existing `internal/output` package provides `Human` and `JSON` output modes. New commands emit JSON-compatible structs that map to the contract defined in `contracts/`.

### `AuthStatusOutput`
```go
type AuthStatusOutput struct {
    Provider  string `json:"provider"`
    LoggedIn  bool   `json:"logged_in"`
    UserID    string `json:"user_id,omitempty"`
    ExpiresAt string `json:"expires_at,omitempty"` // RFC 3339
    Expired   bool   `json:"expired"`
}
```

### `AuthLoginOutput`
```go
type AuthLoginOutput struct {
    Provider string `json:"provider"`
    Success  bool   `json:"success"`
    UserID   string `json:"user_id,omitempty"`
    Message  string `json:"message"`
}
```

### `AuthLogoutOutput`
```go
type AuthLogoutOutput struct {
    Provider string `json:"provider"`
    Success  bool   `json:"success"`
    Message  string `json:"message"`
}
```

### `ChatOutput` (json mode, non-streaming)
```go
type ChatOutput struct {
    Provider  string `json:"provider"`
    Model     string `json:"model"`
    Content   string `json:"content"`    // full assembled response
    Streaming bool   `json:"streaming"`  // always true (SSE used internally)
}
```

---

## 6. Validation Rules Summary

| Field | Rule |
|-------|------|
| `ProviderConfig.ClientID` | non-empty, no URL encoding |
| `ProviderConfig.DeviceEndpoint` | valid HTTPS URL |
| `ProviderConfig.TokenEndpoint` | valid HTTPS URL |
| `ProviderConfig.ChatEndpoint` | valid HTTPS URL |
| `TokenSet.AccessToken` | non-empty after successful auth |
| `TokenSet.Expiry` | UTC; checked via `IsExpired()` before every chat request |
| `ChatRequest.Messages` | at least one message; last message role MUST be "user" |
| `ChatRequest.Model` | non-empty; falls back to `ProviderConfig.DefaultModel` |
| `DeviceFlowState.ExpiresAt` | set to `time.Now().Add(time.Duration(ExpiresIn) * time.Second)`; 300 s if `ExpiresIn == 0` |
