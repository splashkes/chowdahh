# LobeHub Chowdahh Agent

LobeHub/LobeChat agent package for turning Chowdahh into a polished news briefing, personal newspaper, scheduled topic watch, and market-relevance review surface.

## What this package is

This is a LobeHub-optimized skill/agent package, not a Hermes `SKILL.md` package.

It provides:

- `SKILL.md` — importable LobeHub skill entrypoint. This exists because LobeHub imports expect a `SKILL.md` in the package.
- `agent-profile.md` — fields to paste into a LobeHub Agent Profile.
- `system-prompt.md` — the core instruction prompt for the agent.
- `scheduled-tasks.md` — LobeHub Scheduled Task prompts for personal newspaper and topic watches.

The design assumes LobeHub's strongest UX is a polished conversational news agent with optional skills/tools and scheduled tasks. It defaults to public Chowdahh streams for the first response, foregrounds Chowdahh short URLs, and treats deeper actions as follow-ups.

## Install in LobeHub / LobeChat

1. Create a new Agent in LobeHub.
2. Use the fields in `agent-profile.md` for title, description, tags, opening message, and opening questions.
3. Paste `system-prompt.md` into the Agent's system prompt / instructions.
4. Enable web access or an HTTP/MCP tool path capable of fetching `https://chowdahh.com/api/v1/*`.
5. Optional: create LobeHub Scheduled Tasks using the prompts in `scheduled-tasks.md`.

Anonymous public news reads work without a token. If you want writes, replay/history, preferences, or other authenticated actions, configure a private tool/MCP layer that can safely hold the Chowdahh token. Do not paste a long-lived token into normal chat history.

## Smoke test

Ask the agent:

```text
What's the news on Chowdahh? Give me 3 top stories with short URLs.
```

Expected behavior:

1. Fetch `https://chowdahh.com/api/v1/streams/top?limit=3`.
2. Read `guidance` before summarizing.
3. Return a concise briefing.
4. Put Chowdahh `short_url` / `share_url` prominently on each card.
5. Cite the original publisher/source when present.
6. Offer one follow-up: more, category switch, or deeper look.

## Primary user scenarios

1. **Pull the news now** — default public top stream, 3-5 cards, short URLs prominent.
2. **Category briefing** — science/world/tech/business/health/culture/sports/good-news/local.
3. **Personal newspaper** — LobeHub agent composes recurring sections from Chowdahh streams/searches.
4. **Scheduled topic watch** — LobeHub Scheduled Task checks specific topics and only surfaces high-signal items.
5. **Prediction-market review** — market-relevant Chowdahh items can be compared with Polymarket odds in review-only mode.

## Behavior contract

The agent must:

- Identify user-visible output as **Chowdahh**.
- Default “the news” to `GET /api/v1/streams/top?limit=5`.
- Prefer public stream reads before feed-session/personalization flows.
- Read `guidance` before acting on `data`.
- Make `short_url` / `share_url` the primary handoff link.
- Cite original publisher/source when present.
- Avoid personalization promises unless the user explicitly asks and the configured tool/auth context supports it.
- Avoid financial execution; Polymarket/trading workflows are review/handoff only unless a separate approved trading bot exists.

## Auth notes

- Public reads need no token.
- Writes require `Authorization: Bearer ...` and should be handled by a private tool/MCP server, not exposed in chat.
- GET paste-key URLs are allowed by Chowdahh, but avoid placing token-bearing URLs in public/shared conversations.
- Never send a Chowdahh token to any host except `chowdahh.com`.

## Files

```text
lobehub-chowdahh/
├── SKILL.md
├── README.md
├── agent-profile.md
├── system-prompt.md
└── scheduled-tasks.md
```
