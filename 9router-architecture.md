# 9Router Architecture (GitHub Copilot Focus)

This note is intentionally scoped to one provider only: GitHub Copilot.

Goal:
- Understand the GitHub request path in 9Router.
- Copy the token in/out tracking and cost calculation pattern.
- Reuse RTK and Caveman style optimizers to reduce token usage.

## Provider Scope

- Client model prefix `gh/` resolves to provider id `github`.
- Runtime executes through GitHub-specific executor logic.
- Usage and cost persistence are shared, provider-agnostic modules.

## End-to-End Flow (GitHub Only)

```mermaid
flowchart LR
	A[Client POST /v1/chat/completions] --> B[src/sse/handlers/chat.js]
	B --> C[Resolve model/provider]
	C --> D[Get GitHub credentials]
	D --> E[open-sse/handlers/chatCore.js]
	E --> F[translate request if needed]
	F --> G[RTK compress tool results]
	G --> H[Caveman inject terse system instruction]
	H --> I[open-sse/executors/github.js]
	I --> J[api.githubcopilot.com]
	J --> K[Stream/JSON response]
	K --> L[open-sse/utils/stream.js or nonStreamingHandler.js]
	L --> M[extract usage + estimate fallback]
	M --> N[save usage + compute cost]
```

## Key Files for GitHub Path

- Entry and orchestration:
	- `src/sse/handlers/chat.js`
	- `open-sse/handlers/chatCore.js`
- GitHub provider behavior:
	- `open-sse/executors/github.js`
	- `open-sse/config/providers.js`
	- `open-sse/config/appConstants.js`
	- `open-sse/services/tokenRefresh.js`
- Usage and cost:
	- `open-sse/utils/usageTracking.js`
	- `open-sse/utils/stream.js`
	- `open-sse/handlers/chatCore/nonStreamingHandler.js`
	- `open-sse/handlers/chatCore/requestDetail.js`
	- `src/lib/db/repos/usageRepo.js`
	- `src/shared/constants/pricing.js`
	- `src/lib/db/repos/pricingRepo.js`
- Token savers:
	- `open-sse/rtk/index.js`
	- `open-sse/rtk/autodetect.js`
	- `open-sse/rtk/caveman.js`
	- `open-sse/rtk/cavemanPrompts.js`

## GitHub-Specific Request + Auth Behavior

1. Model/provider resolution
- `gh/<model>` is mapped to provider `github` by alias mapping.

2. Credential selection
- Account chosen by `src/sse/services/auth.js`.
- GitHub credential payload includes:
	- `accessToken` (GitHub OAuth token)
	- `refreshToken`
	- optional `providerSpecificData.copilotToken`

3. GitHub executor dispatch
- `open-sse/executors/github.js` applies provider-specific rules:
	- Builds Copilot headers to mimic VS Code client.
	- Uses `copilotToken || accessToken` for bearer auth.
	- Sanitizes message content to supported part types.
	- Handles route switching between:
		- `/chat/completions`
		- `/responses` for known models that require it.

4. Token refresh strategy
- Refreshes Copilot token via GitHub API using access token.
- If needed, refreshes GitHub access token first, then Copilot token.

## Token Tracking Pattern You Can Copy

9Router tracks usage from both streaming and non-streaming responses, with fallback estimation when provider usage is missing.

### Streaming path

In `open-sse/utils/stream.js`:
- Parse each SSE `data:` event.
- Call `extractUsage(parsed)` for every chunk.
- Accumulate output content length while streaming.
- On finish chunk:
	- If usage missing, call `estimateUsage(requestBody, contentLen, format)`.
	- Otherwise apply `addBufferToUsage` before forwarding usage to client.
- At flush:
	- `logUsage(provider, usage, model, connectionId, apiKey)`.

`logUsage` persists usage immediately through DB API and writes request log status.

### Non-streaming path

In `open-sse/handlers/chatCore/nonStreamingHandler.js`:
- Parse JSON (or convert SSE to JSON first).
- `extractUsageFromResponse(responseBody)`.
- `saveUsageStats(...)` persists normalized prompt/completion tokens.

### Estimation fallback

In `open-sse/utils/usageTracking.js`:
- Input tokens estimate: `ceil(JSON.stringify(body).length / 4)`.
- Output tokens estimate: `floor(contentLength / 4)`.

### Buffer behavior

`addBufferToUsage` adds safety buffer (currently +2000 input/prompt tokens) before sending usage to client-side consumers. Raw usage is still retained internally for logging/cost flows where applicable.

## Cost Calculation Pattern You Can Copy

Cost is computed during persistence in `src/lib/db/repos/usageRepo.js`.

Formula (rates are $ per 1M tokens):

```text
input_tokens = prompt_tokens or input_tokens
cached_tokens = cached_tokens or cache_read_input_tokens
non_cached_input = max(0, input_tokens - cached_tokens)

cost = non_cached_input * input_rate / 1_000_000
		 + cached_tokens * cached_rate / 1_000_000
		 + output_tokens * output_rate / 1_000_000
		 + reasoning_tokens * reasoning_rate / 1_000_000
		 + cache_creation_tokens * cache_creation_rate / 1_000_000
```

Pricing resolution order:
1. User override from DB (`pricing` kv).
2. Static model/provider pricing (`src/shared/constants/pricing.js`).
3. Pattern-based pricing fallback.

Persistence tables used:
- `usageHistory` (raw per-request entries).
- `usageDaily` (pre-aggregated daily stats for dashboards).

## RTK Input-Token Reduction Pattern

RTK is applied right before dispatch in `open-sse/handlers/chatCore.js`:

- `compressMessages(translatedBody, rtkEnabled)`
- Targets tool-heavy payloads (tool_result, tool outputs, responses output blocks).
- Auto-detects structure and chooses filter (`git diff`, `grep`, `ls`, `tree`, build logs, etc.).
- Safety guarantees:
	- Skip small blobs.
	- Skip if output grows.
	- On filter failure, passthrough original text.

This means semantic request flow is preserved while reducing prompt size.

## Caveman Output-Token Reduction Pattern

Caveman is also applied pre-dispatch in `open-sse/handlers/chatCore.js`:

- `injectCaveman(translatedBody, finalFormat, cavemanLevel)`
- Injects terse instruction into system slot according to final request format:
	- OpenAI/messages or responses instructions
	- Claude system blocks
	- Gemini system instruction

Levels are defined in `open-sse/rtk/cavemanPrompts.js`:
- `lite`
- `full`
- `ultra`

Design intent is to reduce verbosity while preserving technical content.

## Practical Copy Blueprint (GitHub-Only)

If you are building your own gateway and only need GitHub:

1. Build GitHub executor
- Add one executor with:
	- Copilot header fingerprint
	- model-specific request sanitization
	- optional `/chat/completions` -> `/responses` switch

2. Add pre-dispatch optimization stage
- Run RTK compression on tool outputs.
- Inject Caveman prompt into system/instructions.

3. Add streaming usage tracker
- Parse SSE lines.
- Extract usage chunks when present.
- Accumulate content length for fallback estimation.
- Persist final usage on stream flush.

4. Add non-streaming usage tracker
- Extract usage from response body.
- Normalize token field names to one internal shape.
- Persist to usage history table.

5. Add pricing and cost layer
- Keep pricing table in DB + defaults in code.
- Compute cost at write-time.
- Store both raw token JSON and normalized prompt/completion fields.

6. Expose analytics APIs
- Stats endpoint by period (`today`, `24h`, `7d`, etc.).
- Chart endpoint with bucketed token/cost trend.
- Optional logs endpoint.

## GitHub-Only Caveat to Check

In this codebase, provider-specific static pricing override uses alias key `gh` in `PROVIDER_PRICING`, while runtime usage entries are stored with provider id `github`. If you want provider override pricing to apply for GitHub entries, normalize provider keying (alias vs id) consistently in your own implementation.
