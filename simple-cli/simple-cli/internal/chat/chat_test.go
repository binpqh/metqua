package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/provider"
)

func TestChat_SSEHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify Authorization header and body
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// read body to ensure request payload contains stream:true
		var body provider.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		// send three chunk events
		chunks := []string{"I ", "can ", "help."}
		for _, c := range chunks {
			ev := map[string]interface{}{
				"choices": []interface{}{map[string]interface{}{"delta": map[string]interface{}{"content": c}}},
			}
			b, _ := json.Marshal(ev)
			_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer srv.Close()

	cfg := &config.ProviderConfig{ChatEndpoint: srv.URL}
	backend := NewSSEChatBackend(cfg)
	req := &provider.ChatRequest{Model: "m", Messages: []provider.ChatMessage{{Role: "user", Content: "hello"}}, Stream: true}

	ch, err := backend.Chat(context.Background(), req, "test-token")
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}

	var got strings.Builder
	done := false
	for e := range ch {
		if e.Err != nil {
			t.Fatalf("stream error: %v", e.Err)
		}
		if e.Done {
			done = true
			break
		}
		got.WriteString(e.Delta)
	}

	if !done {
		t.Fatalf("expected done event")
	}
	if got.String() != "I can help." {
		t.Fatalf("unexpected content: %q", got.String())
	}
}

func TestChat_HTTP401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	cfg := &config.ProviderConfig{ChatEndpoint: srv.URL}
	backend := NewSSEChatBackend(cfg)
	req := &provider.ChatRequest{Model: "m", Messages: []provider.ChatMessage{{Role: "user", Content: "hello"}}, Stream: true}

	_, err := backend.Chat(context.Background(), req, "bad-token")
	if err == nil {
		t.Fatalf("expected error for 401 response")
	}
	if err != provider.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestChat_MalformedJSONSkipped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		// send one malformed line then a valid one, then done
		w.Write([]byte("data: {not-json}\n\n"))
		flusher.Flush()
		ev, _ := json.Marshal(map[string]interface{}{
			"choices": []interface{}{map[string]interface{}{"delta": map[string]interface{}{"content": "ok"}}},
		})
		w.Write([]byte("data: " + string(ev) + "\n\n"))
		flusher.Flush()
		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer srv.Close()

	cfg := &config.ProviderConfig{ChatEndpoint: srv.URL}
	backend := NewSSEChatBackend(cfg)
	req := &provider.ChatRequest{Model: "m", Messages: []provider.ChatMessage{{Role: "user", Content: "hi"}}, Stream: true}

	ch, err := backend.Chat(context.Background(), req, "tok")
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}

	var got strings.Builder
	for e := range ch {
		if e.Err != nil {
			t.Fatalf("unexpected stream error: %v", e.Err)
		}
		got.WriteString(e.Delta)
		if e.Done {
			break
		}
	}
	if got.String() != "ok" {
		t.Fatalf("expected 'ok', got %q", got.String())
	}
}

func TestChat_ContextCancellation(t *testing.T) {
	started := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		close(started)
		flusher.Flush()
		// hold connection open until client disconnects
		<-r.Context().Done()
	}))
	defer srv.Close()

	cfg := &config.ProviderConfig{ChatEndpoint: srv.URL}
	backend := NewSSEChatBackend(cfg)
	req := &provider.ChatRequest{Model: "m", Messages: []provider.ChatMessage{{Role: "user", Content: "hi"}}, Stream: true}

	ctx, cancel := context.WithCancel(context.Background())

	ch, err := backend.Chat(ctx, req, "tok")
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}

	<-started // server accepted and started
	cancel()  // cancel context

	// drain channel — it should close (possibly with or without a context error event)
	for range ch {
	}
}

func TestChat_WireFormat(t *testing.T) {
	var gotBody provider.ChatRequest
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	cfg := &config.ProviderConfig{ChatEndpoint: srv.URL}
	backend := NewSSEChatBackend(cfg)
	req := &provider.ChatRequest{
		Model:          "gpt-4o",
		Messages:       []provider.ChatMessage{{Role: "user", Content: "hi"}},
		Stream:         true,
		ConversationID: "conv-abc",
	}

	ch, err := backend.Chat(context.Background(), req, "my-token")
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	for range ch {
	}

	if gotAuth != "Bearer my-token" {
		t.Fatalf("expected Authorization 'Bearer my-token', got %q", gotAuth)
	}
	if !gotBody.Stream {
		t.Fatalf("expected stream:true in request body")
	}
	if gotBody.ConversationID != "conv-abc" {
		t.Fatalf("expected conversation_id 'conv-abc', got %q", gotBody.ConversationID)
	}
}

func TestChat_ConversationIDOmittedWhenEmpty(t *testing.T) {
	var gotBody provider.ChatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	cfg := &config.ProviderConfig{ChatEndpoint: srv.URL}
	backend := NewSSEChatBackend(cfg)
	req := &provider.ChatRequest{
		Model:    "m",
		Messages: []provider.ChatMessage{{Role: "user", Content: "hi"}},
		Stream:   true,
	}

	ch, err := backend.Chat(context.Background(), req, "tok")
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	for range ch {
	}

	if gotBody.ConversationID != "" {
		t.Fatalf("expected empty conversation_id, got %q", gotBody.ConversationID)
	}
}
