package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/provider"
	"github.com/binpqh/simple-cli/internal/tokenstore"
)

// TestChat_AutoRefreshPath ensures that when a chat request returns 401, the CLI
// refreshes tokens and the token store is updated with the new access token.
func TestChat_AutoRefresh_UpdatesTokenStore(t *testing.T) {
	// temp config dir
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// create a dummy provider config that points to our test servers
	// refresh server: returns new tokens (use TLS so ValidateProviderConfig accepts URL)
	refreshSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "new-access-token",
			"refresh_token": "new-refresh-token",
			"expires_in":    3600,
		})
	}))
	defer refreshSrv.Close()

	// chat server: first call -> 401, second call -> SSE stream with a simple data payload (TLS)
	calls := 0
	chatSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("data: {\"type\": \"response\", \"content\": \"hi\"}\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer chatSrv.Close()

	// allow TLS to succeed for test servers (accept self-signed)
	//nolint:gosec
	http.DefaultTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	// write config file with one provider
	cfg := &config.Config{
		DefaultProvider: "test",
		Providers: map[string]config.ProviderConfig{
			"test": {
				ClientID:       "cid",
				DeviceEndpoint: refreshSrv.URL,
				TokenEndpoint:  refreshSrv.URL,
				ChatEndpoint:   chatSrv.URL,
				DefaultModel:   "gpt-test",
			},
		},
	}
	cfgDir := filepath.Join(tmpDir, "simple-cli")
	os.MkdirAll(cfgDir, 0o755)
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	f, err := os.Create(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		t.Fatal(err)
	}

	// seed tokenstore with expired token and refresh_token
	storePath := tokenstore.PathForConfigDir(config.ConfigDir())
	ts := tokenstore.NewFileTokenStore(storePath)
	initial := &provider.TokenSet{
		Provider:     "test",
		AccessToken:  "old",
		RefreshToken: "rt-old",
		Expiry:       time.Now().Add(-time.Hour),
	}
	if err := ts.Set(context.Background(), "test", initial); err != nil {
		t.Fatal(err)
	}

	// run chat command: use flags to select provider and message
	cmd := newChatCmd()
	cmd.SetArgs([]string{"-m", "gpt-test", "--provider", "test", "hello"})
	// set context with config (root normally injects this)
	cctx := context.WithValue(context.Background(), config.CtxKey{}, cfg)
	cmd.SetContext(cctx)
	// execute
	if err := cmd.Execute(); err != nil {
		t.Fatalf("chat command failed: %v", err)
	}

	// ensure tokenstore updated
	got, err := ts.Get(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "new-access-token" {
		t.Fatalf("expected access token to be refreshed, got %q", got.AccessToken)
	}
}
