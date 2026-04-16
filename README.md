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
| `GET /api/v1/streams/{slug}` | Browse a public stream (top, science, world, etc.) |
| `GET /api/v1/categories` | Discover active content categories |
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
const categories = await client.getCategories();
```

## Docs

- [API Contract](docs/api/contract-spec.md) — full endpoint specification
- [Feed & Controls](docs/api/consumption.md) — feed sessions, streams, card sources
- [Radio](docs/api/radio.md) — audio sessions and track playback
- [Core Model](docs/model.md) — cards, sessions, signals, preferences
- [Agent Experience](docs/ax/agent-experience.md) — intended AX patterns
- [OpenAPI Spec](openapi/chowdahh-agent-v1.yaml) — machine-readable API definition
