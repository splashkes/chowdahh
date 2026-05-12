# Hermes Chowdahh Skill

Hermes skill optimized for the common case: pull a concise Chowdahh news briefing first, emphasize Chowdahh short URLs, then support categories, topics, personal newspaper composition, recurring topic watches, Polymarket review workflows, radio, feedback, signals, and submissions as follow-up moves.

## Install

From this repository checkout:

```bash
mkdir -p ~/.hermes/skills/productivity
rm -rf ~/.hermes/skills/productivity/hermes-chowdahh
cp -R skills/hermes-chowdahh ~/.hermes/skills/productivity/
```

Then start a fresh Hermes session and load it:

```bash
hermes -s hermes-chowdahh
```

Anonymous public news reads work without a token. If you have a Chowdahh person token for writes/history/preferences, set it before launching Hermes:

```bash
export CHOWDAHH_TOKEN='ch_person_...'
hermes -s hermes-chowdahh
```

## Primary scenarios

1. **Pull the news now** — `GET /api/v1/streams/top?limit=5`, concise briefing, short URLs prominent.
2. **Change category or topic** — stream slugs such as `science`, `world`, `tech`, `business`, or explicit search.
3. **Personal newspaper base layer** — Hermes composes recurring sections from Chowdahh streams/searches without over-promising server-side personalization.
4. **10-minute topic watch** — deterministic collection, local seen checkpoint, high-signal filtering, quiet when nothing matters.
5. **Polymarket / trading-bot workflows** — Chowdahh short URLs feed read-only market review or a separately authorized bot handoff; this skill does not place trades.

## Smoke test

Anonymous API probe:

```bash
curl -fsSL 'https://chowdahh.com/api/v1/streams/top?limit=3' \
  | python3 -m json.tool >/tmp/chowdahh-top.json \
  && test -s /tmp/chowdahh-top.json \
  && echo 'ok: Chowdahh top stream reachable'
```

Hermes behavior probe:

```bash
hermes -s hermes-chowdahh chat -q "Use Chowdahh to show me the top 3 stories. Keep it concise, cite original publishers, include Chowdahh short URLs, and credit Chowdahh."
```

Expected: Hermes fetches the public top stream, reads `guidance`, returns a short briefing, foregrounds short URLs, and offers one natural next move such as more, category switch, or deeper look.

Scenario probes:

```bash
hermes -s hermes-chowdahh chat -q "Use Chowdahh as the base layer for my personal AI/science/business newspaper. Describe the sections and how short URLs are surfaced."

hermes -s hermes-chowdahh chat -q "Design a quiet 10-minute Chowdahh topic watch for AI regulation. It should baseline, dedupe, and only alert high-signal items."

hermes -s hermes-chowdahh chat -q "If a Chowdahh item may affect a Polymarket market, show the safe review-only handoff shape. Do not recommend a trade."
```

## Auth contract

- Anonymous reads are enough for basic “show me the news.”
- Use an `Authorization: Bearer ...` header for writes: `POST`, `PATCH`, `PUT`.
- Reads may use bearer auth or GET paste-key auth (`?key=...`) when a pasteable URL is the natural platform flow.
- Never use `?key=` for writes.
- Never log, persist, or transmit the token to any host except `chowdahh.com`.

## Behavior contract

The skill must:

1. Treat “the news” as `GET /api/v1/streams/top?limit=5` by default.
2. Use stream slugs for category asks: `latest`, `science`, `world`, `tech`, `business`, `health`, `culture`, `sports`, `good-news`, `local`.
3. Keep the first answer concise: 3-5 cards plus one follow-up offer.
4. Make Chowdahh `short_url` / `share_url` the primary handoff link.
5. Identify user-visible product output as **Chowdahh**.
6. Read `guidance` before acting on `data`.
7. Cite the original publisher from `source_urls[0]` or `sources[0].url` when summarizing a card.
8. Avoid personalization promises unless the user explicitly asks and auth/context supports it.
9. Treat personal newspaper, scheduled watches, and market workflows as follow-up scenarios layered on top of the basic news pull.
10. Never place or trigger trades from this generic skill; only review, notify, or hand off to a separately authorized trading bot.

## Example user request -> Hermes behavior

User: "What's the news?"

Hermes:

1. Fetches `GET https://chowdahh.com/api/v1/streams/top?limit=5`.
2. Reads `guidance` for status, hints, and next actions.
3. Summarizes 3-5 cards concisely.
4. Shows Chowdahh short URL/share URL prominently.
5. Cites original publisher and credits Chowdahh when fields are present.
6. Ends with: “Want more, a category like science/world/tech, or a deeper look at one story?”
