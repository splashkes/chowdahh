# Full Contract Specification

This document defines the intended full API contract for Chowdahh's agent-facing surface.

It is a product contract first, an implementation contract second.

The machine-readable draft lives in [openapi/chowdahh-agent-v1.yaml](/Users/splash/chowdahh_recipes/openapi/chowdahh-agent-v1.yaml). This document explains the behavior, boundaries, and semantics around that surface.

## 1. Goals

The API should let an agent:

- send content for a person
- send more inside the same active session
- inspect and apply the product control surface structurally
- replay what the person has already seen or acted on
- start and manage Chowdahh Radio separately from replay/history
- drill into topics, sources, curators, and searches
- submit content with explicit transformation rules
- send feedback that may be a content request, bug report, feature request, or quality report

## 2. Base Contract

- Base URL: `https://chowdahh.com`
- Version prefix: `/api/v1`
- Default media type:
  `application/json`
- Character encoding:
  `utf-8`

### Stability rule

- The `/api/v1` surface is additive by default.
- Breaking changes require a new version prefix or a published deprecation window.
- New enum values may appear in non-critical descriptive fields; agents should ignore unknown values unless the field is documented as strict.

## 3. Authentication And Delegation

The API must distinguish:

- the `agent` making the call
- the `person` the agent is serving
- the `curator` identity of a submission pipeline, where relevant

### 3.1 Auth modes

#### Anonymous agent

Used for public reads and low-friction submission.

- no bearer token
- strict rate limits
- limited write capabilities

#### Person token

A user-scoped token that authorizes an agent to act for a specific person.

Bearer format example:

```text
Authorization: Bearer ch_person_...
```

Capabilities should include:

- feed sessions
- replay/history
- preferences
- feedback
- personal submissions
- radio

#### Curator token

A system or partner token for higher-volume or domain-specific contribution workflows.

Bearer format example:

```text
Authorization: Bearer ch_cur_...
```

Capabilities should include:

- richer submission quotas
- curator attribution
- higher ingestion ceilings

### 3.2 Delegation headers

Every authenticated request acting on behalf of a person should support:

- `X-Chowdahh-Agent-Id`
- `X-Chowdahh-Agent-Name`
- `X-Chowdahh-Acting-For`

These fields are for audit and attribution. They do not replace bearer auth.

### 3.3 Authorization rule

- The authenticated principal must be allowed to act for the `person_id` in the request.
- The server should reject cross-person access with `403 forbidden`.

## 4. Common Request Rules

### 4.1 Idempotency

**Status: planned.** Redis infrastructure exists but is not yet wired into POST handlers.

All mutating `POST` endpoints that create sessions, submissions, or feedback should accept:

```text
Idempotency-Key: <opaque-client-generated-key>
```

The server should:

- treat repeated keys with the same effective payload as safe retries
- reject conflicting payload reuse with `409 conflict`

### 4.2 Request IDs

The server should return:

```text
X-Request-Id: <request-id>
```

This should be stable in logs and error responses.

### 4.3 Pagination

Cursor-based pagination should be used wherever ordered collections may continue.

Common fields:

- `next_cursor`
- `has_more`

Clients should treat cursors as opaque.

### 4.4 Time

All timestamps are RFC 3339 UTC unless otherwise stated.

### 4.5 Locale

Optional client hints:

- `Accept-Language`
- `X-Chowdahh-Timezone`

These help tune delivery and text, but should not be required.

## 5. Common Response Shape

Successful responses use `{data, guidance, meta}`. The `guidance` block is present on most successful responses but is optional — `guidance.next_best_actions` may be absent:

```json
{
  "data": {},
  "guidance": {
    "status_explanation": "Human-readable explanation of what happened.",
    "next_best_actions": [
      { "action_id": "send_more", "title": "Send more cards", "priority": "primary", "api_hint": { "method": "POST", "path": "/api/v1/feed-sessions/{id}/more" } }
    ],
    "account_state": {
      "auth_mode": "anonymous",
      "rate_limit": { "limit": 30, "remaining": 28, "reset_at": "2026-04-12T12:00:00Z" }
    }
  },
  "meta": {
    "request_id": "f3704a2e6d2f0f82"
  }
}
```

The `guidance` block helps agents understand what happened and what to do next without needing to hardcode flow logic. Clients should treat `next_best_actions` as optional and keep using the documented `/api/v1` endpoints as the source of truth.

## 6. Error Contract

Errors should be machine-usable and human-restatable.

Error responses use `{error, meta}` and may include a `guidance` block when contextual recovery actions are available. Clients should not assume `guidance` is present on all error responses.

```json
{
  "error": {
    "code": "invalid_control",
    "message": "The selected control option is not available in this feed session.",
    "details": {
      "session_id": "feed_sess_123",
      "control_slug": "good-news"
    }
  },
  "guidance": {
    "status_explanation": "The control 'good-news' is not available in this session.",
    "next_best_actions": [
      { "action_id": "list_controls", "title": "See available controls" }
    ]
  },
  "meta": {
    "request_id": "f3704a2e6d2f0f82"
  }
}
```

### 6.1 Standard error codes

- `invalid_request`
- `unauthorized`
- `forbidden`
- `not_found`
- `conflict`
- `rate_limited`
- `validation_error`
- `expired_session`
- `invalid_control`
- `processing`
- `service_unavailable`

### 6.2 Status code guidance

- `400` malformed or invalid request
- `401` missing or invalid auth
- `403` valid auth, insufficient permission
- `404` resource not found
- `409` state conflict or idempotency mismatch
- `422` semantically invalid payload
- `429` rate limited
- `503` temporary service issue

## 7. Core Objects

## 7.1 Person

Represents the human receiving Chowdahh service.

Key fields:

- `person_id`
- `preferences`
- `profile_state_version`

## 7.2 Feed Session

Represents an active delivery session for a person.

Key fields:

- `session_id`
- `person_id`
- `state`
- `items`
- `controls`
- `applied_preferences`
- `why_this_set`
- `expires_at`

States:

- `active`
- `complete`
- `expired`

### Feed session rule

`send more` should happen inside the same session whenever the intent is continuous browsing.

## 7.3 Card (Feed Item)

Represents the atomic surfaced object.

Key fields:

- `id`
- `headline`
- `summary`
- `image_url` — hero image URL when available (original source or CDN)
- `topics`
- `source_count`
- `significance_score`
- `velocity`
- `short_url`
- `canonical_url`
- `share_url`

## 7.4 Control State

Represents the machine-usable version of the top-of-product controls.

Key fields:

- `groups[]`
- each group has `selection_mode`
- each option has `slug`, `label`, `count`, `selected`, `confidence`

### Control groups

Suggested initial groups:

- `sort`
- `topics`
- `tones`
- `places`

## 7.5 Replay Event

Represents a historical event tied to a card.

Key fields:

- `event_id`
- `person_id`
- `card_id`
- `signal_type`
- `topic_id`
- `headline`
- `occurred_at`
- `session_id`

### `seen` definition

This must be defined tightly in implementation.

Recommended draft:

- A card becomes `seen` when the system has enough confidence that it was actually delivered to the person in a meaningful way.
- Mere server inclusion in a response is not enough by itself unless the delivery surface guarantees visibility.

## 7.6 Radio Session

Represents audio delivery mode.

Key fields:

- `radio_session_id`
- `state`
- `queue_length`
- `tracks[]` — each track has `id`, `headline`, `audio_url`, `topics`, `source_count`

States:

- `ready`
- `playing`
- `paused`
- `ended`

## 7.7 Submission

Represents content brought into Chowdahh.

Key fields:

- `submission_id`
- `person_id`
- `status`
- `submission_kind`
- `transformation_policy`
- `topic_id` or `library_id`
- `estimated_ready_at`

Statuses:

- `queued`
- `processing`
- `ready`
- `failed`

## 7.8 Feedback

Represents a person-directed request back into the system.

Types:

- `content_request`
- `bug_report`
- `feature_request`
- `quality_report`

## 8. Endpoint Contract

## 8.1 Feed Sessions

### `POST /api/v1/feed-sessions`

Start a new feed session.

Required:

- `intent` (e.g. `browse`)

Optional:

- `budget_minutes`
- `include_controls`

The person is identified by the bearer token, not a request body field.

Behavior:

- creates a resumable delivery session (stored in Redis, 12h TTL)
- returns `items[]` plus the available control state
- applies durable preferences automatically if they exist for the authenticated person

### `GET /api/v1/feed-sessions/{session_id}`

Fetch session state, including position, controls, and card list. Useful for resuming or debugging.

### `POST /api/v1/feed-sessions/{session_id}/more`

Continue the same session. Parameters (`limit`) are sent as query params.

Behavior:

- returns the next increment of cards with a `position` counter
- preserves previously applied controls
- does not silently create a fresh unrelated session

### `PATCH /api/v1/feed-sessions/{session_id}/controls`

Apply or remove controls.

Behavior:

- validates requested controls against current available options
- returns the updated session view
- may change ranking, filtering, or both

## 8.2 Public Lanes And Discovery

### `GET /api/v1/streams`

Returns the list of available public stream slugs with labels and descriptions. Clients should call this endpoint to discover streams rather than hardcoding slug lists. The current default set includes: `top`, `latest`, `science`, `world`, `tech`, `business`, `health`, `culture`, `sports`, `good-news`, `local`.

### `GET /api/v1/streams/{stream_slug}`

Public or shared lane access for high-volume browsing.

### `GET /api/v1/search`

Searches clusters by topic match. Results are card objects with `id`, `headline`, `summary`, etc.

Note: anonymous search results do not currently expose stable result types or drill-down IDs. Treat search primarily as a browse surface unless the response includes an explicit identifier you can use safely.

## 8.3 Topic And Curator Drill-Down

These endpoints require true internal identifiers. Anonymous search results do not reliably expose those identifiers, so clients should not assume a direct anonymous search → drill-down flow.

### `GET /api/v1/topics/{topic_id}`

Returns:

- topic summary
- timeline
- sources
- related topics
- canonical URL

### `GET /api/v1/curators/{curator_id}`

Returns:

- curator identity
- specialties
- top topics
- explanatory metadata

## 8.4 Replay And Stats

### `GET /api/v1/replay`

Ordered event history for the authenticated person. Requires a person token.

Useful filters:

- `signal_type`
- `period`
- `cursor`
- `limit`

### `GET /api/v1/stats/activity`

Aggregate view over replay/signals.

Useful filters:

- `signal_type`
- `period`
- `group_by`

## 8.5 Radio

Radio is session-based. Start a session, and the server builds a queue of tracks from current content. Each track includes an `audio_url` that streams MP3 audio. Control the session via PATCH.

### `POST /api/v1/radio-sessions`

Start a radio session.

Optional:

- `mode` — `headlines`, `briefing`, or `topic_run`
- `duration_minutes`
- `topic_lenses`

Returns `data.radio_session_id`, `data.state`, `data.queue_length`, and `data.tracks[]` with `audio_url` per track.

### `GET /api/v1/radio-sessions/{radio_session_id}`

Fetch playback state, position, and remaining tracks with audio URLs.

### `PATCH /api/v1/radio-sessions/{radio_session_id}`

Control an active radio session. Send `{"action": "<action>"}`:

- `pause`
- `resume`
- `skip`
- `stop`

Returns updated state and remaining tracks.

Session states: `ready` → `playing` → `paused` / `ended`.

### `GET /audio/{track_id}`

Fetch MP3 audio for a single track. Returns `Content-Type: audio/mpeg`. Audio is synthesized on first request and cached.

## 8.6 Preferences

### `GET /api/v1/preferences/{person_id}`

Fetch durable stored preferences. Requires a person token matching the `person_id`.

### `PUT /api/v1/preferences/{person_id}`

Replace or merge durable preferences. Requires a person token matching the `person_id`.

Behavior:

- must only reflect durable Chowdahh state
- should not absorb private local-memory hints unless explicitly intended

Production note: only `topics_followed` (mapped to interest slugs) is currently persisted. Tone, delivery, and source preferences are accepted but not yet stored durably beyond interests.

## 8.7 Submission

### `POST /api/v1/submissions/items`

Submit one item. Minimal required fields: `title` and `source_url`.

### `POST /api/v1/submissions/collections`

Submit a batch of items as an array. Each item needs `title` and `source_url`.

Note: collection submissions may succeed at the request level (201) while skipping individual items. Inspect `accepted`, `results[]`, and per-item statuses before treating the submission as complete.

### `GET /api/v1/submissions/{submission_id}`

Fetch state and output bindings.

## 8.8 Signals

### `POST /api/v1/signals`

Record one or more interaction events as a batch array. Each signal needs `signal_type` and `card_id`.

First-party signal types:

- `seen`
- `open`
- `save`
- `share`
- `dismiss`
- `source_open`

Note: invalid or unrecognized signal types are silently skipped. Callers should inspect `recorded` in the response body — a `200` status does not mean all signals were accepted.

## 8.9 Feedback

### `POST /api/v1/feedback`

Send product- or content-directed feedback. Required fields: `feedback_type` and `title`.

Use cases:

- request more content
- report broken behavior
- request a feature
- report bad quality

Note: feedback validation failures return `{error, meta}` without a `guidance` block.

## 9. Submission Transformation Policy

Every submission should state transformation rules.

### `synthesis_mode`

- `preserve`
- `light_synthesis`
- `full_synthesis`

### `voice_preservation`

- `preserve_verbatim`
- `normalize_lightly`
- `allow_rewrite`

### `media_policy`

- `preserve_embeds`
- `derive_previews`
- `allow_transcodes`

## 10. Rate Limiting

The server should return:

- `429 Too Many Requests`
- `Retry-After`

Rate limits may vary by auth mode:

- anonymous agent
- person token
- curator token

## 11. Auditability

Every mutating request should be attributable to:

- authenticated principal
- acting person
- agent headers when present
- request ID
- idempotency key when present

## 12. Worked Flows

## 12.1 Send content

1. `POST /api/v1/feed-sessions` with `intent: "browse"`
2. receive `{data, guidance, meta}` envelope with cards + controls
3. present cards, use `guidance.next_best_actions` for suggested follow-ups
4. record `seen`/`open`/`save`/`share` via `POST /api/v1/signals`

## 12.2 Send more

1. `POST /api/v1/feed-sessions/{session_id}/more`
2. receive more cards under the same control state, with position tracking

## 12.3 Adjust controls

1. `PATCH /api/v1/feed-sessions/{session_id}/controls`
2. receive updated session view with updated control selections

## 12.4 Replay history

1. `GET /api/v1/replay?period=this_month` (requires person token)
2. optionally `GET /api/v1/stats/activity`

## 12.5 Start radio

1. `POST /api/v1/radio-sessions` with `mode` and `duration_minutes`
2. receive `data.state: "ready"`, `data.queue_length`, and `guidance.next_best_actions`
3. `PATCH /api/v1/radio-sessions/{id}` with `{"action": "resume"}` to start playback
4. `PATCH` with `skip`, `pause`, or `stop` to control the session

## 12.6 Submit content

1. confirm source with the person
2. `POST /api/v1/submissions/items` or `POST /api/v1/submissions/collections`
3. `GET /api/v1/submissions/{submission_id}` until `ready`

## 12.7 Send feedback

1. classify as content request, bug report, feature request, or quality report
2. `POST /api/v1/feedback`

## 13. Content Lifecycle

### 13.1 Time-Based Decay

Content naturally decays out of the stream over time. The ranking formula applies an evergreen-aware decay — breaking news fades in hours, reference content persists for days. See [ranking.md](ranking.md) for the formula.

### 13.2 Staleness Triage (LLM-Powered)

Every 3 hours, an LLM evaluates active clusters to detect **supersession** — when a story has been overtaken by newer developments. The triage checks each cluster's headline against recent facts on its topics and other active clusters on the same topics.

Three outcomes:

- **active** — no action, cluster stays in the stream
- **superseded** — cluster is suppressed and annotated with `superseded_by` pointing to the successor. Permalinks follow the chain to the current story.
- **dissolve** — cluster is dissolved and its member articles are freed for re-clustering. Used when the story is stale but no successor cluster exists yet.

Agents do not need to interact with this system. Superseded content is automatically filtered from stream responses. Permalinks to superseded clusters resolve to the current successor.

## 14. Resolved Questions

The following questions from the original spec have been resolved in production:

- **Envelope**: Successful responses use `{data, guidance, meta}`; error responses use `{error, meta}` with optional `guidance`.
- **GET endpoints**: `GET /api/v1/feed-sessions/{id}` and `GET /api/v1/preferences/{person_id}` are included.
- **Batch signals**: Signal writes are batch-only (array of events).
- **Radio controls**: Radio sessions have independent controls from feed sessions.
- **Preferences boundary**: Only `topics_followed` is durably persisted as interest slugs. Tone, delivery, and source preferences are accepted but not yet stored beyond interests.
