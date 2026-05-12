# LobeHub Scheduled Tasks for Chowdahh

Use these prompts in LobeHub Scheduled Tasks attached to the Chowdahh News agent.

LobeHub scheduled tasks run a prompt on a cadence. They may not have durable state unless your deployment/tooling provides one, so phrase tasks to use prior run history when available and to avoid noisy alerts.

## Daily personal newspaper

Suggested frequency: Daily, morning local time.

```text
Create my Chowdahh personal newspaper for today.

Use Chowdahh as the news substrate. Start with top stories, then include science/tech, world/business, good news, and any previously configured watchlist topics from this conversation.

For each item, include:
- headline
- one-line why it matters
- Chowdahh short URL or share URL
- original publisher/source when present

Keep the whole briefing skimmable. Do not claim server-side personalization; describe this as your composed Chowdahh newspaper. End with one question only if a choice would improve tomorrow's edition.
```

## 10-minute topic watch

Suggested frequency: Every 10 minutes only for time-sensitive topics. Otherwise use hourly or daily.

```text
Run a Chowdahh topic watch for: [TOPICS_OR_QUERIES].

Fetch relevant Chowdahh streams/searches. Compare against prior scheduled-task outputs in this conversation when available. Treat the first run as a baseline unless I explicitly asked for current results.

Only surface genuinely high-signal new items: major update, new fact, reputable/source-backed report, time-sensitive development, material policy/legal/market/security consequence, or direct match to the watchlist.

Stay quiet or say only “No high-signal Chowdahh updates” if there is nothing notable.

If there are high-signal items, include for each:
- headline
- one-line why it matters
- Chowdahh short URL/share URL
- original publisher/source when present
- why it crossed the alert threshold

Do not report API health unless collection failed. Do not alert on duplicate or low-signal commentary.
```

## Category digest

Suggested frequency: Daily or weekly.

```text
Create a Chowdahh [CATEGORY] digest.

Use the `[CATEGORY]` stream if available. If the stream slug fails or the available category list is uncertain, fetch the stream catalog and choose the closest matching stream.

Return 5 concise items. For each item, foreground the Chowdahh short URL/share URL and cite the original publisher/source when present. Offer one follow-up: more items, another category, or a deeper look at one story.
```

## Prediction-market relevance watch

Suggested frequency: Every 30-60 minutes for active market periods; otherwise daily.

```text
Review Chowdahh for reports that may be relevant to these prediction-market themes: [MARKET_THEMES].

Start from Chowdahh streams/searches. Identify only high-signal news items that could plausibly affect a market outcome. If a Polymarket read-only tool is available, search for candidate markets and report current odds. If no market tool is available, produce a review-only watchlist of candidate events.

For each item, include:
- headline
- Chowdahh short URL/share URL
- original publisher/source when present
- detected event/outcome
- candidate market question if available
- current odds if available
- why the report may matter
- uncertainty/caveat

Do not recommend or place trades. Allowed labels only: review_only, notify, or send_to_trading_bot if a separately authorized trading bot exists.
```

## Quiet scheduled-task principles

- Short URL first: every surfaced card needs a Chowdahh short URL or share URL when present.
- No noisy “new item” spam: alert only high-signal changes.
- No personalization overclaim: LobeHub composes the newspaper unless authenticated Chowdahh personalization is explicitly configured.
- No financial execution: market workflows are review/handoff only.
- No token exposure: do not ask for or print long-lived Chowdahh tokens in task output.
