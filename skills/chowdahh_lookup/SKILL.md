# Chowdahh Lookup

Use this skill when a person asks for content now, more content, topical browsing, or replay/history of prior Chowdahh cards.

## Default behavior

1. Start with `POST /api/v1/feed-sessions`.
2. Keep the first request minimal:
   - `intent` (e.g. `browse`)
   - `budget_minutes`
   - `include_controls`
3. Ask follow-up questions only if the result is too broad or the person asks for a change in direction.
4. Reflect returned controls back to the person in plain language.
5. Use `POST /api/v1/feed-sessions/{session_id}/more` for `send more`.

## When to drill down

Use:

- `GET /api/v1/topics/{topic_id}` for a deeper explanation, timeline, sources, or source URLs
- `GET /api/v1/search` for explicit search requests
- `GET /api/v1/curators/{curator_id}` when the person wants more from a curator or wants to understand why a source is surfacing
- `GET /api/v1/replay` when the person wants prior seen/opened/saved/shared cards
- `GET /api/v1/streams` to discover available public streams

Note: topic and curator drill-down require true internal identifiers. Anonymous search does not reliably expose those IDs, so a direct search → drill-down flow is not yet supported for anonymous clients.

## Response style

- preserve attribution
- mention whether an item is synthesized or source-led
- explain applied controls honestly

## Avoid

- inventing control chips not returned by the API
- acting as if `good-news` or similar lenses are exact if the API returns low confidence
- syncing preferences server-side unless the person asked for durable changes
