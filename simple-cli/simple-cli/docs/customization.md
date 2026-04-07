# Customization Guide

This guide explains how to extend **simple-cli** without modifying core files.
Each section is independent — complete only the section you need.

---

## (a) Adding a New Sub-Command

1. Copy `cmd/example_cmd.go` to `cmd/mycommand.go`.

2. Rename the function and `Use` field:

   ```go
   // cmd/mycommand.go
   package cmd

   import "github.com/spf13/cobra"

   func newMyCmd() *cobra.Command {
       return &cobra.Command{
           Use:   "mycommand",
           Short: "Short description of mycommand",
           RunE: func(cmd *cobra.Command, args []string) error {
               // your logic here
               return nil
           },
       }
   }
   ```

3. Register the command in `cmd/root.go` inside `init()`:

   ```go
   func init() {
       // existing registrations ...
       rootCmd.AddCommand(newMyCmd())
   }
   ```

4. Build and verify:

   ```bash
   go build ./...
   simple-cli mycommand --help
   ```

---

## (b) Implementing a New OAuth Provider Adapter

All OAuth adapters implement the `provider.ProviderAdapter` interface
(see `internal/provider/provider.go`). The built-in `HTTPProviderAdapter`
handles RFC 8628 Device Authorization Flow; replace it to support any other OAuth flow.

1. Create a new package `internal/auth/<name>/<name>.go`:

   ```go
   package myauth

   import (
       "context"

       "github.com/binpqh/simple-cli/internal/config"
       "github.com/binpqh/simple-cli/internal/provider"
   )

   type MyProviderAdapter struct {
       cfg *config.ProviderConfig
   }

   func New(cfg *config.ProviderConfig) *MyProviderAdapter {
       return &MyProviderAdapter{cfg: cfg}
   }

   func (a *MyProviderAdapter) StartDeviceFlow(ctx context.Context) (*provider.DeviceFlowState, error) { /* ... */ }
   func (a *MyProviderAdapter) PollToken(ctx context.Context, s *provider.DeviceFlowState) (*provider.TokenSet, error) { /* ... */ }
   func (a *MyProviderAdapter) RefreshToken(ctx context.Context, rt string) (*provider.TokenSet, error) { /* ... */ }

   // Compile-time interface check
   var _ provider.ProviderAdapter = (*MyProviderAdapter)(nil)
   ```

2. Register the adapter in `cmd/auth/login.go` by adding an entry to `providerRegistry`:

   ```go
   import myauth "github.com/binpqh/simple-cli/internal/auth/myauth"

   // in providerRegistry map:
   "myprovider": func(cfg *config.ProviderConfig) provider.ProviderAdapter {
       return myauth.New(cfg)
   },
   ```

3. Add a matching entry in your `config.yaml`:

   ```yaml
   default_provider: myprovider
   providers:
     myprovider:
       client_id: "my-client-id"
       device_endpoint: "https://myprovider.example.com/device"
       token_endpoint: "https://myprovider.example.com/token"
       chat_endpoint: "https://api.myprovider.example.com/chat"
   ```

4. Test:

   ```bash
   simple-cli auth login --provider myprovider
   simple-cli auth status --provider myprovider
   ```

See `contracts/interfaces.md` for the full interface contract and error semantics.

---

## (c) Implementing a New Chat Backend

All chat backends implement the `provider.ChatBackend` interface.
The built-in `SSEChatBackend` speaks OpenAI-compatible SSE.

1. Create `internal/chat/<name>/<name>.go`:

   ```go
   package mybackend

   import (
       "context"

       "github.com/binpqh/simple-cli/internal/config"
       "github.com/binpqh/simple-cli/internal/provider"
   )

   type MyBackend struct {
       cfg *config.ProviderConfig
   }

   func New(cfg *config.ProviderConfig) *MyBackend { return &MyBackend{cfg: cfg} }

   func (b *MyBackend) Chat(ctx context.Context, req *provider.ChatRequest, token string) (<-chan provider.StreamEvent, error) {
       ch := make(chan provider.StreamEvent)
       go func() {
           defer close(ch)
           // send events ...
           ch <- provider.StreamEvent{Delta: "hello"}
           ch <- provider.StreamEvent{Done: true}
       }()
       return ch, nil
   }

   // Compile-time interface check
   var _ provider.ChatBackend = (*MyBackend)(nil)
   ```

2. Register the backend in `cmd/chat.go` by replacing or augmenting the backend instantiation:

   ```go
   import mybackend "github.com/binpqh/simple-cli/internal/chat/mybackend"

   // swap the default backend for a specific provider name:
   var backend provider.ChatBackend
   if providerName == "myprovider" {
       backend = mybackend.New(pc)
   } else {
       backend = chat.NewSSEChatBackend(pc)
   }
   ```

3. Test using `go test ./internal/chat/mybackend/...`.

---

## (d) Config File Reference

All provider settings live under `providers.<name>` in `config.yaml`:

```yaml
default_provider: my-api
providers:
  my-api:
    client_id: "your-client-id"
    device_endpoint: "https://auth.example.com/device"
    token_endpoint: "https://auth.example.com/token"
    chat_endpoint: "https://api.example.com/v1/chat/completions"
    scopes: ["chat", "offline_access"]
    default_model: "gpt-4o"
```

Override any field at runtime via environment variables (useful in CI):

| Env Variable                          | Overrides                       |
| ------------------------------------- | ------------------------------- |
| `SIMPLE_CLI_PROVIDER_CLIENT_ID`       | `providers.<n>.client_id`       |
| `SIMPLE_CLI_PROVIDER_DEVICE_ENDPOINT` | `providers.<n>.device_endpoint` |
| `SIMPLE_CLI_PROVIDER_TOKEN_ENDPOINT`  | `providers.<n>.token_endpoint`  |
| `SIMPLE_CLI_PROVIDER_CHAT_ENDPOINT`   | `providers.<n>.chat_endpoint`   |
| `SIMPLE_CLI_PROVIDER_DEFAULT_MODEL`   | `providers.<n>.default_model`   |

Tokens are persisted to `<config-dir>/tokens.json` (mode `0600`).

For the full interface and output contracts see:

- `specs/003-oauth-ai-chat-provider/contracts/interfaces.md`
- `specs/003-oauth-ai-chat-provider/contracts/provider-config.md`
- `docs/configuration.md`
