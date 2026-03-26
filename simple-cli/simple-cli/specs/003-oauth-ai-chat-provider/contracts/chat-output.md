# Contract: Chat Command Output

**Branch**: `003-oauth-ai-chat-provider`
**Command**: `chat`
**Output modes**: human streaming (default), json buffered (`--output json`)

---

## `chat [message]`

Sends a message to the configured AI provider and streams the response.

### Human Output (default, streaming)

Content deltas are flushed to stdout incrementally as SSE chunks arrive. No prefix or label is added in human mode.

```
I can help you with writing code, answering questions, reviewing pull requests,
and much more. What would you like to work on today?
```

If not logged in:

```
Not authenticated. Run 'auth login' first.
```

Token expired (auto-refresh fails):

```
Token expired. Run 'auth login' to continue.
```

### JSON Output (`--output json`, buffered)

The full response is assembled in memory and emitted as a single JSON object after streaming completes.

```json
{
  "provider": "my-api",
  "model": "gpt-4o",
  "content": "I can help you with writing code, answering questions, reviewing pull requests, and much more. What would you like to work on today?",
  "streaming": true
}
```

Error (exit code 1):

```json
{
  "provider": "my-api",
  "model": "gpt-4o",
  "content": "",
  "streaming": true,
  "error": "stream closed unexpectedly"
}
```

---

## Flags

| Flag             | Short | Type   | Default                   | Description                                                   |
| ---------------- | ----- | ------ | ------------------------- | ------------------------------------------------------------- |
| `--provider`     |       | string | `config.default_provider` | Override active provider                                      |
| `--model`        | `-m`  | string | `provider.default_model`  | Model name sent to chat endpoint                              |
| `--conversation` | `-c`  | string | `""`                      | Conversation ID for multi-turn (opt-in, never auto-generated) |
| `--system`       |       | string | `""`                      | System prompt prepended to messages                           |
| `--output`       | `-o`  | string | `human`                   | Output format: `human` or `json`                              |

---

## Request Wire Format

The CLI sends the following JSON body to `POST {chat_endpoint}`:

```json
{
  "model": "gpt-4o",
  "messages": [
    { "role": "system", "content": "You are a helpful assistant." },
    { "role": "user", "content": "hello, what can you help me with?" }
  ],
  "stream": true,
  "conversation_id": "conv-abc123"
}
```

- `system` message is only included when `--system` flag is provided
- `conversation_id` is only included when `--conversation` flag is provided
- `stream` is always `true`; the CLI does not support non-streaming mode

### Required HTTP Headers

```
Authorization: Bearer <access_token>
Content-Type: application/json
Accept: text/event-stream
```

---

## SSE Wire Protocol

Server responds with `Content-Type: text/event-stream` or `application/x-ndjson`. The CLI handles both.

Each server-sent event:

```
data: {"id":"chatcmpl-abc","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":"I "},"finish_reason":null}]}

data: {"id":"chatcmpl-abc","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"can "},"finish_reason":null}]}

data: [DONE]
```

Termination signals:

- `data: [DONE]` — normal completion
- `choices[0].finish_reason == "stop"` — model finished (may precede `[DONE]`)
- `choices[0].finish_reason == "length"` — max tokens reached (treated as normal completion)
- HTTP error (4xx/5xx) before SSE begins — fatal error, exit code 1

---

## Exit Codes

| Code | Meaning                                       |
| ---- | --------------------------------------------- |
| 0    | Chat completed successfully                   |
| 1    | Auth error (not logged in / token expired)    |
| 1    | Stream error (HTTP error / malformed SSE)     |
| 2    | Configuration error (provider not configured) |
