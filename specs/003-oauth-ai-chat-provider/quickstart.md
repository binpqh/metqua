# Quickstart: OAuth + AI Chat Provider

**Feature**: 003-oauth-ai-chat-provider | **Version**: v2.1.0

Get from zero to your first AI chat response in 5 minutes.

---

## Prerequisites

- `simple-cli` v2.1.0 installed (`simple-cli --version`)
- Your API server running with:
  - An OAuth 2.0 Device Authorization endpoint (RFC 8628)
  - An OpenAI-compatible Chat Completions endpoint
- A registered OAuth 2.0 client ID (public client, no secret)

---

## Step 1 — Configure Your Provider

Open (or create) the config file:

**Linux / macOS**
```
$XDG_CONFIG_HOME/simple-cli/config.yaml
# Default: ~/.config/simple-cli/config.yaml
```

**Windows**
```
%APPDATA%\simple-cli\config.yaml
```

Paste the following minimal configuration and fill in your values:

```yaml
default_provider: my-api

providers:
  my-api:
    client_id: "YOUR_CLIENT_ID"
    device_endpoint: "https://auth.example.com/oauth/device/code"
    token_endpoint: "https://auth.example.com/oauth/token"
    chat_endpoint: "https://api.example.com/v1/chat/completions"
    scopes: ["chat"]
    default_model: "gpt-4o"
```

To verify config is valid:
```
simple-cli auth status
```
Expected output:
```
Provider : my-api
Status   : Not logged in
```

---

## Step 2 — Authenticate

```
simple-cli auth login
```

The CLI initiates the Device Flow and displays:

```
Open this URL in your browser:

  https://auth.example.com/device

Enter code: ABCD-1234

Waiting for authorization...
```

1. Open the URL in your browser
2. Enter (or confirm) the displayed code
3. Approve the access request on your server's consent screen

Once approved:
```
✓  Logged in to my-api (user: user@example.com)
```

Your tokens are saved to `ConfigDir()/tokens.json` (mode 0600).

---

## Step 3 — Verify Login Status

```
simple-cli auth status
```

```
Provider : my-api
Status   : Logged in
User     : user@example.com
Expires  : 2026-03-25 10:00 UTC (in 23h 45m)
```

---

## Step 4 — Send Your First Message

```
simple-cli chat "What is the capital of France?"
```

Response streams to stdout incrementally:
```
The capital of France is Paris.
```

---

## Step 5 — Try Advanced Options

**Specify a model:**
```
simple-cli chat --model gpt-4o-mini "Summarize the history of Go in 2 sentences."
```

**Add a system prompt:**
```
simple-cli chat --system "You are a concise technical writer." "Explain variadic functions in Go."
```

**Multi-turn conversation (server must support):**
```
simple-cli chat --conversation conv-abc123 "What did we just discuss?"
```

**JSON output (buffered after stream completes):**
```
simple-cli chat --output json "Hello"
```
```json
{
  "provider": "my-api",
  "model": "gpt-4o",
  "content": "Hello! How can I help you today?",
  "streaming": true
}
```

**Pipe input to chat:**
```
cat myfile.go | simple-cli chat "Review this code for security issues:"
```

> Note: When stdin is a pipe, the piped content is appended to the message.

---

## Step 6 — Log Out

```
simple-cli auth logout
```
```
✓  Logged out from my-api
```

Tokens are deleted from `tokens.json`. Your config file is not modified.

---

## Multiple Providers

Add multiple provider blocks to your config:

```yaml
default_provider: work-api

providers:
  work-api:
    client_id: "prod-abc"
    device_endpoint: "https://auth.mycompany.com/device"
    token_endpoint: "https://auth.mycompany.com/token"
    chat_endpoint: "https://ai.mycompany.com/v1/chat/completions"
    scopes: ["ai:chat"]
    default_model: "gpt-4o"

  personal-api:
    client_id: "dev-xyz"
    device_endpoint: "https://auth.myservice.io/device"
    token_endpoint: "https://auth.myservice.io/token"
    chat_endpoint: "https://api.myservice.io/v1/chat/completions"
    scopes: []
    default_model: "claude-3-5-sonnet"
```

Authenticate and chat with a specific provider:
```
simple-cli auth login --provider personal-api
simple-cli chat --provider personal-api "Hello from my personal API"
```

---

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `provider not configured; check config file` | Add provider block to `config.yaml`; verify spelling matches `default_provider` |
| `provider "my-api" is missing required field: client_id` | Set `client_id` in config or env var `SIMPLE_CLI_PROVIDER_CLIENT_ID` |
| `Device code expired` | Re-run `auth login` and approve within the time limit (typically 5 minutes) |
| `Token expired. Run 'auth login' to continue.` | Run `simple-cli auth login` to get a fresh token |
| `stream closed unexpectedly` | Check your chat endpoint is reachable; verify token with `auth status` |
| Tokens not saved between sessions | Check `tokens.json` file permissions; ensure `ConfigDir()` is writable |

---

## Next Steps

- **Customize the CLI**: See [docs/customization.md](../../../docs/customization.md) to add new provider adapters, token backends, or chat transports
- **Config reference**: See [contracts/provider-config.md](contracts/provider-config.md) for the full config schema and all environment variable overrides
- **Interface contracts**: See [contracts/interfaces.md](contracts/interfaces.md) to implement `ProviderAdapter`, `TokenStore`, or `ChatBackend`
