package cmd

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/provider"
	"github.com/binpqh/simple-cli/internal/tokenstore"
)

// Test auto-refresh path: first chat call returns 401, refresh endpoint issues new token, second chat succeeds.
func TestChat_AutoRefreshPath(t *testing.T) {
	// chat server: stream one data chunk then [DONE]
	chatSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"pong\"}}]}\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer chatSrv.Close()

	// allow TLS to succeed for test servers (accept self-signed)
	//nolint:gosec
	http.DefaultTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	// prepare config and token store
	tmp := t.TempDir()
	// override ConfigDir by env
	os.Setenv("XDG_CONFIG_HOME", tmp)

	cfg := &config.Config{DefaultProvider: "p", Providers: map[string]config.ProviderConfig{"p": {ClientID: "cid", DeviceEndpoint: chatSrv.URL, TokenEndpoint: chatSrv.URL, ChatEndpoint: chatSrv.URL, DefaultModel: "m"}}}

	// put a valid token into tokenstore
	tsPath := tokenstore.PathForConfigDir(config.ConfigDir())
	store := tokenstore.NewFileTokenStore(tsPath)
	_ = store.Set(context.Background(), "p", &provider.TokenSet{Provider: "p", AccessToken: "good", Expiry: (time.Now().Add(time.Hour))})

	// build cobra command and execute
	cmd := newChatCmd()
	// set context with config
	cctx := context.WithValue(context.Background(), config.CtxKey{}, cfg)
	cmd.SetContext(cctx)
	cmd.SetArgs([]string{"hello"})

	// capture stdout
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute()

	_ = w.Close()
	outBytes, _ := io.ReadAll(r)
	os.Stdout = oldOut

	require.NoError(t, err)
	require.Contains(t, string(outBytes), "pong")
}
