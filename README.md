# chowdahh_recipes

An AX-first companion repo for Chowdahh.

This repo defines the agent API specification for Chowdahh, including a JavaScript client, runnable examples, and a comprehensive test suite. **All endpoints in this spec are implemented and running in production at `chowdahh.com/api/v1/`.**

It is organized around agent intents:

- send content for a person
- send more
- inspect today’s topics and control chips
- replay card history and signals
- start Chowdahh Radio
- submit content
- send feedback

## Design Stance

- the primary abstraction is a `feed_session`, not a raw endpoint list
- controls and chips should mirror the product structure, but stay machine-usable
- replay means history of cards and signals, not audio playback
- `Chowdahh Radio` is its own clear capability
- toggles must be real and reflected in ranking or filtering
- local agent memory and Chowdahh profile state are separate on purpose
- feedback is broader than reporting a bad card

## Repo Map

```text
chowdahh_recipes/
├── agent.txt
├── docs/
│   ├── model.md
│   ├── principles.md
│   ├── api/
│   │   ├── consumption.md
│   │   ├── feedback.md
│   │   ├── radio.md
│   │   ├── submission.md
│   │   └── stats-and-replay.md
│   └── ax/
│       ├── agent-experience.md
│       └── preference-memory.md
├── examples/
├── install/zeroclaw/
├── openapi/
├── schemas/
├── skills/
│   ├── chowdahh_feedback/
│   ├── chowdahh_lookup/
│   ├── chowdahh_preferences/
│   └── chowdahh_submit/
└── src/
```

## Core Surfaces

- `POST /api/v1/feed-sessions`
  Start a feed session for a person. This is the default "send content now" call.
- `POST /api/v1/feed-sessions/{session_id}/more`
  Continue the session with more cards.
- `PATCH /api/v1/feed-sessions/{session_id}/controls`
  Apply or remove control chips.
- `GET /api/v1/streams`
  Discover available public streams — returns slugs, labels, and descriptions.
- `GET /api/v1/streams/{slug}`
  Browse a public stream.
- `GET /api/v1/replay`
  Show cards the person has already seen or signaled on.
- `GET /api/v1/stats/activity`
  Aggregate over replay and signals.
- `POST /api/v1/radio-sessions`
  Start Chowdahh Radio — returns tracks with audio URLs.
- `GET /api/v1/topics/{topic_id}`
  Deep drill-down: sources, timeline, curator, related topics, source URLs.
- `GET /api/v1/search`
  Search clusters by topic match. Anonymous results do not currently expose drill-down IDs.
- `PUT /api/v1/preferences/{person_id}`
  Sync durable Chowdahh preferences that an agent has confirmed with its person.
- `POST /api/v1/submissions/items`
  Submit a single story, poem, image, video, audio item, or structured object.
- `POST /api/v1/submissions/collections`
  Submit a corpus or knowledge bundle.
- `POST /api/v1/signals`
  Record `seen`, `open`, `save`, `share`, `dismiss`, and related card actions.
- `POST /api/v1/feedback`
  Send content requests, bug reports, feature requests, or quality reports.
- `GET /audio/{track_id}`
  Stream MP3 audio for a radio track (returned in radio session responses).

The full narrative contract is in [docs/api/contract-spec.md](/Users/splash/chowdahh_recipes/docs/api/contract-spec.md).
The OpenAPI spec is a planned work item; the narrative docs are the current canonical reference.

## Cards, Images, and Audio

Every card returned by the API includes an `image_url` when a hero image is available. Image URLs point to the original source or a CDN-cached version — no additional endpoint is needed to fetch them.

Radio tracks include both `audio_url` (streams MP3 via `GET /audio/{id}`) and `image_url`.

## Response Envelope

Successful responses use `{data, guidance, meta}`. Error responses use `{error, meta}` with optional `guidance`. The `guidance.next_best_actions` field is present on many responses but is optional — agents should always be able to fall back to the documented endpoints:

```json
{
  "data": { "session_id": "abc-123", "items": [...], "count": 6 },
  "guidance": {
    "status_explanation": "Feed session started with 6 cards.",
    "next_best_actions": [
      { "action_id": "send_more", "title": "Send more cards", "api_hint": { "method": "POST", "path": "/api/v1/feed-sessions/abc-123/more" } }
    ],
    "account_state": { "auth_mode": "anonymous", "rate_limit": { "limit": 30, "remaining": 28 } }
  },
  "meta": { "request_id": "f3704a2e6d2f0f82" }
}
```

The `guidance` block helps agents understand what happened and suggests what to do next — without needing to hardcode flow logic.

## First Calls

```bash
# Start a feed session
curl -X POST https://chowdahh.com/api/v1/feed-sessions \
  -H 'content-type: application/json' \
  -d '{"intent": "browse", "budget_minutes": 5, "include_controls": true}'

# Browse a stream
curl 'https://chowdahh.com/api/v1/streams/science?limit=5'

# Search
curl 'https://chowdahh.com/api/v1/search?q=NASA&limit=5'

# Replay (requires person token)
curl 'https://chowdahh.com/api/v1/replay?period=this_month' \
  -H 'Authorization: Bearer ch_person_xxx'
```

## Running Tests

```bash
npm test                    # test against default (chowdahh.com)
npm run test:prod           # explicit production test
CHOWDAHH_BASE_URL=http://localhost:8081 npm test  # test locally
```

## Agent Assets

- [agent.txt](/Users/splash/chowdahh_recipes/agent.txt) is the general agent entrypoint.
- [docs/model.md](/Users/splash/chowdahh_recipes/docs/model.md) defines the core nouns.
- [docs/ax/agent-experience.md](/Users/splash/chowdahh_recipes/docs/ax/agent-experience.md) defines the intended AX.
- [docs/ax/preference-memory.md](/Users/splash/chowdahh_recipes/docs/ax/preference-memory.md) explains what belongs in local memory vs Chowdahh.
- [skills/chowdahh_lookup/SKILL.md](/Users/splash/chowdahh_recipes/skills/chowdahh_lookup/SKILL.md) is the feed/replay lookup skill.

## ZeroClaw / OpenClaw

Install instructions live in [install/zeroclaw/README.md](/Users/splash/chowdahh_recipes/install/zeroclaw/README.md). The short version is:

```bash
git clone https://github.com/splashkes/chowdahh_recipes.git
zeroclaw skills install /path/to/chowdahh_recipes/skills/chowdahh_lookup
zeroclaw skills install /path/to/chowdahh_recipes/skills/chowdahh_submit
zeroclaw skills install /path/to/chowdahh_recipes/skills/chowdahh_preferences
zeroclaw skills install /path/to/chowdahh_recipes/skills/chowdahh_feedback
```

## Minimal JS Client

The repo includes a small fetch client in [src/client.js](/Users/splash/chowdahh_recipes/src/client.js) and runnable examples in [examples](/Users/splash/chowdahh_recipes/examples).

## Critical Review

The explicit critique and next-pass recommendations are in [suggestions_for_improvement.md](/Users/splash/chowdahh_recipes/suggestions_for_improvement.md).
