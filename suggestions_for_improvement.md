# suggestions_for_improvement

## Status (2026-04-12)

The full agent API was implemented and deployed to production on 2026-04-12. All 20+ endpoints are live, tested, and passing the SDK test suite (18/18 tests pass).

## Addressed in Implementation

1. **Auth** — 3-tier auth implemented (anonymous/person_token/curator_token) with per-token rate limiting. Delegation headers supported. Person tokens use `ch_person_*` prefix.

2. **Feed session lifecycle** — Sessions stored in Redis with 12h TTL, shared between agent API and web. Position tracking, control persistence, and resume all work. Sessions expire naturally via TTL.

3. **Response envelope** — Every endpoint uses `{data, guidance, meta}` envelope with structured error codes, rate limit state, and contextual next-best-actions.

4. **AX guidance engine** — Behavioral pattern detection (send-more counting) drives escalating suggestions. Rate limit, auth recovery, and empty-state guidance all implemented.

5. **Feedback** — Unified feedback endpoint supporting content_request, bug_report, feature_request, quality_report with contextual follow-up suggestions.

6. **Submissions** — Wraps existing intake evaluator. Single items and batch collections both work. Curator tokens get higher quotas.

## Still Open

1. **Control ontology** — Which chips are editorial vs algorithmic vs person-specific is not yet formally defined. The current implementation uses interest groups (7 categories) and sort mode as the control surface.

2. **Replay event semantics** — `seen` still needs a tighter definition. Current implementation records `seen` signals from agents but doesn't validate that delivery actually happened.

3. **Radio queue semantics** — Radio sessions now return rich track objects with `audio_url` per track. Topic-specific queues and queue progress persistence across disconnects still need work.

4. **Submission lifecycle** — Large uploads, ownership review, and object lifecycle remain open. The current implementation wraps the intake evaluator for URL-based submissions only.

5. **Idempotency-Key** — Redis infrastructure exists (`SetIdempotencyResult`/`GetIdempotencyResult`) but is not yet wired into POST handlers.

6. **Cursor-based pagination** — Currently uses offset-encoded cursors. True keyset pagination for performance at scale is future work.

7. **Preferences depth** — Only `topics_followed` (mapped to interest slugs) is persisted. Tone preferences, delivery defaults, and source preferences are accepted but not yet stored durably beyond interests.

8. **OpenAPI spec** — The `openapi/chowdahh-agent-v1.yaml` draft needs updating to match the implemented endpoints and response shapes.

## Addressed in Triage Implementation (2026-04-13)

7. **Staleness triage** — LLM-powered supersession detection runs every 3 hours as pipeline Stage 3.7. Evaluates active clusters against topic fact timelines and sibling clusters. Superseded clusters are suppressed with `superseded_by` annotation; stale clusters with no successor are dissolved for re-clustering. Permalinks follow supersession chains.

## Suggested Next Pass

1. Wire idempotency-key handling into all POST mutation handlers.
2. Expand preferences storage to persist tone, delivery, and source preferences in a dedicated table.
3. Add topic-specific radio queues (audio URLs are now integrated).
4. Define the control ontology formally and add per-chip confidence to the control state.
5. Update the OpenAPI spec to match production reality.
6. Add one worked end-to-end flow per major intent (partially done via examples/).

## TUI Client Roadmap

### 1.0 — Paste-token auth + feed reader

- **Auth**: prompt for `ch_person_*` token, cache at `~/.config/chowdahh/token`, `logout` command to clear
- **Streams**: browse top/science/world/business/culture, paginate with cursor
- **Card view**: headline, summary, topics, source count, share link (`short_url`) prominently displayed
- **Signals**: record seen/open/save/dismiss as user navigates
- **Replay**: view card history
- **Preferences**: view and edit followed topics
- **Share links**: `short_url`/`share_url` shown on every card, easy copy-to-clipboard
- **Feedback**: submit content requests and bug reports

### 1.1 — Radio player

- **Radio sessions**: start headlines/briefing/topic_run modes
- **Audio playback**: stream MP3 from `/audio/{track_id}` via system player or inline
- **Now-playing display**: current track headline, topics, progress, share link
- **Controls**: skip, pause, resume, stop
- **Queue view**: upcoming tracks with headlines

### 2.0 — Browser-based login (requires main app changes)

- **Device auth flow**: TUI generates one-time code, opens `chowdahh.com/auth/device?code=XXX`
- **User approves in browser**, server creates and delivers `ch_person_*` token via polling
- **New endpoint**: `POST /api/v1/auth/device` + `GET /api/v1/auth/device/{code}/poll`
- **Token auto-cached** same as 1.0 from there
