package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/provider"
)

// DeviceAuthResponse represents the RFC 8628 device authorization response.
type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// TokenResponse represents a token endpoint response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// HTTPProviderAdapter implements provider.ProviderAdapter using net/http.
type HTTPProviderAdapter struct {
	cfg    *config.ProviderConfig
	client *http.Client
}

func NewHTTPProviderAdapter(cfg *config.ProviderConfig) *HTTPProviderAdapter {
	return &HTTPProviderAdapter{cfg: cfg, client: &http.Client{Timeout: 30 * time.Second}}
}

func (h *HTTPProviderAdapter) StartDeviceFlow(ctx context.Context) (*provider.DeviceFlowState, error) {
	form := url.Values{}
	form.Set("client_id", h.cfg.ClientID)
	if len(h.cfg.Scopes) > 0 {
		form.Set("scope", strings.Join(h.cfg.Scopes, " "))
	}
	req, err := http.NewRequestWithContext(ctx, "POST", h.cfg.DeviceEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("device endpoint returned status %d", resp.StatusCode)
	}
	var dar DeviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&dar); err != nil {
		return nil, err
	}
	expires := dar.ExpiresIn
	if expires == 0 {
		expires = 300
	}
	interval := dar.Interval
	if interval == 0 {
		interval = 5
	}
	return &provider.DeviceFlowState{
		DeviceCode:      dar.DeviceCode,
		UserCode:        dar.UserCode,
		VerificationURI: verificationURICompleteOr(dar.VerificationURIComplete, dar.VerificationURI),
		ExpiresAt:       time.Now().Add(time.Duration(expires) * time.Second),
		Interval:        time.Duration(interval) * time.Second,
	}, nil
}

func verificationURICompleteOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func (h *HTTPProviderAdapter) PollToken(ctx context.Context, state *provider.DeviceFlowState) (*provider.TokenSet, error) {
	ticker := time.NewTicker(state.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			tr, err := h.pollOnce(ctx, state.DeviceCode)
			if err != nil {
				return nil, err
			}
			if tr == nil {
				if time.Now().After(state.ExpiresAt) {
					return nil, provider.ErrDeviceExpired
				}
				continue
			}
			return &provider.TokenSet{
				Provider:     "", // caller should set provider name when storing
				AccessToken:  tr.AccessToken,
				RefreshToken: tr.RefreshToken,
				Expiry:       time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
				TokenType:    tr.TokenType,
			}, nil
		}
	}
}

func (h *HTTPProviderAdapter) pollOnce(ctx context.Context, deviceCode string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	form.Set("device_code", deviceCode)
	form.Set("client_id", h.cfg.ClientID)
	req, err := http.NewRequestWithContext(ctx, "POST", h.cfg.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	if tr.Error != "" {
		switch tr.Error {
		case "authorization_pending":
			return nil, nil
		case "slow_down":
			return nil, nil
		case "access_denied":
			return nil, provider.ErrAuthDenied
		case "expired_token":
			return nil, provider.ErrDeviceExpired
		default:
			return nil, errors.New(tr.ErrorDesc)
		}
	}
	return &tr, nil
}

func (h *HTTPProviderAdapter) RefreshToken(ctx context.Context, refreshToken string) (*provider.TokenSet, error) {
	if refreshToken == "" {
		return nil, provider.ErrNoRefreshToken
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", h.cfg.ClientID)
	req, err := http.NewRequestWithContext(ctx, "POST", h.cfg.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	if tr.AccessToken == "" {
		return nil, provider.ErrNoRefreshToken
	}
	return &provider.TokenSet{
		Provider:     "",
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
		TokenType:    tr.TokenType,
	}, nil
}
