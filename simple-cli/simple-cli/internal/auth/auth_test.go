package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binpqh/simple-cli/internal/config"
)

func TestStartDeviceFlowAndPollSuccess(t *testing.T) {
	var tokenCalls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/device":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"device_code":               "dev-123",
				"user_code":                 "USER-ABC",
				"verification_uri":          "https://example.com/activate",
				"verification_uri_complete": "https://example.com/activate?user=USER-ABC",
				"expires_in":                300,
				"interval":                  1,
			})
		case "/token":
			n := atomic.AddInt32(&tokenCalls, 1)
			if n < 3 {
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "authorization_pending"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "AT-XYZ",
				"expires_in":   3600,
				"token_type":   "Bearer",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	cfg := &config.ProviderConfig{
		ClientID:       "cid",
		DeviceEndpoint: ts.URL + "/device",
		TokenEndpoint:  ts.URL + "/token",
		ChatEndpoint:   ts.URL + "/chat",
	}
	adapter := NewHTTPProviderAdapter(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	state, err := adapter.StartDeviceFlow(ctx)
	if err != nil {
		t.Fatalf("StartDeviceFlow failed: %v", err)
	}
	if state.UserCode == "" || state.DeviceCode == "" {
		t.Fatalf("invalid device state: %+v", state)
	}

	tset, err := adapter.PollToken(ctx, state)
	if err != nil {
		t.Fatalf("PollToken failed: %v", err)
	}
	if tset == nil || tset.AccessToken == "" {
		t.Fatalf("expected token, got %+v", tset)
	}
}

func TestPollToken_AccessDenied(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/device":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"device_code":      "dev-1",
				"user_code":        "USER-1",
				"verification_uri": "https://example.com/activate",
				"expires_in":       300,
				"interval":         1,
			})
		case "/token":
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "access_denied"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	cfg := &config.ProviderConfig{
		ClientID:       "cid",
		DeviceEndpoint: ts.URL + "/device",
		TokenEndpoint:  ts.URL + "/token",
		ChatEndpoint:   ts.URL + "/chat",
	}
	adapter := NewHTTPProviderAdapter(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state, err := adapter.StartDeviceFlow(ctx)
	if err != nil {
		t.Fatalf("StartDeviceFlow failed: %v", err)
	}
	_, err = adapter.PollToken(ctx, state)
	if err == nil {
		t.Fatalf("expected ErrAuthDenied, got nil")
	}
}

func TestPollToken_ExpiredDevice(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/device":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"device_code":      "dev-exp",
				"user_code":        "USR-EXP",
				"verification_uri": "https://example.com/activate",
				"expires_in":       300,
				"interval":         1,
			})
		case "/token":
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "expired_token"})
		}
	}))
	defer ts.Close()

	cfg := &config.ProviderConfig{ClientID: "cid", DeviceEndpoint: ts.URL + "/device", TokenEndpoint: ts.URL + "/token", ChatEndpoint: ts.URL + "/chat"}
	adapter := NewHTTPProviderAdapter(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state, _ := adapter.StartDeviceFlow(ctx)
	_, err := adapter.PollToken(ctx, state)
	if err == nil {
		t.Fatalf("expected ErrDeviceExpired, got nil")
	}
}

func TestRefreshToken_HappyPath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "new-at",
			"refresh_token": "new-rt",
			"expires_in":    3600,
			"token_type":    "Bearer",
		})
	}))
	defer ts.Close()

	cfg := &config.ProviderConfig{ClientID: "cid", DeviceEndpoint: ts.URL, TokenEndpoint: ts.URL, ChatEndpoint: ts.URL}
	adapter := NewHTTPProviderAdapter(cfg)

	tset, err := adapter.RefreshToken(context.Background(), "old-rt")
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if tset == nil || tset.AccessToken != "new-at" {
		t.Fatalf("expected new access token, got %+v", tset)
	}
	if tset.RefreshToken != "new-rt" {
		t.Fatalf("expected new refresh token, got %q", tset.RefreshToken)
	}
}

func TestRefreshToken_EmptyRefreshToken(t *testing.T) {
	cfg := &config.ProviderConfig{ClientID: "cid", DeviceEndpoint: "http://x", TokenEndpoint: "http://x", ChatEndpoint: "http://x"}
	adapter := NewHTTPProviderAdapter(cfg)

	_, err := adapter.RefreshToken(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error for empty refresh token")
	}
}

func TestRefreshToken_NoAccessTokenInResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"expires_in": 3600})
	}))
	defer ts.Close()

	cfg := &config.ProviderConfig{ClientID: "cid", DeviceEndpoint: ts.URL, TokenEndpoint: ts.URL, ChatEndpoint: ts.URL}
	adapter := NewHTTPProviderAdapter(cfg)

	_, err := adapter.RefreshToken(context.Background(), "some-rt")
	if err == nil {
		t.Fatalf("expected error when access_token absent from response")
	}
}

func TestStartDeviceFlow_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	cfg := &config.ProviderConfig{ClientID: "cid", DeviceEndpoint: ts.URL + "/device", TokenEndpoint: ts.URL + "/token", ChatEndpoint: ts.URL}
	adapter := NewHTTPProviderAdapter(cfg)

	_, err := adapter.StartDeviceFlow(context.Background())
	if err == nil {
		t.Fatalf("expected error from 500 endpoint")
	}
}

func TestPollToken_ContextCancellation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/device":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"device_code": "d", "user_code": "u",
				"verification_uri": "https://example.com", "expires_in": 300, "interval": 1,
			})
		case "/token":
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "authorization_pending"})
		}
	}))
	defer ts.Close()

	cfg := &config.ProviderConfig{ClientID: "cid", DeviceEndpoint: ts.URL + "/device", TokenEndpoint: ts.URL + "/token", ChatEndpoint: ts.URL}
	adapter := NewHTTPProviderAdapter(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	state, _ := adapter.StartDeviceFlow(context.Background())
	_, err := adapter.PollToken(ctx, state)
	if err == nil {
		t.Fatalf("expected context cancellation error, got nil")
	}
}
