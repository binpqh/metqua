# Contract: Auth Command Output

**Branch**: `003-oauth-ai-chat-provider`
**Commands covered**: `auth login`, `auth logout`, `auth status`
**Output modes**: human (default), json (`--output json`)

---

## `auth login`

Authenticates the user against the configured provider via RFC 8628 Device Flow.

### Human Output (during flow — stdout)

```
Open this URL in your browser:

  https://auth.example.com/device

Enter code: ABCD-1234

Waiting for authorization...
```

After polling completes:

```
✓  Logged in to my-api (user: user@example.com)
```

On timeout:

```
✗  Device code expired. Run 'auth login' again.
```

On denial:

```
✗  Authorization denied. Run 'auth login' again.
```

### JSON Output (`--output json`)

Success:
```json
{
  "provider": "my-api",
  "success": true,
  "user_id": "user@example.com",
  "message": "Logged in successfully"
}
```

Timeout (exit code 1):
```json
{
  "provider": "my-api",
  "success": false,
  "message": "Device code expired"
}
```

Denial (exit code 1):
```json
{
  "provider": "my-api",
  "success": false,
  "message": "Authorization denied"
}
```

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--provider` | string | Override active provider (default: `config.default_provider`) |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Login successful |
| 1 | Auth failed (timeout / denial / config error) |
| 2 | Misconfigured provider (missing client_id or endpoints) |

---

## `auth logout`

Deletes stored tokens for the given provider.

### Human Output

```
✓  Logged out from my-api
```

If not logged in:

```
Not logged in to my-api
```

### JSON Output

```json
{
  "provider": "my-api",
  "success": true,
  "message": "Logged out successfully"
}
```

Not logged in (exit code 0 — idempotent):
```json
{
  "provider": "my-api",
  "success": true,
  "message": "Not logged in"
}
```

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--provider` | string | Override active provider |
| `--all` | bool | Logout from all configured providers |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Logout successful (including idempotent) |
| 1 | Storage error (cannot delete token file) |

---

## `auth status`

Displays current authentication state for the active provider.

### Human Output

Logged in, token valid:
```
Provider : my-api
Status   : Logged in
User     : user@example.com
Expires  : 2026-03-25 10:00 UTC (in 23h 45m)
```

Logged in but token expired:
```
Provider : my-api
Status   : Token expired — run 'auth login' to refresh
User     : user@example.com
Expired  : 2026-03-24 08:00 UTC (3h ago)
```

Not logged in:
```
Provider : my-api
Status   : Not logged in
```

### JSON Output

```json
{
  "provider": "my-api",
  "logged_in": true,
  "user_id": "user@example.com",
  "expires_at": "2026-03-25T10:00:00Z",
  "expired": false
}
```

Not logged in:
```json
{
  "provider": "my-api",
  "logged_in": false,
  "expired": false
}
```

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--provider` | string | Override active provider |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Status retrieved (logged in or not) |
| 1 | Provider not configured |
