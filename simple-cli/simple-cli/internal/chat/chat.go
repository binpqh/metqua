package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/provider"
)

// SSEChatBackend streams OpenAI-compatible SSE chat responses.
type SSEChatBackend struct {
	client *http.Client
	cfg    *config.ProviderConfig
}

func NewSSEChatBackend(cfg *config.ProviderConfig) *SSEChatBackend {
	return &SSEChatBackend{client: http.DefaultClient, cfg: cfg}
}

// Chat posts the chat request and returns a channel of StreamEvent.
func (s *SSEChatBackend) Chat(ctx context.Context, req *provider.ChatRequest, token string) (<-chan provider.StreamEvent, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.cfg.ChatEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chat request: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, provider.ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("chat http status: %d", resp.StatusCode)
	}

	ch := make(chan provider.StreamEvent)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				ch <- provider.StreamEvent{Done: true}
				return
			}
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(data), &obj); err != nil {
				// skip malformed JSON
				continue
			}
			if choicesI, ok := obj["choices"].([]interface{}); ok && len(choicesI) > 0 {
				if c0, ok := choicesI[0].(map[string]interface{}); ok {
					if deltaI, ok := c0["delta"].(map[string]interface{}); ok {
						if content, ok := deltaI["content"].(string); ok && content != "" {
							ch <- provider.StreamEvent{Delta: content}
						}
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- provider.StreamEvent{Err: err}
		}
	}()

	return ch, nil
}
