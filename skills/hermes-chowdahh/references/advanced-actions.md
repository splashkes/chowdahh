# Hermes Chowdahh Advanced Actions

These examples are escalation/reference material for the Hermes Chowdahh skill. The main `SKILL.md` should remain focused on the default public-news briefing, short URLs, personal newspaper composition, topic watches, and safety rules.

Use these actions only after the user asks, auth is available if needed, and the side effect is clear.

## Auth helper

```bash
AUTH_ARGS=()
if [ -n "${CHOWDAHH_TOKEN:-}" ]; then
  AUTH_ARGS=(-H "Authorization: Bearer ${CHOWDAHH_TOKEN}")
fi
```

Writes require bearer auth. Do not use `?key=` on `POST`, `PATCH`, or `PUT`.

## Radio

Use radio only when the user asks for audio, a radio briefing, or playback.

```bash
AUTH_ARGS=()
if [ -n "${CHOWDAHH_TOKEN:-}" ]; then
  AUTH_ARGS=(-H "Authorization: Bearer ${CHOWDAHH_TOKEN}")
fi
curl -fsSL -X POST 'https://chowdahh.com/api/v1/radio-sessions' \
  -H 'content-type: application/json' \
  "${AUTH_ARGS[@]}" \
  -d '{"mode":"briefing","duration_minutes":5}' \
  | python3 -m json.tool
```

Radio responses contain `tracks[]` with `audio_url`. Fetch audio only if the user wants playback or a file:

```bash
track_path='/audio/abc-123'
curl -fsSL "https://chowdahh.com${track_path}" -o /tmp/chowdahh-track.mp3
```

## Signals

Record signals only after the action happens or the user clearly asks you to record it.

```bash
curl -fsSL -X POST 'https://chowdahh.com/api/v1/signals' \
  -H 'content-type: application/json' \
  -H "Authorization: Bearer ${CHOWDAHH_TOKEN}" \
  -d '{"signals":[{"card_id":"card_...","signal_type":"save"}]}' \
  | python3 -m json.tool
```

Common signal types: `seen`, `open`, `source_open`, `save`, `share`, `dismiss`, `close`, `react`, `dwell`, `tts_play`, `tts_listen`, `track`.

## Feedback

Use feedback when the user wants to ask for more content, report a bug, request a feature, or flag a quality issue.

```bash
curl -fsSL -X POST 'https://chowdahh.com/api/v1/feedback' \
  -H 'content-type: application/json' \
  -H "Authorization: Bearer ${CHOWDAHH_TOKEN}" \
  -d '{"type":"quality_report","message":"This card grouped unrelated stories together.","context":{"card_id":"card_..."}}' \
  | python3 -m json.tool
```

Feedback types: `content_request`, `bug_report`, `feature_request`, `quality_report`.

## Submissions

Before submission, confirm title/headline, source URL, creator attribution if known, and preservation vs synthesis preference.

```bash
curl -fsSL -X POST 'https://chowdahh.com/api/v1/submissions/items' \
  -H 'content-type: application/json' \
  -H "Authorization: Bearer ${CHOWDAHH_TOKEN}" \
  -d '{"title":"Example story","source_url":"https://example.com/story","creator_attribution":"Example Publisher","transform":"preserve"}' \
  | python3 -m json.tool
```

## Polymarket / trading-bot handoff

The Hermes Chowdahh skill may identify market-relevant news and hand off structured signals. It must not place trades or phrase a news item as a trade instruction.

Allowed `recommended_action` values from this skill:

- `review_only`
- `notify`
- `send_to_trading_bot`

Do not emit `buy`, `sell`, `market_order`, `limit_order`, or similar execution instructions from this generic Chowdahh skill.

Structured handoff shape for an authorized downstream bot:

```json
{
  "source": "chowdahh",
  "headline": "...",
  "short_url": "...",
  "share_url": "...",
  "original_source_url": "...",
  "publisher": "...",
  "detected_event": "...",
  "candidate_markets": [
    {
      "market_id": "...",
      "question": "...",
      "yes_price": 0.42,
      "no_price": 0.58,
      "relevance": "direct"
    }
  ],
  "signal_strength": "high",
  "reason": "...",
  "recommended_action": "review_only"
}
```

Any trading workflow must be separately authorized, use a dedicated trading bot, and include risk limits, market matching verification, and human approval unless the user has explicitly approved an autonomous trading policy.
