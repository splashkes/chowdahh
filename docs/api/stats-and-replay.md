# Replay And Stats

## Why This Matters

Agents need more than retrieval. They need memory hooks backed by real product state.

The product should answer:

- what did this person open recently
- what did they save
- what did they share this month
- what did they dismiss
- what did they already see

## Signals

`POST /api/v1/signals`

Signals are sent as a batch array. Each signal needs `signal_type` and `card_id`.

First-party signal types:

- `seen`
- `open`
- `save`
- `share`
- `dismiss`
- `source_open`

Each signal can optionally include:

- `topic_id`
- `submission_id`
- `source_url`
- `shared_to`
- `session_id`

The person is identified by the bearer token when present.

Note: invalid or unrecognized signal types are silently skipped. Callers should inspect `recorded` in the response body — a `200` status does not mean all signals were accepted.

## Replay Query

`GET /api/v1/replay`

Requires a person token. The person is identified by the bearer token.

Example queries:

- `?period=today`
- `?signal_type=share&period=this_month`
- `?signal_type=save&period=last_7_days`

Suggested response fields:

- `events[]`
- `next_cursor`
- `window`

Replay is the ordered history.

## Stats Query

`GET /api/v1/stats/activity`

Example queries:

- `?signal_type=share&period=this_month`
- `?signal_type=save&period=last_7_days`
- `?group_by=topic&period=this_month`

Suggested response fields:

- `total`
- `items[]`
- `grouped_counts[]`
- `window`

## Replay

Replay can be a view over signals, not a separate object model at first.

Useful derived endpoints:

- `GET /api/v1/replay?period=today`
- `GET /api/v1/replay?signal_type=share&period=this_month`
- `GET /api/v1/stats/activity?group_by=topic&period=this_month`

This keeps the initial model small while still enabling:

- "show me what I shared this month"
- "show me the last three cards I opened"
- "show me what they dismissed this week"
