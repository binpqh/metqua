# Contract: Go Interfaces

**Branch**: `003-oauth-ai-chat-provider`
**Package**: `internal/provider`
**Purpose**: Define the three extension interfaces that decouple commands from implementation.
            These are the primary extension points described in US3 (Customization).

---

## Overview

The three interfaces form a layered stack:

```
cmd/auth/login.go
    └── ProviderAdapter  (internal/provider)
            └── TokenStore  (internal/provider)

cmd/chat.go
    └── ChatBackend  (internal/provider)
    └── TokenStore   (internal/provider)
```

Default implementations:

| Interface | Default Implementation | Package |
|-----------|----------------------|---------|
| `ProviderAdapter` | `HTTPProviderAdapter` | `internal/auth` |
| `TokenStore` | `FileTokenStore` | `internal/tokenstore` |
| `ChatBackend` | `SSEChatBackend` | `internal/chat` |

---

## `ProviderAdapter`

```go
package provider

import "context"

// ProviderAdapter abstracts the OAuth 2.0 Device Authorization Flow
// (RFC 8628) for a single provider.
//
// Implement this interface to support a custom auth server or alternate
// OAuth flow (e.g. Authorization Code + PKCE via browser redirect).
type ProviderAdapter interface {
    // StartDeviceFlow initiates the device authorization request.
    // It returns a DeviceFlowState containing the user code and polling parameters.
    // The caller is responsible for displaying VerificationURI and UserCode.
    StartDeviceFlow(ctx context.Context) (*DeviceFlowState, error)

    // PollToken polls the token endpoint until authorization succeeds,
    // the device code expires (DeviceFlowState.ExpiresAt), or ctx is cancelled.
    //
    // Polling interval is DeviceFlowState.Interval.
    // Returns ErrDeviceExpired if ExpiresAt is reached before approval.
    // Returns ErrAuthDenied if the user explicitly denies the request.
    PollToken(ctx context.Context, state *DeviceFlowState) (*TokenSet, error)

    // RefreshToken exchanges an existing refresh token for a new TokenSet.
    // Returns ErrNoRefreshToken if the provider does not issue refresh tokens.
    RefreshToken(ctx context.Context, refreshToken string) (*TokenSet, error)
}
```

---

## `TokenStore`

```go
package provider

import "context"

// TokenStore persists and retrieves TokenSets keyed by provider name.
//
// The default implementation (FileTokenStore) writes tokens.json with
// mode 0600. Implement this interface to use OS keychain, HashiCorp Vault,
// or any other secure storage backend.
type TokenStore interface {
    // Get returns the stored TokenSet for the given provider.
    // Returns ErrTokenNotFound if no token exists.
    Get(ctx context.Context, provider string) (*TokenSet, error)

    // Set stores or replaces the TokenSet for the given provider.
    Set(ctx context.Context, provider string, t *TokenSet) error

    // Delete removes the stored TokenSet for the given provider.
    // It is idempotent: returns nil if no token exists.
    Delete(ctx context.Context, provider string) error
}

// ErrTokenNotFound is returned by TokenStore.Get when no token is stored.
var ErrTokenNotFound = errors.New("no token stored for provider")
```

---

## `ChatBackend`

```go
package provider

import "context"

// ChatBackend streams chat completion responses from an OpenAI-compatible endpoint.
//
// Implement this interface to support a non-SSE transport (e.g. WebSocket,
// gRPC streaming) or to add request/response middleware.
type ChatBackend interface {
    // Chat sends req to the provider's chat completions endpoint,
    // authenticated with the given Bearer token.
    //
    // It returns a channel of StreamEvents. The channel is closed after
    // either a StreamEvent with Done=true or a StreamEvent with Err!=nil.
    // Callers must drain the channel after the first Err or Done event.
    Chat(ctx context.Context, req *ChatRequest, token string) (<-chan StreamEvent, error)
}
```

---

## Supporting Types

```go
package provider

import "time"

// DeviceFlowState holds in-flight device authorization data.
// It is never persisted to disk.
type DeviceFlowState struct {
    DeviceCode      string
    UserCode        string
    VerificationURI string
    ExpiresAt       time.Time
    Interval        time.Duration
}

// TokenSet holds the credential material for one provider.
// WARNING: Never log fields; Stringer is intentionally omitted.
type TokenSet struct {
    Provider     string    `json:"provider"`
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token,omitempty"`
    Expiry       time.Time `json:"expiry"`
    UserID       string    `json:"user_id,omitempty"`
    TokenType    string    `json:"token_type,omitempty"`
}

// IsExpired reports whether the access token has expired (with 30 s buffer).
func (t *TokenSet) IsExpired() bool {
    return time.Now().Add(30 * time.Second).After(t.Expiry)
}

// ChatRequest is the payload sent to the chat completions endpoint.
type ChatRequest struct {
    Model          string        `json:"model"`
    Messages       []ChatMessage `json:"messages"`
    Stream         bool          `json:"stream"`
    ConversationID string        `json:"conversation_id,omitempty"`
}

// ChatMessage is one message in the conversation history.
type ChatMessage struct {
    Role    string `json:"role"`    // "user" | "assistant" | "system"
    Content string `json:"content"`
}

// StreamEvent is one item emitted by ChatBackend.Chat.
type StreamEvent struct {
    Delta string // non-empty when content is available
    Done  bool   // true when the stream is complete (data: [DONE])
    Err   error  // non-nil on fatal stream error
}
```

---

## Sentinel Errors

```go
package provider

import "errors"

var (
    // ErrDeviceExpired is returned when the device code expires before the user approves.
    ErrDeviceExpired = errors.New("device code expired; run 'auth login' again")

    // ErrAuthDenied is returned when the user explicitly denies the device flow request.
    ErrAuthDenied = errors.New("authorization denied")

    // ErrNoRefreshToken is returned when RefreshToken is called but the provider
    // does not issue refresh tokens.
    ErrNoRefreshToken = errors.New("provider does not issue refresh tokens")

    // ErrProviderNotFound is returned when the requested provider is not in config.
    ErrProviderNotFound = errors.New("provider not configured; check config file")

    // ErrTokenNotFound is returned by TokenStore.Get when no token is stored.
    ErrTokenNotFound = errors.New("no token stored for provider")
)
```

---

## Extension Guide: Custom ProviderAdapter

To replace the default HTTP Device Flow adapter with a custom implementation:

```go
// File: internal/auth/myprovider/myprovider.go
package myprovider

import (
    "context"
    "github.com/binpqh/simple-cli/internal/provider"
)

type MyProviderAdapter struct {
    clientID string
    // ... your fields
}

var _ provider.ProviderAdapter = (*MyProviderAdapter)(nil) // compile-time check

func (a *MyProviderAdapter) StartDeviceFlow(ctx context.Context) (*provider.DeviceFlowState, error) {
    // your implementation
}

func (a *MyProviderAdapter) PollToken(ctx context.Context, state *provider.DeviceFlowState) (*provider.TokenSet, error) {
    // your implementation
}

func (a *MyProviderAdapter) RefreshToken(ctx context.Context, refreshToken string) (*provider.TokenSet, error) {
    // your implementation
}
```

Then register it in `cmd/auth/login.go`'s provider registry map:

```go
providerRegistry := map[string]func(*config.ProviderConfig) provider.ProviderAdapter{
    "default":    auth.NewHTTPProviderAdapter,
    "myprovider": myprovider.New,
}
```

See [docs/customization.md](../../../docs/customization.md) for a full walkthrough.
