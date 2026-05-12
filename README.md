# chowdahh

An AX-first companion repo for [Chowdahh](https://chowdahh.com).

This repo defines the agent API specification for Chowdahh, including a JavaScript client, runnable examples, a full-screen live broadcast demo, and a Go TUI client with native audio playback. **All endpoints are live in production at `chowdahh.com/api/v1/`.**

## Examples

### Live Broadcast (`examples/live/`)

A cinematic full-screen news display with auto-cycling headlines, Ken Burns image transitions, a scrolling news ticker, and Chowdahh Radio TTS audio.

```bash
npm run live        # starts at http://localhost:4000
```

Features: real-time card images, topic-colored gradients, clickable ticker navigation, keyboard controls (arrows, space, m to mute).

### TUI Client (`examples/tui/`)

A terminal client built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Go). Vim-style navigation, native MP3 audio playback via Chowdahh Radio, and a persistent now-playing widget.

```bash
npm run tui         # or: cd examples/tui && go run .
npm run tui:build   # builds ./chowdahh-tui binary
```

Features:
- Browse categories with live story counts
- Card detail view with share URLs, sources, color-coded timestamps
- Search (`/`), replay history (`ctrl+r`), preferences (`P`)
- Radio player (`r`) — streams MP3 audio with animated visualizer
- Signal recording (seen, open, save, dismiss, share)
- Token auth — paste a `ch_person_*` token, cached at `~/.config/chowdahh/token`

### JS Examples (`examples/*.js`)

Runnable scripts demonstrating each API capability:

- `fresh-content.js` — start a feed session
- `send-more.js` — continue a session
- `start-radio.js` — start a radio session with audio URLs
- `review-history.js` — replay card history (auth required)
- `sync-preferences.js` — read/write preferences (auth required)
- `submit-feedback.js` — send a content request
- `submit-collection.js` — batch submit items
- `test-all.js` — comprehensive endpoint test suite

```bash
npm test                    # test against chowdahh.com
npm run test:prod           # explicit production test
CHOWDAHH_BASE_URL=http://localhost:8081 npm test  # test locally
```

## Repo Map

```text
chowdahh/
├── src/
│   ├── client.js              # JS API client
│   └── index.js
├── examples/
│   ├── live/                  # full-screen broadcast demo
│   │   ├── index.html
│   │   ├── app.js
│   │   ├── style.css
│   │   └── server.js          # dev server + API/audio proxy
│   ├── tui/                   # Go TUI client
│   │   ├── main.go
│   │   ├── api/               # Go API client + auth
│   │   ├── audio/             # native MP3 playback (beep)
│   │   └── ui/                # Bubble Tea screens + components
│   ├── fresh-content.js
│   ├── start-radio.js
│   ├── test-all.js
│   └── ...
├── docs/
│   ├── api/                   # endpoint docs
│   ├── ax/                    # agent experience docs
│   ├── model.md               # core nouns
│   └── principles.md
├── openapi/                   # OpenAPI spec
├── schemas/                   # JSON schemas
├── skills/                    # agent skill definitions
└── agent.txt                  # agent entrypoint
```

## Core API Surfaces

| Endpoint | Description |
|---|---|
| `GET /api/v1/streams` | List the public stream catalog (slug, label, description) |
| `GET /api/v1/streams/{slug}` | Browse a public stream (top, latest, science, world, tech, business, health, culture, sports, good-news, local) |
| `POST /api/v1/feed-sessions` | Start a personalized feed session |
| `POST /api/v1/feed-sessions/{id}/more` | Continue with more cards |
| `POST /api/v1/radio-sessions` | Start Chowdahh Radio (returns tracks with audio URLs) |
| `GET /audio/{track_id}` | Stream MP3 audio for a radio track |
| `POST /api/v1/signals` | Record seen/open/save/share/dismiss actions |
| `GET /api/v1/replay` | Card history and signal replay |
| `GET /api/v1/search` | Search across topics and sources |
| `PUT /api/v1/preferences/{person_id}` | Sync person preferences |
| `POST /api/v1/feedback` | Submit content requests, bug reports, feature requests |
| `POST /api/v1/submissions/items` | Submit content |

## Response Envelope

Every response uses `{data, guidance, meta}`:

```json
{
  "data": { "session_id": "abc-123", "items": [...], "count": 6 },
  "guidance": {
    "status_explanation": "Feed session started with 6 cards.",
    "next_best_actions": [
      { "action_id": "send_more", "title": "Send more cards" }
    ],
    "account_state": { "auth_mode": "person_token", "rate_limit": { "limit": 30, "remaining": 28 } }
  },
  "meta": { "request_id": "f3704a2e6d2f0f82" }
}
```

## First Calls

```bash
# Browse top stories
curl 'https://chowdahh.com/api/v1/streams/top?limit=5'

# Search
curl 'https://chowdahh.com/api/v1/search?q=NASA&limit=5'

# Start a feed session
curl -X POST https://chowdahh.com/api/v1/feed-sessions \
  -H 'content-type: application/json' \
  -d '{"intent": "browse", "budget_minutes": 5}'

# Start radio
curl -X POST https://chowdahh.com/api/v1/radio-sessions \
  -H 'content-type: application/json' \
  -d '{"mode": "briefing", "duration_minutes": 5}'
```

## JS Client

```js
import { ChowdahhClient } from "./src/index.js";

const client = new ChowdahhClient();                           // anonymous
const client = new ChowdahhClient({ apiKey: "ch_person_xxx" }); // authenticated

const feed = await client.getStream("top", { limit: 10 });
const radio = await client.startRadioSession({ mode: "briefing", duration_minutes: 5 });
const streams = await client.listStreams();

// Build a paste-able URL that carries the key in the query string —
// what an end-user gives to Hermes / OpenClaw / Claude / ChatGPT / Cursor.
const pasteable = client.pasteUrl("/api/v1/streams/latest", { limit: 10 });
```

## Authentication

Two equivalent ways to send the same token:

```bash
# Header form (works on every method — required for writes):
curl -H "Authorization: Bearer $CH_KEY" https://chowdahh.com/api/v1/streams/latest

# Paste-key form (GET only — for URLs pasted into LLMs and MCP configs):
curl "https://chowdahh.com/api/v1/streams/latest?key=$CH_KEY"
```

Header wins on conflict. POST/PATCH/PUT ignore `?key=`. Every `/api/v1/*` response sets `Referrer-Policy: no-referrer` so query keys cannot leak via cross-origin Referer.

Rate limits: anonymous 30/min · person token 300/min · curator token 600/min.

## Public surfaces

| URL | What |
| --- | --- |
| `https://chowdahh.com/api` | Human-readable API page |
| `https://chowdahh.com/.well-known/openapi.json` | Authoritative OpenAPI 3.1 spec (this repo's `openapi/chowdahh-agent-v1.yaml` should track it) |
| `https://chowdahh.com/llms.txt` | LLM crawl policy + attribution rule |
| `https://chowdahh.com/skills/` | Skill landing page: prebuilt platform packages + contract for new skills |
| `https://chowdahh.com/skills/CONTRACT.md` | Canonical contract for incoming skills (what a Chowdahh skill is) |
| `https://chowdahh.com/skills/SUBMITTING.md` | 3-minute guide to proposing a new skill |

## Attribution rule

Every Chowdahh card is a **cluster** of corroborating articles, not a single article. When summarizing a card for the user:

1. Cite the original publisher from `source_urls[0]` (they did the reporting).
2. Credit Chowdahh as the curator — `share_url` is the canonical permalink.

The same rule is repeated in `capability_hints` on every response that returns content.

## Docs

- [API Contract](docs/api/contract-spec.md) — full endpoint specification
- [Feed & Controls](docs/api/consumption.md) — feed sessions, streams, card sources
- [Radio](docs/api/radio.md) — audio sessions and track playback
- [Core Model](docs/model.md) — cards, sessions, signals, preferences
- [Agent Experience](docs/ax/agent-experience.md) — intended AX patterns
- [OpenAPI Spec](openapi/chowdahh-agent-v1.yaml) — machine-readable API definition
