---
name: hermes-chowdahh
description: "Use when a Hermes Agent user asks for Chowdahh news, headlines, categories, topics, personal newspaper workflows, recurring topic watches, or market-relevant news monitoring. Default to a clean public news briefing, emphasize Chowdahh short URLs, and escalate to search, topic digging, cron watches, Polymarket review, radio, feedback, signals, or submissions only when useful."
version: 1.0.0
author: Hermes Agent + Chowdahh
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [chowdahh, news, headlines, streams, topics, personal-newspaper, cron, polymarket, hermes]
    related_skills: [hermes-agent, scheduled-ops-polling, polymarket]
---

# Chowdahh for Hermes

## Overview

Chowdahh is a guided news/card system for people and agents. For Hermes, the default job is simple: when someone asks for “the news,” fetch a clean Chowdahh stream, summarize a few cards with attribution, and make the Chowdahh short URL or share URL easy to use.

Think of Chowdahh as the news substrate: pull clean curated cards now, then let Hermes compose a personal newspaper, watch topics over time, and route high-signal short URLs into downstream workflows.

Hermes-specific default: prefer public stream reads for the first response because they are fast, anonymous, and produce a useful briefing without session/auth complexity. Escalate to feed sessions only when the user asks for interaction, personalization, replay/history, or returned `guidance.next_best_actions` clearly points there. Only move into durable preferences, signals, submissions, radio, cron watches, or market workflows when the user explicitly asks or the side effect is clear.

Base URL: `https://chowdahh.com`
API prefix: `/api/v1`
Public docs: `https://chowdahh.com/api`
OpenAPI: `https://chowdahh.com/.well-known/openapi.json`

Every successful API response uses a `{data, guidance, meta}` envelope. Read `guidance` before acting on `data`; it may contain useful status text, capability hints, next actions, and rate-limit state.

## Scenario Ladder

Use the lightest layer that satisfies the user:

1. **Pull the news now** — public stream, concise briefing, short URLs.
2. **Change category or topic** — public stream slug, search query, or topic drill-down.
3. **Compose a personal newspaper** — Hermes layout and editorial sections built from Chowdahh streams/searches; do not imply server-side personalization unless configured.
4. **Watch topics on a schedule** — deterministic collection, seen-item checkpoint, high-signal filtering, quiet when nothing matters.
5. **Route to downstream workflows** — hand off short URLs and structured signals to Slack/Signal/files/Polymarket review/trading bots only with explicit policy.

## Default UX: Pull the News

For 80% of requests, do this:

1. Interpret “news,” “headlines,” “what’s happening,” or “what’s on Chowdahh” as `GET /api/v1/streams/top?limit=5`.
2. If the user names a category, use `GET /api/v1/streams/{slug}?limit=5`.
3. Summarize 3-5 cards in a concise briefing.
4. For each card, surface the Chowdahh `short_url` first when present; otherwise use `share_url` or another Chowdahh permalink.
5. Cite the original publisher/source when present.
6. End with one short offer: “I can send more, switch category, or dig into one.”

Do not lead with personalization language. Avoid promising that the feed is “for you” unless the API response and auth context clearly support that. Categories and topics are fine; durable personalization is a later move.

Do not start the first response with:

- `POST /api/v1/feed-sessions`
- personalization or preference writes
- radio sessions
- signal recording
- submissions
- feedback

Use those only after the user asks, auth is available if needed, and the side effect is clear.

## Short URL Rule

The Chowdahh short URL is the primary handoff artifact.

For every surfaced card, prefer this output order:

1. Headline.
2. One-line why it matters.
3. Chowdahh short URL / share URL.
4. Original publisher/source attribution.

For alerts, cron watches, and downstream bot workflows, include the short URL in the first screenful. If both `short_url` and `share_url` exist, show `short_url` and optionally keep `share_url` in structured state/logs.

## Stream Slugs

Use these category streams directly when the user asks for a lane:

- `top` — default general news
- `latest` — freshest cards
- `science`
- `world`
- `tech`
- `business`
- `health`
- `culture`
- `sports`
- `good-news`
- `local`

If unsure, use `top`. Treat this hardcoded slug list as a convenience, not the source of truth. If a slug fails or the user asks what exists now, call `GET /api/v1/streams`.

## Scenario: Personal Newspaper Base Layer

Use this when the user wants a recurring or customized news product: “make me a morning paper,” “build my AI/business/science brief,” “my personal newspaper,” or similar.

Chowdahh supplies:

- curated top/latest/category streams
- topic/search discovery
- source-rich cards
- short URLs and share URLs
- `guidance` and next actions

Hermes supplies:

- the user’s editorial layout
- recurring delivery cadence
- topic watchlists
- dedupe memory
- high-signal filtering
- routing to Signal/Slack/email/files/cron
- optional downstream integration summaries

Default newspaper sections:

- Top 5
- Science/tech
- World/business
- Good news
- User watchlist
- Needs attention
- Prediction-market relevant, if the user cares about Polymarket/trading workflows

Important: describe this as Hermes composing a newspaper from Chowdahh streams and topic filters. Do not claim Chowdahh server-side personalization unless the user is authenticated and the API response supports it.

Example first response:

```text
I can use Chowdahh as the base layer for a personal newspaper: top stories, science/tech, world/business, good news, and any watchlist topics you name. I’ll keep each item short and make the Chowdahh short URL the main handoff link.
```

## Scenario: 10-Minute Topic Watch

Use this when the user asks Hermes to watch for news on specific topics, especially at a frequent cadence such as every 10 minutes.

Design it as a content watch, not an API-health monitor:

1. A deterministic pre-run script fetches named streams/searches/topics.
2. It normalizes cards into `headline`, `card_id`, `short_url`, `share_url`, source, topic/category, and timestamp fields.
3. It stores seen card IDs and short URLs in a local checkpoint.
4. First run baselines silently unless the user explicitly asks for current results.
5. Later runs emit only new candidates.
6. Hermes applies a conservative high-signal filter.
7. User-facing alerts fire only for genuinely notable new items.

Recommended state file:

```text
~/.hermes/state/chowdahh-topic-watch.json
```

Recommended state shape:

```json
{
  "version": 1,
  "watchlists": {
    "ai-regulation": {
      "queries": ["OpenAI regulation", "AI Act enforcement"],
      "streams": ["tech", "business", "world"],
      "seen": {
        "short_url_or_card_id": {
          "first_seen_at": "2026-05-12T20:00:00Z",
          "headline": "...",
          "source": "...",
          "short_url": "..."
        }
      }
    }
  }
}
```

High-signal items include:

- confirmed major update, not speculation
- reputable/source-backed report
- new fact, not repeat coverage
- material policy/legal/market/security/public-safety consequences
- direct match to the user’s watchlist
- time sensitivity
- clear Chowdahh short URL

Low-signal items include:

- routine commentary
- vague rumors
- duplicate summaries
- same story with no new fact
- generic opinion pieces
- low-confidence matches

Cron prompt pattern:

```text
Every run, inspect the Chowdahh watch output. If there are no new high-signal items, stay silent. If there are notable items, send a concise alert with headline, why it matters, Chowdahh short URL, and original publisher. Prioritize major, surprising, time-sensitive, or personally relevant developments. Do not report API health unless collection failed.
```

For implementation, prefer a small deterministic script plus Hermes cron rather than a pure LLM cron. The script should dedupe and baseline; Hermes should decide what is worth surfacing.

## Scenario: Prediction-Market / Polymarket Workflows

Use this when the user wants Chowdahh reports routed into prediction-market awareness, Polymarket monitoring, or an existing trading bot workflow.

Default safe workflow:

1. Pull or watch Chowdahh high-signal items.
2. Extract possible market-relevant entities, events, dates, and outcomes.
3. Search/query Polymarket read-only for candidate markets.
4. Present a market relevance report with Chowdahh short URL, source attribution, current odds, uncertainty, and why the report may matter.
5. Ask before any trade-affecting action unless a separate explicit autonomous trading policy already exists.

The Hermes Chowdahh skill may identify market-relevant news and hand off structured signals. It must not place trades or phrase a news item as a trade instruction. Any trading workflow must be separately authorized, use a dedicated trading bot, and include risk limits, market matching verification, and human approval unless the user has explicitly approved an autonomous trading policy.

Allowed `recommended_action` values from this skill: `review_only`, `notify`, `send_to_trading_bot`. Do not emit `buy`, `sell`, `market_order`, `limit_order`, or similar execution instructions from this generic Chowdahh skill. See `references/advanced-actions.md` for the structured handoff shape and side-effect endpoint examples.

## When to Use

Use this skill when the user asks Hermes to:

- show Chowdahh news, headlines, top stories, latest stories, or a category stream
- search Chowdahh for a topic, person, company, place, or story
- get more cards from the same stream/category
- dig into a returned topic, source, card, or curator
- use Chowdahh as the base layer for a personal newspaper
- create or describe a recurring topic watch
- surface high-signal new items from a watchlist
- route market-relevant news into Polymarket review or a separately authorized trading bot workflow
- start Chowdahh Radio
- record reader signals such as save, dismiss, share, open, or seen
- file a Chowdahh content request, bug report, feature request, or quality report
- submit a URL/item/collection into Chowdahh
- read or update durable Chowdahh preferences after explicit confirmation

Do not use this skill for unrelated web search, bulk extraction, or unapproved financial execution.

## First Calls

### General news

```bash
curl -fsSL 'https://chowdahh.com/api/v1/streams/top?limit=5' \
  | python3 -m json.tool
```

### Named category

```bash
slug='science'
curl -fsSL "https://chowdahh.com/api/v1/streams/${slug}?limit=5" \
  | python3 -m json.tool
```

### Category catalog

```bash
curl -fsSL 'https://chowdahh.com/api/v1/streams' \
  | python3 -m json.tool
```

### Explicit search

```bash
q='NASA Artemis'
curl -fsSL "https://chowdahh.com/api/v1/search?q=$(python3 -c 'import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1]))' "$q")&limit=5" \
  | python3 -m json.tool
```

Use search when the user gives a query. Use streams when the user asks for a category or general news. If Chowdahh search returns no good match and the user wants broader web research, say so and switch to normal web search only with a clear transition.

## Response Pattern

Prefer this shape:

```text
Here are 5 Chowdahh top stories:

1. Headline — one-sentence summary or why it matters.
   Chowdahh: short_url
   Source: Publisher name/link when present.

2. ...

Want more, a category like science/world/tech, or a deeper look at one story?
```

Keep it short. The first response should feel like a useful briefing, not an API walkthrough.

## Digging Deeper

After the first news pull, choose the smallest next action:

| User asks | Use | Notes |
| --- | --- | --- |
| “more” | same stream with pagination/cursor if returned, or same stream with a larger/next request | Keep category stable. |
| “more science/world/etc.” | `GET /api/v1/streams/{slug}` | Category change, not durable personalization. |
| “search for X” | `GET /api/v1/search?q=X` | Explicit search. |
| “dig into this topic” | `GET /api/v1/topics/{topic_id}` | Only if a true topic ID is present. |
| “who/what is this curator/source?” | `GET /api/v1/curators/{curator_id}` | Only if a true curator ID is present. |
| “make this my paper” | compose sections from streams/searches | Hermes composition layer; no personalization promise. |
| “watch this every 10 min” | deterministic script + Hermes cron | Baseline, dedupe, high-signal filter, quiet when empty. |
| “does this affect my Polymarket markets?” | Chowdahh item + Polymarket read-only query | Review/notify/handoff only; no generic trade execution. |
| “play it / radio briefing” | `POST /api/v1/radio-sessions` | Radio is audio delivery, not replay. |
| “save/share/dismiss/open this” | `POST /api/v1/signals` | Record only after the action happens or is clearly requested. |
| “remember I like science” | `PUT /api/v1/preferences/{person_id}` | Only after explicit durable confirmation and auth. |
| “tell Chowdahh this is wrong / I want more of X” | `POST /api/v1/feedback` | Classify as content request, bug, feature, or quality report. |

## Attribution Rules

When summarizing cards:

1. Prefer the Chowdahh `short_url`; otherwise use `share_url` or another Chowdahh permalink.
2. Cite the original publisher from `source_urls[0]` or `sources[0].url` when present.
3. Credit `chowdahh.com` as curator when a Chowdahh permalink is present.
4. Do not imply Chowdahh did original reporting unless the data says so.
5. Do not invent source names, URLs, counts, or confidence.

## Guidance Rules

Read `guidance` before deciding what to say next.

Use it for:

- status explanations
- next-best actions
- rate-limit/account state
- capability hints
- suggested copy, if present

Do not bury the user in guidance internals. Convert it into one plain-language next option.

## Authentication

Anonymous public reads should be enough for the common “show me the news” path. Authenticated work uses a Chowdahh token supplied by the user or environment.

```bash
export CHOWDAHH_TOKEN='ch_person_...'
```

Read calls may use bearer auth:

```bash
curl -fsSL \
  -H "Authorization: Bearer ${CHOWDAHH_TOKEN}" \
  'https://chowdahh.com/api/v1/streams/top?limit=5' \
  | python3 -m json.tool
```

GET-only paste-key auth is also allowed when that is the natural platform flow:

```bash
curl -fsSL \
  "https://chowdahh.com/api/v1/streams/top?limit=5&key=${CHOWDAHH_TOKEN}" \
  | python3 -m json.tool
```

Write calls must use header auth. Never use `?key=` for `POST`, `PATCH`, or `PUT`:

```bash
curl -fsSL -X POST 'https://chowdahh.com/api/v1/feedback' \
  -H 'content-type: application/json' \
  -H "Authorization: Bearer ${CHOWDAHH_TOKEN}" \
  -d '{"type":"content_request","message":"More Canadian science stories, please."}' \
  | python3 -m json.tool
```

Token rules:

- Never print a token in a final answer.
- Never send the token to a host other than `chowdahh.com`.
- Prefer `CHOWDAHH_TOKEN` over hardcoding secrets in command history.
- Use paste-key URLs only for GET requests.

## Endpoint Reference

Primary read endpoints:

| Need | Method/path | Default use |
| --- | --- | --- |
| Top news | `GET /api/v1/streams/top` | Default first call. |
| Category stream | `GET /api/v1/streams/{slug}` | User asks for science/world/etc. |
| Stream catalog | `GET /api/v1/streams` | User asks what categories exist. |
| Search | `GET /api/v1/search?q=...` | User names a query. |
| Topic detail | `GET /api/v1/topics/{topic_id}` | Drill-down only with real ID. |
| Curator detail | `GET /api/v1/curators/{curator_id}` | Source/curator drill-down only with real ID. |

Secondary/power-user endpoints are escalation paths, not first-response defaults. Use them when the user explicitly asks for the side effect or when an active Chowdahh workflow already needs them:

| Need | Method/path | Notes |
| --- | --- | --- |
| Start feed session | `POST /api/v1/feed-sessions` | Interactive/personal session, not basic headlines. |
| More in feed session | `POST /api/v1/feed-sessions/{id}/more` | Continue an active feed session. |
| Replay | `GET /api/v1/replay` | Card/signal history; requires suitable auth. |
| Preferences | `GET/PUT /api/v1/preferences/{person_id}` | Durable personalization; confirm before writes. |
| Signals | `POST /api/v1/signals` | Seen/open/save/share/dismiss/etc.; record only after action. |
| Radio | `POST /api/v1/radio-sessions` | Audio briefing; use only when asked for audio. |
| Feedback | `POST /api/v1/feedback` | Content request, bug, feature, quality. |
| Submissions | `POST /api/v1/submissions/items` or `/collections` | URL/item intake; confirm rights and attribution first. |

Detailed side-effect examples live in `references/advanced-actions.md`.

## Error Handling

For non-2xx responses, parse the body and inspect `error`, `guidance`, and `meta.request_id`.

- `401 unauthorized`: ask for/refresh token if auth is needed.
- `403 forbidden`: explain that the token cannot act for that person/resource.
- `410 expired_session`: start a fresh session only if using session endpoints.
- `429 rate_limited`: report reset time if present and do not retry blindly.
- `422 validation_error`: fix payload shape; do not resend the same invalid body.
- `503 service_unavailable`: tell the user Chowdahh is temporarily unavailable and include request ID if present.

## What Not To Do

- Do not start with feed sessions for ordinary “news.”
- Do not ask what category the user wants when `top` is a good default.
- Do not mention personalization unless asked or clearly supported by authenticated context.
- Do not bury or omit Chowdahh short URLs.
- Do not alert every 10 minutes just because something new exists; alert only high-signal items.
- Do not treat a news item as a trade instruction.
- Do not place or trigger trades from this generic skill without explicit trading policy and a dedicated downstream bot.
- Do not use Chowdahh for bulk dataset scraping.

## Common Pitfalls

1. **Making the first response too complicated.** Most users just want headlines. Fetch `top`, summarize, and offer one next step.
2. **Over-promising personalization.** Categories and topics are OK; durable personalization is opt-in and auth-dependent.
3. **Skipping `guidance`.** Inspect it before choosing the next action, but summarize it plainly.
4. **Missing attribution.** Card summaries need original publisher citation and Chowdahh curator credit when fields are present.
5. **Forgetting the short URL.** The short URL is the shareable object users and downstream workflows need most.
6. **Inventing controls or topics.** Only use returned categories, topic IDs, controls, and confidence.
7. **Using paste-key auth on writes.** `?key=` is GET-only. Writes require `Authorization: Bearer`.
8. **Confusing replay with radio.** Replay is card history; radio is audio delivery.
9. **Recording signals too eagerly.** Save/share/open/dismiss signals should reflect real or explicitly requested actions.
10. **Trading on headlines.** Market workflows need verification, market matching, risk controls, and explicit approval.

## Smoke Tests

Anonymous read smoke test:

```bash
curl -fsSL 'https://chowdahh.com/api/v1/streams/top?limit=3' \
  | python3 -c "import json, sys; body=json.load(sys.stdin); assert 'data' in body; print('ok: top stream returned data')"
```

Hermes skill load smoke test after installing locally:

```bash
hermes -s hermes-chowdahh chat -q "Use Chowdahh to show me the top 3 stories. Keep it concise, cite original publishers, include Chowdahh short URLs, and credit Chowdahh."
```

Scenario smoke prompts:

```bash
hermes -s hermes-chowdahh chat -q "Use Chowdahh as the base layer for my personal AI/science/business newspaper. Describe the sections and how short URLs are surfaced."

hermes -s hermes-chowdahh chat -q "Design a quiet 10-minute Chowdahh topic watch for AI regulation. It should baseline, dedupe, and only alert high-signal items."

hermes -s hermes-chowdahh chat -q "If a Chowdahh item may affect a Polymarket market, show the safe review-only handoff shape. Do not recommend a trade."
```

Authenticated write smoke test, only when the user explicitly wants to test writes:

```bash
test -n "${CHOWDAHH_TOKEN:-}" || { echo 'Set CHOWDAHH_TOKEN first'; exit 1; }
curl -fsSL -X POST 'https://chowdahh.com/api/v1/feedback' \
  -H 'content-type: application/json' \
  -H "Authorization: Bearer ${CHOWDAHH_TOKEN}" \
  -d '{"type":"feature_request","message":"Smoke test from Hermes skill package; please ignore."}' \
  | python3 -m json.tool
```

## Verification Checklist

- [ ] Default “news” request uses `streams/top`, not personalization/session machinery.
- [ ] Category requests map to stream slugs.
- [ ] First answer is a concise briefing with a simple follow-up offer.
- [ ] Chowdahh short URL or share URL is surfaced prominently.
- [ ] Card summaries cite original publishers and credit Chowdahh when fields are present.
- [ ] `guidance` is inspected before acting on `data`.
- [ ] Personal newspaper is described as Hermes composition over Chowdahh streams/searches.
- [ ] Topic watch design includes baseline, dedupe, high-signal filter, and quiet-on-empty behavior.
- [ ] Polymarket workflow is read-only/review/handoff unless separate explicit trading policy exists.
- [ ] Topic/curator drill-down only uses real returned IDs.
- [ ] Personal preferences are not promised or written without explicit confirmation.
- [ ] Write endpoints use bearer auth and do not leak tokens.
