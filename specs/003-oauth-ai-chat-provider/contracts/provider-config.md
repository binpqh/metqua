# Contract: Provider Config Schema

**Branch**: `003-oauth-ai-chat-provider`
**Config file**: `$XDG_CONFIG_HOME/simple-cli/config.yaml` (Linux/macOS)
                 `%APPDATA%\simple-cli\config.yaml` (Windows)

---

## Full Schema

```yaml
# ─────────────────────────────────────────────────────────────────────────────
# simple-cli configuration
# ─────────────────────────────────────────────────────────────────────────────

# Global output settings (existing)
output: human           # human | json
log_level: warn         # debug | info | warn | error
no_color: false
quiet: false

# ─────────────────────────────────────────────────────────────────────────────
# AI Provider Configuration (new in v2.1.0)
# ─────────────────────────────────────────────────────────────────────────────

# The provider name to use when --provider flag is not specified.
# Must match one of the keys in the providers map below.
default_provider: my-api

providers:
  # Provider name (arbitrary, used as --provider value and in token storage)
  my-api:
    # OAuth 2.0 client identifier registered with your auth server
    client_id: "your-client-id-here"

    # RFC 8628 §3.1: Device Authorization Endpoint
    # POST with: client_id, scope
    # Returns: device_code, user_code, verification_uri, expires_in, interval
    device_endpoint: "https://auth.example.com/oauth/device/code"

    # RFC 6749 §4.1.3 / RFC 8628 §3.4: Token Endpoint
    # POST with: grant_type=urn:ietf:params:oauth:grant-type:device_code,
    #            device_code, client_id
    token_endpoint: "https://auth.example.com/oauth/token"

    # OpenAI-compatible Chat Completions endpoint
    # POST /v1/chat/completions (or custom path)
    chat_endpoint: "https://api.example.com/v1/chat/completions"

    # OAuth scopes to request during device flow
    # Leave empty ([]) if your provider does not require explicit scopes
    scopes: ["chat", "read"]

    # Default model name sent in ChatRequest.model
    # Can be overridden per-invocation with: chat --model <name>
    default_model: "gpt-4o"
```

---

## Multiple Providers Example

```yaml
default_provider: work-api

providers:
  work-api:
    client_id: "prod-client-abc"
    device_endpoint: "https://auth.mycompany.com/device"
    token_endpoint: "https://auth.mycompany.com/token"
    chat_endpoint: "https://ai.mycompany.com/v1/chat/completions"
    scopes: ["ai:chat"]
    default_model: "gpt-4o"

  personal-api:
    client_id: "dev-client-xyz"
    device_endpoint: "https://auth.myservice.io/device"
    token_endpoint: "https://auth.myservice.io/token"
    chat_endpoint: "https://api.myservice.io/v1/chat/completions"
    scopes: []
    default_model: "claude-3-5-sonnet"
```

Switch at runtime:
```
chat --provider personal-api "hello"
auth login --provider personal-api
```

---

## Environment Variable Overrides

Environment variables override config file values at runtime. They affect the **active** provider (resolved after `--provider` flag and `default_provider`).

| Variable | Overrides |
|----------|-----------|
| `SIMPLE_CLI_DEFAULT_PROVIDER` | `default_provider` |
| `SIMPLE_CLI_PROVIDER_CLIENT_ID` | active provider `client_id` |
| `SIMPLE_CLI_PROVIDER_DEVICE_ENDPOINT` | active provider `device_endpoint` |
| `SIMPLE_CLI_PROVIDER_TOKEN_ENDPOINT` | active provider `token_endpoint` |
| `SIMPLE_CLI_PROVIDER_CHAT_ENDPOINT` | active provider `chat_endpoint` |

All existing env vars (e.g. `SIMPLE_CLI_OUTPUT`, `SIMPLE_CLI_LOG_LEVEL`) remain unchanged.

---

## Validation Errors Output

When required config fields are missing, `auth login` and `chat` print to stderr:

Human mode:
```
Error: provider "my-api" is missing required field: client_id
Run 'simple-cli example --config' to see the full config example.
```

JSON mode:
```json
{
  "error": "provider \"my-api\" is missing required field: client_id",
  "hint": "Run 'simple-cli example --config' to see the full config example."
}
```

---

## Notes

- Provider names are normalized to **lowercase** by Viper during unmarshalling
- `client_secret` is intentionally absent — device flow (public clients, RFC 8628) does not require a secret
- The `tokens.json` file is separate from `config.yaml` and stores runtime credentials with mode `0600`
- To add a new provider type with custom OAuth flow, see [docs/customization.md](../../../docs/customization.md)
