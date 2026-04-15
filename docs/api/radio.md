# Chowdahh Radio

## Goal

Give `Chowdahh Radio` a clear, separate machine surface.

Radio is not replay. Replay is prior card history. Radio is audio delivery mode.

## Key Concept

Radio is **session-based**. Start a session, and the server builds a queue of tracks from current content. Each track includes an `audio_url` that streams MP3 audio.

1. Start a radio session — get back a `tracks[]` array with audio URLs.
2. Play audio by fetching each track's `audio_url` (returns `audio/mpeg`).
3. Control the session (pause, resume, skip, stop) via PATCH.

The `guidance.next_best_actions` in every response tells you what to do next.

## Audio URLs

Each track in the response includes an `audio_url` field (e.g. `/audio/abc-123`). To play it:

```bash
curl -o track.mp3 https://chowdahh.com/audio/abc-123
# or stream directly to an audio player
```

The audio endpoint returns `Content-Type: audio/mpeg`. Audio is generated on first request and cached — subsequent fetches are instant.

## Endpoints

### `POST /api/v1/radio-sessions`

Start a radio session.

```bash
curl -X POST https://chowdahh.com/api/v1/radio-sessions \
  -H 'content-type: application/json' \
  -d '{"mode": "briefing", "duration_minutes": 5}'
```

Request fields:

- `mode` — `headlines`, `briefing`, or `topic_run`
- `duration_minutes` — target length for the session
- `topic_lenses[]` — optional topic filters (e.g. `["science", "world"]`)

The person is identified by the bearer token when present.

Response:

```json
{
  "data": {
    "radio_session_id": "radio_abc123",
    "state": "ready",
    "mode": "briefing",
    "queue_length": 8,
    "tracks": [
      {
        "id": "abc-123",
        "headline": "TSMC accelerates 1.4nm timeline",
        "audio_url": "/audio/abc-123",
        "image_url": "https://d1example.cloudfront.net/images/abc123ef/full.jpg",
        "topics": ["science", "business"],
        "source_count": 4
      },
      {
        "id": "def-456",
        "headline": "WHO declares new vaccine milestone",
        "audio_url": "/audio/def-456",
        "image_url": "https://example.com/vaccine-milestone.jpg",
        "topics": ["health"],
        "source_count": 7
      }
    ]
  },
  "guidance": {
    "status_explanation": "Radio session ready with 8 tracks in briefing mode. Each track has an audio_url that streams MP3 audio.",
    "next_best_actions": [
      { "action_id": "play_audio", "title": "Play the first track" },
      { "action_id": "control_radio", "title": "Control playback" }
    ]
  },
  "meta": { "request_id": "..." }
}
```

### `GET /api/v1/radio-sessions/{radio_session_id}`

Fetch current queue, playback position, and remaining tracks with audio URLs.

```bash
curl https://chowdahh.com/api/v1/radio-sessions/radio_abc123
```

Returns `tracks[]` from the current position onward.

### `PATCH /api/v1/radio-sessions/{radio_session_id}`

Control an active radio session. Returns updated state and remaining tracks.

```bash
curl -X PATCH https://chowdahh.com/api/v1/radio-sessions/radio_abc123 \
  -H 'content-type: application/json' \
  -d '{"action": "skip"}'
```

Actions: `pause`, `resume`, `skip`, `stop`

## Session States

- `ready` — queue built, not yet playing
- `playing` — audio in progress
- `paused` — paused by the user or agent
- `ended` — session complete or stopped

## Worked Flow

```
1. POST /api/v1/radio-sessions        → state: "ready", queue_length: 8
2. PATCH {id} {"action": "resume"}     → state: "playing"
3. PATCH {id} {"action": "skip"}       → state: "playing" (next track)
4. GET {id}                            → check current position
5. PATCH {id} {"action": "stop"}       → state: "ended"
```

## Audio Delivery

Audio is delivered per-track via the `audio_url` field on each track object. The URL pattern is `/audio/{track_id}`, which returns `audio/mpeg` (MP3).

Audio is synthesized on first request and cached on disk. Subsequent requests for the same track are served instantly from cache. Tracks are excluded from future sessions for 36 hours after a user listens to them.

## What Radio Is Not

- **Not per-card audio.** There is no endpoint to fetch audio for a single card ID.
- **Not replay.** Replay is prior card history. Radio is forward-looking audio delivery.
- **Not a TTS passthrough.** When implemented, the server will handle track generation; the agent manages the session.

## Why This Separation Helps AX

An agent can say:

> "I can show you your recent cards, or I can start Chowdahh Radio — 5 minutes of briefings."

Those are distinct product moves and should stay distinct in the API.
