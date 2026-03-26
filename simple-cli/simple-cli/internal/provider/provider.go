package provider

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors used across provider implementations.
var (
	ErrDeviceExpired    = errors.New("device code expired; run auth login again")
	ErrAuthDenied       = errors.New("authorization denied by user")
	ErrNoRefreshToken   = errors.New("provider does not issue refresh tokens")
	ErrProviderNotFound = errors.New("provider not configured; check config file")
	ErrTokenNotFound    = errors.New("no token stored for provider")
	ErrUnauthorized     = errors.New("request unauthorized: token expired or invalid")
)

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

// ChatMessage is one message in the conversation history.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the payload sent to the chat completions endpoint.
type ChatRequest struct {
	Model          string        `json:"model"`
	Messages       []ChatMessage `json:"messages"`
	Stream         bool          `json:"stream"`
	ConversationID string        `json:"conversation_id,omitempty"`
}

// StreamEvent is one item emitted by ChatBackend.Chat.
type StreamEvent struct {
	Delta string
	Done  bool
	Err   error
}

// ProviderAdapter abstracts the OAuth 2.0 Device Authorization Flow for a single provider.
type ProviderAdapter interface {
	StartDeviceFlow(ctx context.Context) (*DeviceFlowState, error)
	PollToken(ctx context.Context, state *DeviceFlowState) (*TokenSet, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenSet, error)
}

// TokenStore persists and retrieves TokenSets keyed by provider name.
type TokenStore interface {
	Get(ctx context.Context, provider string) (*TokenSet, error)
	Set(ctx context.Context, provider string, t *TokenSet) error
	Delete(ctx context.Context, provider string) error
}

// ChatBackend streams chat completion responses from an OpenAI-compatible endpoint.
type ChatBackend interface {
	Chat(ctx context.Context, req *ChatRequest, token string) (<-chan StreamEvent, error)
}
