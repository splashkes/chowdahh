# Chowdahh Public API Clarity Plan

## Goal

Make Chowdahh's public news API obvious, scriptable, and safe enough for real use:

- Anonymous public stream access stays available with no signup.
- First-run usage is GET-only and copy-paste friendly.
- A user can upgrade a GET stream request with a personal key.
- Production clients are clearly guided to `Authorization: Bearer ...`.
- Existing Agent API, web stream, and Bearer-token clients remain compatible.

This plan replaces the earlier implementation outline. It keeps the same product objective, but tightens security, envelope shape, auth precedence, and file targets based on the current repo.

## Current Reality

Verified source and live behavior:

- Agent API service is in `cmd/agentapi`.
- Agent API route group is `/api/v1`.
- Existing public stream routes:
  - `GET /api/v1/streams`
  - `GET /api/v1/streams/{slug}`
  - `GET /api/v1/streams/latest?limit=1`
- Missing ergonomic alias:
  - `GET /api/v1/stream` currently returns 404.
- Existing auth middleware:
  - `internal/agentapi/authmw.go`
  - Bearer person tokens use `database.GetUserByPAT`.
  - Curator tokens use `database.GetCuratorByAPIKey`.
- Existing identity/context code:
  - `internal/agentapi/context.go`
- Existing envelope and response writers:
  - `internal/agentapi/envelope.go`
  - There is no separate `internal/agentapi/response.go`.
- Existing guidance:
  - `internal/agentapi/guidance/engine.go`
- Existing middleware logger logs `r.URL.Path`, not query strings.
- Web analytics can store raw query strings in `models.PageView.QueryParams`, but that middleware is on the web service, not the Agent API service. Do not rely on that separation as a security guarantee.
- Cloudflared routes `/api/v1/*` to the Agent API, except specific web-owned exceptions like `/api/v1/tokens*`.

## Product Decision

Use the existing Agent API as the canonical public API. Do not create a second API.

Add `GET /api/v1/stream` as the public default stream alias. Treat it as the main documentation path.

Support `?key=` only as an onboarding convenience for safe methods. It is not the preferred production auth mechanism.

Production recommendation remains:

```bash
curl \
  -H 'Authorization: Bearer PASTE_YOUR_KEY_HERE' \
  'https://chowdahh.com/api/v1/stream?limit=10'
```

## Public API Shape

### Anonymous Stream

```bash
curl 'https://chowdahh.com/api/v1/stream?limit=10'
```

Expected:

- HTTP 200.
- Standard successful Agent API envelope.
- `guidance.account_state.auth_mode` is `anonymous`.
- Guidance says anonymous access does not require a key.
- Guidance includes upgrade examples using placeholders only.

### Personalized Stream With Header

```bash
curl \
  -H 'Authorization: Bearer PASTE_YOUR_KEY_HERE' \
  'https://chowdahh.com/api/v1/stream?limit=10'
```

Expected:

- HTTP 200 for a valid person token.
- Standard successful Agent API envelope.
- `guidance.account_state.auth_mode` remains the existing schema value: `person_token`.
- If guidance exposes auth source, it is separate from auth mode, for example `auth_source: "bearer"`.

### Personalized Stream With Query Key

```bash
curl 'https://chowdahh.com/api/v1/stream?limit=10&key=PASTE_YOUR_KEY_HERE'
```

Expected:

- HTTP 200 for a valid person token.
- Standard successful Agent API envelope.
- Same person identity and rate-limit semantics as Bearer auth.
- `guidance.account_state.auth_mode` is `person_token`.
- If guidance exposes auth source, use `auth_source: "query_key"`.
- Response must include no-store and no-referrer safeguards listed below.

### Invalid Query Key

```bash
curl 'https://chowdahh.com/api/v1/stream?key=bad'
```

Expected:

- HTTP 401.
- Standard Agent API error envelope.
- Do not fall back to anonymous.
- Do not include the supplied key in response body, logs, metrics labels, analytics fields, or test names.

Exact response shape should follow `internal/agentapi/envelope.go`:

```json
{
  "error": {
    "code": "unauthorized",
    "message": "Invalid or expired key parameter."
  },
  "guidance": {
    "status_explanation": "The key parameter was provided but is invalid or expired. Remove key for anonymous public access, or paste a valid personal key.",
    "next_best_actions": [
      {
        "action_id": "paste_valid_person_key_or_remove_key",
        "title": "Use a valid key or remove it",
        "why": "Anonymous public access works without a key. Personalization requires a valid Chowdahh person token.",
        "kind": "auth_recovery",
        "priority": "high",
        "available": true,
        "user_facing_prompt": "Paste a valid personal key, or remove the key parameter to use the public stream."
      }
    ],
    "account_state": {
      "auth_mode": "anonymous"
    }
  },
  "meta": {
    "request_id": "..."
  }
}
```

## Security Contract

Query-string keys are accepted only because they make first-run use easy for people, agents, no-code tools, and terminals. They are less safe than headers.

Implementation must include these safeguards:

1. Accept `?key=` only for `GET` and `HEAD`.
2. If `?key=` is present on `POST`, `PUT`, `PATCH`, or `DELETE`, return `400 invalid_request`. Do not silently ignore it.
3. If `Authorization` and `?key=` are both present, return `400 invalid_request` with a clear mixed-credentials message. Do not choose one silently.
4. If `?key=` is present but empty, return `401 unauthorized`.
5. If `?key=` is present but invalid or expired, return `401 unauthorized`.
6. Never log the full key.
7. Never echo the full key.
8. Never put the full key in metrics labels.
9. Never persist the full key in analytics or page-view query fields.
10. For any request where `key` is present, set:

```text
Cache-Control: no-store, private
Pragma: no-cache
Referrer-Policy: no-referrer
```

11. For authenticated responses, also set:

```text
Vary: Authorization
```

12. Error paths must set the same no-store and no-referrer headers when `key` is present.
13. Docs must warn that URL keys can appear in browser history, proxies, and infrastructure logs. Headers are the recommended production form.
14. If future Agent API request logging starts storing query strings, `key` must be redacted before storage.

## Auth Semantics

Preserve current auth mode values:

- `anonymous`
- `person_token`
- `curator_token`

Do not replace them with `authenticated` or `personalized`.

If the response needs to communicate how auth was supplied, add a separate field. Recommended additive field:

```go
type AccountState struct {
    AuthMode        string                `json:"auth_mode"`
    AuthSource      string                `json:"auth_source,omitempty"` // anonymous, bearer, query_key
    PlanTier        string                `json:"plan_tier,omitempty"`
    RateLimit       *RateState            `json:"rate_limit,omitempty"`
    Upgrade         *AuthUpgrade          `json:"upgrade,omitempty"`
    Personalization *PersonalizationState `json:"personalization,omitempty"`
}
```

Recommended supporting structs:

```go
type AuthUpgrade struct {
    Required            bool   `json:"required"`
    Label               string `json:"label"`
    PasteKeyURLTemplate string `json:"paste_key_url_template"`
    RecommendedHeader   string `json:"recommended_header"`
    SecurityNote        string `json:"security_note"`
}

type PersonalizationState struct {
    Enabled bool   `json:"enabled"`
    Source  string `json:"source,omitempty"` // person_token
}
```

For anonymous responses:

```json
{
  "auth_mode": "anonymous",
  "auth_source": "anonymous",
  "upgrade": {
    "required": false,
    "label": "Add a key for personalized news",
    "paste_key_url_template": "https://chowdahh.com/api/v1/stream?key=PASTE_YOUR_KEY_HERE",
    "recommended_header": "Authorization: Bearer PASTE_YOUR_KEY_HERE",
    "security_note": "For production apps, prefer the Authorization header over query-string keys."
  }
}
```

For valid Bearer:

```json
{
  "auth_mode": "person_token",
  "auth_source": "bearer",
  "personalization": {
    "enabled": true,
    "source": "person_token"
  }
}
```

For valid query key:

```json
{
  "auth_mode": "person_token",
  "auth_source": "query_key",
  "personalization": {
    "enabled": true,
    "source": "person_token"
  }
}
```

Only set `personalization.enabled=true` if current stream code actually applies user-specific behavior or can safely say the response is authenticated for personalization. If the current code only authenticates but does not personalize ranking yet, use `enabled=false` and `source:"person_token"` or omit the block.

## Back-Prompting And Suggested Actions

The API should not only return cards. It should give agents safe, concrete next moves they can offer back to the user.

Use the existing guidance shape:

- `guidance.next_best_actions[]`
- `NextBestAction.action_id`
- `NextBestAction.kind`
- `NextBestAction.api_hint`
- `NextBestAction.user_facing_prompt`
- `guidance.suggested_copy[]` when the best next step is a phrase the agent can say rather than an API call.

Back-prompts must be short, optional, and grounded in capabilities that exist now. Do not invent actions that the current API cannot perform.

Recommended action types for stream responses:

| User-facing intent | Current support | Recommended guidance action |
| --- | --- | --- |
| "Show me more like this" | Supported in feed sessions | `POST /api/v1/feed-sessions/{sessionID}/more` when a session exists |
| "Open/read this story" | Supported by card fields | Use card `short_url`, `canonical_url`, or `share_url`; no API call needed |
| "Tell me more about this topic" | Supported | `GET /api/v1/topics/{topicID}` when a stable topic ID/slug is available |
| "Search this topic" | Supported | `GET /api/v1/search?q=<topic>` |
| "Show other streams/topics" | Supported | `GET /api/v1/streams` and `GET /api/v1/streams/{slug}` |
| "Less of this category for this session" | Supported in feed sessions | `PATCH /api/v1/feed-sessions/{sessionID}/controls` with `block_topics` or narrower `interests` |
| "More of this category for this session" | Supported in feed sessions | `PATCH /api/v1/feed-sessions/{sessionID}/controls` with `pin_topics` or `interests` |
| "Remember my preference" | Partially supported | `PUT /api/v1/preferences/{personID}` only for supported preference fields |
| "Dismiss/close/react to this card" | Supported | `POST /api/v1/signals` with existing signal types |
| "Change the content itself" | Not supported | Do not promise mutation of source content; offer feedback/signals or a preference/control action instead |

Important distinction:

- Session controls are temporary and can support "less of this topic/category right now."
- Persistent preferences require authentication and should only be suggested for fields the server actually saves.
- The current preferences handler accepts `topics_avoided`, but implementation should verify whether it is persisted before advertising "remember less of X forever."

For plain `GET /api/v1/stream` responses that are not tied to a feed session, prefer discovery-oriented back-prompts:

- `open_short_url` for a returned card that has `short_url`.
- `view_related_topic` with `GET /api/v1/topics/{topicID}` when a topic can be safely mapped.
- `search_topic` with `GET /api/v1/search?q=<topic>`.
- `browse_stream` with `GET /api/v1/streams/{slug}` when a topic maps to a public stream slug.
- `start_feed_session` with `POST /api/v1/feed-sessions` for agents that want interactive controls.

For feed-session responses, prefer control-oriented back-prompts:

- `send_more`
- `less_of_topic`
- `more_of_topic`
- `switch_to_latest`
- `switch_to_top`
- `save_preferences` only when authenticated and persistence is implemented.

Example guidance:

```json
{
  "next_best_actions": [
    {
      "action_id": "search_topic",
      "title": "Search this topic",
      "why": "Find related cards without changing your feed preferences.",
      "kind": "discover",
      "priority": "medium",
      "available": true,
      "api_hint": {
        "method": "GET",
        "path": "/api/v1/search?q=tech-ai"
      },
      "user_facing_prompt": "Want more on this topic?"
    },
    {
      "action_id": "less_of_topic",
      "title": "Less of this topic",
      "why": "Session controls can suppress this topic in the next batch.",
      "kind": "personalize",
      "priority": "low",
      "available": true,
      "api_hint": {
        "method": "PATCH",
        "path": "/api/v1/feed-sessions/{sessionID}/controls"
      },
      "user_facing_prompt": "Want less of this topic in the next batch?"
    }
  ]
}
```

Do not include raw personal keys in `api_hint.path`.

## Credential Matrix

| Request | Result |
| --- | --- |
| No `Authorization`, no `key` | Anonymous |
| Valid Bearer, no `key` | Authenticated as Bearer identity |
| Invalid Bearer, no `key` | `401 unauthorized` |
| No Bearer, valid `key`, GET/HEAD | Authenticated as person token |
| No Bearer, invalid `key`, GET/HEAD | `401 unauthorized` |
| No Bearer, empty `key`, GET/HEAD | `401 unauthorized` |
| Any Bearer plus any `key` | `400 invalid_request` mixed credentials |
| `key` on non-GET/HEAD | `400 invalid_request` unsupported query-key method |

Query keys are for person tokens only. Do not accept curator tokens through `?key=` unless a later product/security decision explicitly allows it.

## Files To Change

Backend:

- `cmd/agentapi/main.go`
  - Add `GET /api/v1/stream` alias.
  - Keep existing `/api/v1/streams` and `/api/v1/streams/{slug}` routes.
- `internal/agentapi/authmw.go`
  - Add query-key parsing and credential matrix behavior.
  - Validate person tokens with the same database path as Bearer person tokens.
  - Add query-key response safety headers.
  - Redact logs.
- `internal/agentapi/context.go`
  - Add auth source to `AgentIdentity` if response guidance needs it.
- `internal/agentapi/envelope.go`
  - Add `AccountState` fields and supporting typed structs if needed.
  - Reuse existing `WriteError` and `ErrorEnvelope`; do not introduce a new error format.
- `internal/agentapi/guidance/engine.go`
  - Add anonymous upgrade guidance.
  - Add invalid query-key and mixed-credential guidance.
  - Add capability-grounded back-prompts for stream, search, topic, and feed-session responses.
- `internal/agentapi/handlers/streams.go`
  - Only change if the alias needs a shared helper, canonical stream marker, or stream-specific suggested actions.
- `internal/agentapi/handlers/feed_sessions.go`
  - Keep or refine existing `send_more` and `adjust_controls` suggestions.
  - Add explicit more/less topic guidance only where a session ID exists.
- `internal/agentapi/handlers/signals.go`
  - Use existing signal endpoint for dismiss/close/react guidance; do not create a separate "tweak content" endpoint unless needed.
- `internal/agentapi/wire/card.go`
  - Keep short/canonical/share URL fields available for open/read/share prompts.

Tests:

- Prefer package-local tests under `internal/agentapi`.
- Add route-level tests where current route wiring can be tested without a live database.
- If database-backed token tests are expensive, isolate token validation behind test fixtures already used in the repo.
- Add guidance tests for optional next actions without making action order brittle.

Docs and web discoverability:

- `docs/api.md`
- `cmd/web/main.go` only if serving `/api` or `/api/docs` from the web service.
- Static/template files only after finding current route and page patterns.
- Do not add a web-owned `/api/v1/...` route without updating `deploy/ec2/cloudflared.yml`; `/api/v1/*` currently routes to Agent API.

## Implementation Plan

### Phase 1: Baseline

Run:

```bash
git status --short --branch
go test ./internal/agentapi/... ./cmd/agentapi
AWS_PROFILE=wem go run ./observations/cmd/observe 2>&1
```

Capture current live behavior:

```bash
curl -i 'https://chowdahh.com/api/v1/streams/latest?limit=1'
curl -i 'https://chowdahh.com/api/v1/stream?limit=1'
curl -i 'https://chowdahh.com/api/v1/streams/latest?limit=1&key=bad'
```

Expected before implementation:

- `/api/v1/streams/latest?limit=1` returns 200.
- `/api/v1/stream?limit=1` returns 404.
- Invalid `?key=bad` is ignored or treated as anonymous.

### Phase 2: Add Singular Stream Alias

Tests first:

- `GET /api/v1/stream?limit=1` returns 200.
- Response uses the same successful envelope as `/api/v1/streams/latest?limit=1`.
- `/api/v1/streams`
- `/api/v1/streams/latest`
- `/api/v1/streams/{slug}`

Implementation:

- Add `r.Get("/stream", h.ListStream)` or a dedicated default-stream helper under the `/api/v1` route group.
- Prefer mapping to the existing `latest` stream behavior.
- Do not duplicate stream query logic.

Run:

```bash
go test ./cmd/agentapi ./internal/agentapi/...
```

### Phase 3: Add Query-Key Auth Tests

Add tests for the credential matrix:

- Anonymous GET succeeds.
- Valid Bearer still works.
- Invalid Bearer still returns 401.
- Valid query key on GET authenticates as `person_token`.
- Valid query key on HEAD authenticates as `person_token`.
- Invalid query key on GET returns 401.
- Empty query key on GET returns 401.
- Valid Bearer plus valid key returns 400.
- Valid Bearer plus invalid key returns 400.
- Invalid Bearer plus valid key returns 400.
- Query key on POST returns 400.
- Query key on PATCH returns 400.

Security assertions:

- Response body does not contain the supplied key.
- Log capture, if available, does not contain the supplied key.
- Query-key responses include `Cache-Control: no-store, private`.
- Query-key responses include `Referrer-Policy: no-referrer`.

### Phase 4: Implement Query-Key Auth

Implementation notes:

- Parse presence with `r.URL.Query()["key"]`, not only `Get`, so empty key is explicit.
- Trim surrounding spaces before validation.
- Reject mixed credentials before validating either one.
- Reject query keys on non-GET/HEAD before validating the token.
- Query keys only authenticate person tokens.
- Reuse `database.GetUserByPAT`.
- Do not accept `ch_cur_` through query key.
- Add `AuthSource` to identity only if needed for guidance.

Pseudo-flow:

```text
hasBearer := extractBearer(r) != ""
keyValues, hasQueryKey := r.URL.Query()["key"]

if hasQueryKey {
  setQueryKeySafetyHeaders(w)
}

if hasBearer && hasQueryKey:
  400 invalid_request

if hasQueryKey && method not GET/HEAD:
  400 invalid_request

if hasQueryKey:
  if empty or invalid:
    401 unauthorized
  else:
    identity = person_token with auth_source=query_key

else if hasBearer:
  existing Bearer behavior

else:
  existing anonymous behavior
```

Run:

```bash
go test ./internal/agentapi/...
```

### Phase 5: Add Guidance

Anonymous guidance:

- Keep current `auth_mode: "anonymous"`.
- Add upgrade fields with placeholder-only examples.
- Include both paste-key URL and recommended header.
- Say anonymous public access works without a key.

Invalid query-key guidance:

- Explain that the supplied key was invalid or expired.
- Tell user to remove `key` for anonymous access or paste a valid key.
- Do not include the supplied key.

Mixed-credential guidance:

- Explain that the request used both `Authorization` and `key`.
- Tell user to use one auth method.
- Recommend Authorization header for production.

Unsupported-method guidance:

- Explain that URL keys are only accepted for GET/HEAD onboarding.
- Tell clients to use the Authorization header for write requests.

Back-prompt guidance:

- Add optional `next_best_actions` for stream responses, but keep them capability-grounded.
- For `GET /api/v1/stream`, suggest discovery actions first:
  - open a returned card's `short_url` if present,
  - search a topic,
  - browse a related stream,
  - start a feed session for interactive controls.
- For feed-session responses, suggest control actions:
  - send more,
  - less of a topic,
  - more of a topic,
  - switch rank mode.
- Use `POST /api/v1/signals` for dismiss/close/react/open feedback.
- Only suggest persistent preferences when authenticated and after verifying the server persists that exact preference field.
- Do not suggest "edit this content" or "change the article" because current APIs do not mutate source content. Reframe as feedback, controls, or preferences.

Back-prompt tests:

- Stream guidance contains at least one safe discovery action.
- Actions with `api_hint` point to existing routes only.
- No action path contains `key=`.
- Feed-session guidance includes controls only when a session ID exists.
- Anonymous guidance does not suggest authenticated-only preference writes as immediately available.
- Persistent "less of topic" is not advertised unless `topics_avoided` persistence is implemented and tested.

Run:

```bash
go test ./internal/agentapi/...
```

### Phase 6: Add Canonical API Docs

Create or update `docs/api.md`.

Required content:

- `GET /api/v1/stream`
- `GET /api/v1/streams`
- `GET /api/v1/streams/{slug}`
- Anonymous example.
- Paste-key example with warning.
- Authorization header example as recommended production form.
- Invalid key behavior.
- Mixed credential behavior.
- Query keys are GET/HEAD-only.
- No real tokens.

Recommended copy:

```markdown
# Chowdahh API

GET JSON. Anonymous by default. Add a personal key for authenticated results.

Public stream:

`https://chowdahh.com/api/v1/stream?limit=10`

Paste-key onboarding:

`https://chowdahh.com/api/v1/stream?limit=10&key=PASTE_YOUR_KEY_HERE`

URL keys are convenient, but can appear in browser history, proxies, and logs.
For production apps, use:

`Authorization: Bearer PASTE_YOUR_KEY_HERE`
```

### Phase 7: Serve Human Docs

Preferred route:

- `GET /api`

Fallback:

- `GET /api/docs`
- Optionally redirect `/api` to `/api/docs`.

Important routing constraint:

- Do not add `/api/v1/...` docs routes to the web app unless `deploy/ec2/cloudflared.yml` is updated. `/api/v1/*` is Agent API traffic.

Tests:

- `/api` or `/api/docs` returns HTML 200.
- Body includes `/api/v1/stream`.
- Body includes `PASTE_YOUR_KEY_HERE`.
- Body includes `Authorization: Bearer PASTE_YOUR_KEY_HERE`.
- Body contains no token-like real secrets.

### Phase 8: Stream Page Discovery

Add a small API discovery affordance to `/stream`, but do not interrupt feed use.

Acceptable options:

- Small footer/link: "API"
- Account/API modal link.
- Compact feed card only if it does not displace core content.

Avoid making the stream page feel like a marketing page.

Required copy points:

- Public JSON URL.
- Paste-key URL with placeholder.
- Production header recommendation.
- Link to `/api` or `/api/docs`.

Do not display a user's full existing token unless the existing account/token UX already deliberately shows newly-created tokens once.

### Phase 9: Legacy `/stream?offset=0` JSON

Preserve current behavior.

If adding a canonical pointer, prefer a response header instead of changing the JSON body:

```text
Link: <https://chowdahh.com/api/v1/stream>; rel="canonical"; type="application/json"
```

Tests:

- `GET /stream?offset=0` with `Accept: application/json` still returns 200.
- Existing JSON fields still exist.
- HTML `/stream` still loads.

### Phase 10: Smoke Tests

Start local services using the repo standard flow, then run:

```bash
curl -fsS 'http://localhost:<AGENT_PORT>/api/v1/stream?limit=1' | jq '.data, .guidance, .meta'
curl -fsS 'http://localhost:<AGENT_PORT>/api/v1/streams/latest?limit=1' | jq '.data, .guidance, .meta'
curl -i 'http://localhost:<AGENT_PORT>/api/v1/stream?limit=1&key=bad'
curl -i -H 'Authorization: Bearer bad' 'http://localhost:<AGENT_PORT>/api/v1/stream?limit=1&key=bad'
curl -i -X POST 'http://localhost:<AGENT_PORT>/api/v1/signals?key=bad'
curl -fsS 'http://localhost:<WEB_PORT>/api' | grep '/api/v1/stream'
curl -fsS 'http://localhost:<WEB_PORT>/stream' | grep '/api/v1/stream'
```

Expected:

- Public stream returns 200 JSON.
- Existing plural route returns 200 JSON.
- Invalid key returns 401 JSON.
- Mixed credentials return 400 JSON.
- Query key on POST returns 400 JSON.
- Stream guidance includes safe next actions.
- Feed-session guidance includes more/control next actions when a session is used.
- Docs mention `/api/v1/stream`.
- Stream page discovery exists if implemented.

### Phase 11: Final Verification

Run:

```bash
go test ./...
AWS_PROFILE=wem go run ./observations/cmd/observe 2>&1
git status --short
```

Live checks after deploy:

```bash
curl -i 'https://chowdahh.com/api/v1/stream?limit=1'
curl -i 'https://chowdahh.com/api/v1/streams/latest?limit=1'
curl -i 'https://chowdahh.com/api/v1/stream?key=bad'
curl -i -H 'Authorization: Bearer bad' 'https://chowdahh.com/api/v1/stream?key=bad'
curl -i -X POST 'https://chowdahh.com/api/v1/signals?key=bad'
curl -i 'https://chowdahh.com/api'
curl -i 'https://chowdahh.com/stream'
```

Expected:

- New singular endpoint works.
- Existing plural endpoints still work.
- Invalid explicit query key returns 401 JSON.
- Mixed credentials return 400 JSON.
- Query key on non-GET/HEAD returns 400 JSON.
- Query-key responses have no-store and no-referrer headers.
- Anonymous responses include upgrade guidance.
- Guidance suggests safe next actions such as open short URL, related topic/search, streams, feed-session controls, or signals.
- Guidance does not suggest unsupported source-content mutation.
- Docs page is public.
- No response includes a real token.
- Logs do not include raw `key=` values.

## Backward Compatibility

Must keep working:

- `GET /api/v1/streams`
- `GET /api/v1/streams/{slug}`
- `GET /api/v1/streams/latest?limit=1`
- `GET /stream` HTML
- `GET /stream?offset=0` JSON with `Accept: application/json`
- Existing Bearer person-token auth
- Existing Bearer curator-token auth
- Existing anonymous access

Intentional changes:

- Explicit invalid `?key=` changes from anonymous fallback to `401`.
- Mixed `Authorization` plus `?key=` returns `400`.
- `?key=` on non-GET/HEAD returns `400`.

## Rollout Notes

Deploy Agent API changes first.

Then deploy docs/web discovery.

After deploy:

- Watch `kiteloop_agentapi_auth_*` metrics if available.
- Watch 400/401 rates on Agent API.
- Check logs for raw `key=` leakage.
- Run observer.
- Run live curl checks.

If leakage is found:

1. Disable query-key support behind a small code switch or deploy revert.
2. Keep `/api/v1/stream` alias and docs that recommend Bearer.
3. Re-enable query-key only after redaction is verified end to end.

## PR Summary Template

```markdown
## Summary
- Add `GET /api/v1/stream` as the canonical default stream alias
- Add GET/HEAD query-key onboarding for person tokens
- Reject invalid query keys with JSON 401
- Reject mixed credentials and query keys on write methods with JSON 400
- Add no-store/no-referrer safeguards for query-key requests
- Add anonymous upgrade guidance using current auth-mode schema
- Add capability-grounded next-best-action prompts for discovery, short links, topics, controls, and feedback signals
- Add public API docs and stream-page discovery

## Test Plan
- `go test ./...`
- `AWS_PROFILE=wem go run ./observations/cmd/observe 2>&1`
- `curl -i /api/v1/stream?limit=1`
- `curl -i /api/v1/stream?key=bad`
- `curl -i -H 'Authorization: Bearer bad' /api/v1/stream?key=bad`
- `curl -i -X POST /api/v1/signals?key=bad`
- `curl -i /api`
- `curl -i /stream`

## Security
- Query-string keys are onboarding-only
- Production docs recommend `Authorization: Bearer`
- Query-key responses set `Cache-Control: no-store, private`
- Query-key responses set `Referrer-Policy: no-referrer`
- Invalid query keys do not fall back to anonymous
- Mixed credentials are rejected
- Tests assert supplied keys are not echoed
```
