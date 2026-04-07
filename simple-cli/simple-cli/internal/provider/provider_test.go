package provider

import (
	"context"
	"testing"
	"time"
)

func TestTokenSet_IsExpired(t *testing.T) {
	past := &TokenSet{Expiry: time.Now().Add(-time.Minute)}
	if !past.IsExpired() {
		t.Fatalf("expected past token to be expired")
	}

	future := &TokenSet{Expiry: time.Now().Add(10 * time.Minute)}
	if future.IsExpired() {
		t.Fatalf("expected future token to NOT be expired")
	}
}

// package-level mocks to validate interface shapes at compile time.
type mockAdapter struct{}

func (m *mockAdapter) StartDeviceFlow(ctx context.Context) (*DeviceFlowState, error) {
	return &DeviceFlowState{DeviceCode: "d", UserCode: "u", VerificationURI: "https://example.com", ExpiresAt: time.Now().Add(time.Minute)}, nil
}
func (m *mockAdapter) PollToken(ctx context.Context, state *DeviceFlowState) (*TokenSet, error) {
	return &TokenSet{Provider: "mock", AccessToken: "a", Expiry: time.Now().Add(time.Hour)}, nil
}
func (m *mockAdapter) RefreshToken(ctx context.Context, refreshToken string) (*TokenSet, error) {
	return &TokenSet{Provider: "mock", AccessToken: "b", Expiry: time.Now().Add(time.Hour)}, nil
}

type mockStore struct{}

func (m *mockStore) Get(ctx context.Context, provider string) (*TokenSet, error) { return nil, nil }
func (m *mockStore) Set(ctx context.Context, provider string, t *TokenSet) error { return nil }
func (m *mockStore) Delete(ctx context.Context, provider string) error           { return nil }

type mockChat struct{}

func (m *mockChat) Chat(ctx context.Context, req *ChatRequest, token string) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 1)
	ch <- StreamEvent{Delta: "hi", Done: true}
	close(ch)
	return ch, nil
}

func TestInterfaceCompliance(t *testing.T) {
	var _ ProviderAdapter = (*mockAdapter)(nil)
	var _ TokenStore = (*mockStore)(nil)
	var _ ChatBackend = (*mockChat)(nil)
}
