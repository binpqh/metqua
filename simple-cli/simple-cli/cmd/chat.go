package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	authpkg "github.com/binpqh/simple-cli/internal/auth"
	"github.com/binpqh/simple-cli/internal/chat"
	"github.com/binpqh/simple-cli/internal/config"
	"github.com/binpqh/simple-cli/internal/provider"
	"github.com/binpqh/simple-cli/internal/tokenstore"
)

func newChatCmd() *cobra.Command {
	var providerName string
	var model string
	var conversation string
	var system string

	cmd := &cobra.Command{
		Use:   "chat [message]",
		Short: "Send a message to the configured AI provider and stream the response",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ctx.Value(config.CtxKey{}).(*config.Config)

			// resolve provider name early so env var overrides can be bound to the active provider
			effectiveProvider := providerName
			if effectiveProvider == "" {
				effectiveProvider = cfg.DefaultProvider
			}
			// bind provider-specific env overrides to the active provider keys
			_ = v.BindEnv("providers."+effectiveProvider+".client_id", "SIMPLE_CLI_PROVIDER_CLIENT_ID")
			_ = v.BindEnv("providers."+effectiveProvider+".device_endpoint", "SIMPLE_CLI_PROVIDER_DEVICE_ENDPOINT")
			_ = v.BindEnv("providers."+effectiveProvider+".token_endpoint", "SIMPLE_CLI_PROVIDER_TOKEN_ENDPOINT")
			_ = v.BindEnv("providers."+effectiveProvider+".chat_endpoint", "SIMPLE_CLI_PROVIDER_CHAT_ENDPOINT")
			_ = v.BindEnv("providers."+effectiveProvider+".default_model", "SIMPLE_CLI_PROVIDER_DEFAULT_MODEL")

			pc, err := cfg.ActiveProvider(providerName)
			if err != nil {
				return err
			}
			if err := config.ValidateProviderConfig(pc, cfg.Insecure); err != nil {
				return err
			}

			// resolve provider name if empty
			if providerName == "" {
				providerName = cfg.DefaultProvider
			}

			// resolve model
			if model == "" {
				model = pc.DefaultModel
			}

			// build message
			var userMsg string
			if len(args) > 0 {
				userMsg = strings.Join(args, " ")
			}
			// if stdin is piped, append it
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				// read all
				b, _ := io.ReadAll(os.Stdin)
				if len(strings.TrimSpace(string(b))) > 0 {
					if userMsg != "" {
						userMsg = userMsg + "\n" + string(b)
					} else {
						userMsg = string(b)
					}
				}
			}
			if userMsg == "" {
				return errors.New("no message provided")
			}

			slog.Debug("chat: resolved provider", "provider", providerName, "model", model)

			// token store
			store := tokenstore.NewFileTokenStore(tokenstore.PathForConfigDir(config.ConfigDir()))
			ts, err := store.Get(ctx, providerName)
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			// refresh if expired
			if ts.IsExpired() {
				slog.Debug("chat: token expired, attempting refresh", "provider", providerName)
				adapter := authpkg.NewHTTPProviderAdapter(pc)
				newts, err := adapter.RefreshToken(ctx, ts.RefreshToken)
				if err == nil && newts != nil {
					newts.Provider = providerName
					_ = store.Set(ctx, providerName, newts)
					ts = newts
				}
			}

			buildReq := func() *provider.ChatRequest {
				msgs := []provider.ChatMessage{}
				if system != "" {
					msgs = append(msgs, provider.ChatMessage{Role: "system", Content: system})
				}
				msgs = append(msgs, provider.ChatMessage{Role: "user", Content: userMsg})
				return &provider.ChatRequest{Model: model, Messages: msgs, Stream: true, ConversationID: conversation}
			}

			// attempt chat, with one refresh-retry on ErrUnauthorized
			backend := chat.NewSSEChatBackend(pc)
			attempt := 0
			var contentBuilder strings.Builder
			for {
				attempt++
				req := buildReq()
				ch, err := backend.Chat(ctx, req, ts.AccessToken)
				if err != nil {
					if errors.Is(err, provider.ErrUnauthorized) && attempt == 1 {
						// try refresh
						adapter := authpkg.NewHTTPProviderAdapter(pc)
						newts, rerr := adapter.RefreshToken(ctx, ts.RefreshToken)
						if rerr == nil && newts != nil {
							newts.Provider = providerName
							_ = store.Set(ctx, providerName, newts)
							ts = newts
							continue // retry
						}
					}
					return err
				}

				// stream
				done := false
				for e := range ch {
					if e.Err != nil {
						return e.Err
					}
					if e.Done {
						done = true
						break
					}
					// human mode vs json
					slog.Debug("chat: SSE chunk received", "length", len(e.Delta))
					if cfg.Output == "json" {
						contentBuilder.WriteString(e.Delta)
					} else {
						fmt.Fprint(os.Stdout, e.Delta)
					}
				}
				if !done {
					return errors.New("stream closed unexpectedly")
				}
				break
			}

			if cfg.Output == "json" {
				out := map[string]interface{}{
					"provider":  providerName,
					"model":     model,
					"content":   contentBuilder.String(),
					"streaming": true,
				}
				enc := json.NewEncoder(os.Stdout)
				return enc.Encode(out)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name from config (default: default_provider)")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model name to use (overrides provider.default_model)")
	cmd.Flags().StringVarP(&conversation, "conversation", "c", "", "Conversation id for multi-turn (opt-in)")
	cmd.Flags().StringVar(&system, "system", "", "System prompt to prepend to messages")
	return cmd
}
