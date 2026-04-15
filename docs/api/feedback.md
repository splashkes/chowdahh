# Feedback

## Goal

Make it easy for agents to send useful requests back into Chowdahh without forcing everything into "report a bad card."

## Endpoint

### `POST /api/v1/feedback`

Suggested feedback types:

- `content_request`
- `bug_report`
- `feature_request`
- `quality_report`

## Why This Matters

Agents often learn things the system cannot infer directly:

- the person wants more of a topic
- a card is grouped incorrectly
- a flow is broken
- a product capability is missing

These should all use one clear surface.

Note: feedback validation failures return `{error, meta}` without a `guidance` block. Clients should not assume guidance is present on error responses.

## Request Shape

- `feedback_type` (required)
- `title` (required)
- `detail`
- optional `topic_id`
- optional `card_id`
- optional `session_id`

The person is identified by the bearer token when present.

## Quality Reports

Use `quality_report` when the issue is with current content quality:

- bad summary
- wrong grouping
- missing context
- bad attribution

## Content Requests

Use `content_request` when the person wants more or different content:

- "more Canada science"
- "send good news for mornings"
- "show more like this topic"

That keeps product requests and content steering in one coherent system.
