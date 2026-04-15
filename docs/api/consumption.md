# Feed And Controls

## Goal

Give agents one obvious entry point for "send content now", then clear ways to send more, adjust controls, and drill deeper.

## Primary Call

### `POST /api/v1/feed-sessions`

This is the default "send content for this person" call.

Request shape:

- `intent` (e.g. `browse`)
- `budget_minutes`
- `include_controls`

The person is identified by the bearer token, not a request body field.

Response envelope `{data, guidance, meta}`:

- `data.session_id`
- `data.items[]` — each item includes `id`, `headline`, `summary`, `image_url`, `topics`, `source_count`, `share_url`
- `data.count`
- `data.controls`
- `guidance.status_explanation`
- `guidance.next_best_actions[]`

Cards include `image_url` when a hero image is available. Image URLs point to the original source or a CDN-cached version — fetch them directly, no auth needed.

The response intentionally includes `controls` so an agent can say:

> "I can keep this balanced, tilt it toward science and good news, or send more."

without inventing its own taxonomy.

## Continue The Session

### `POST /api/v1/feed-sessions/{session_id}/more`

This is the explicit `send more` call.

The session should preserve:

- already sent cards
- applied controls
- why the session exists

## Update Controls

### `PATCH /api/v1/feed-sessions/{session_id}/controls`

Controls should mirror the product structurally.

Suggested groups:

- sort or mode
- topic chips
- tone chips
- location chips

## Drill-Down Calls

### `GET /api/v1/topics/{topic_id}`

Returns:

- topic summary
- timeline or developments
- original sources
- source URLs
- curator info
- related topics

### `GET /api/v1/search`

Searches clusters by topic match. Results are card objects with `id`, `headline`, `summary`, etc. Anonymous search results do not currently expose stable result types or drill-down IDs.

### `GET /api/v1/curators/{curator_id}`

Useful when the person wants "more from this kind of source" or to understand why a topic keeps surfacing.

Note: topic and curator drill-down endpoints require true internal identifiers. Anonymous search results do not reliably expose those identifiers, so a direct search → drill-down flow is not yet supported for anonymous clients.

## Card Sources

Each card includes a `sources[]` array with the top source articles backing it. This lets clients display attribution and link to original reporting.

Each source has:

- `title` — article headline
- `source_url` — link to the original article
- `domain` — publisher domain (e.g. `reuters.com`)
- `creator_name` — author or publisher name, when known
- `published_at` — publication timestamp

The array is ordered by relevance and may be truncated. Use `source_count` for the true total.

```json
{
  "sources": [
    {
      "title": "Iran rejects US demands at Islamabad talks",
      "source_url": "https://reuters.com/world/iran-rejects-...",
      "domain": "reuters.com",
      "published_at": "2026-04-12T14:30:00Z"
    },
    {
      "title": "Hormuz blockade begins as ceasefire expires",
      "source_url": "https://theguardian.com/world/2026/...",
      "domain": "theguardian.com",
      "creator_name": "Patrick Wintour",
      "published_at": "2026-04-12T16:00:00Z"
    }
  ]
}
```

## Public Lanes

Public high-volume browsing still matters. This surface should support:

- `GET /api/v1/streams/{stream_slug}`
- `GET /api/v1/topics/{topic_id}`
- `GET /api/v1/search`

Discover available streams via `GET /api/v1/streams`. The current default set includes:

- `top`
- `latest`
- `science`
- `world`
- `tech`
- `business`
- `health`
- `culture`
- `sports`
- `good-news`
- `local`

These are not the entire personalization model. They are public lanes and entry ramps. Clients should call the discovery endpoint rather than hardcoding this list.

## Honest Good News

`good-news` should not mean "soft vibes." It should be backed by explicit content classification with transparent caveats:

- uplifting
- constructive
- recovery
- delight

If confidence is weak, the API should say so in metadata instead of silently pretending the lens is precise.
