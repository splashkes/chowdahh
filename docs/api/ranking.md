# Stream Ranking and Delivery

How cards are scored, ranked, and selected for delivery through the default API routes.

## Overview

When an agent calls `GET /api/v1/streams/top` or `POST /api/v1/feed-sessions`, the server:

1. Filters active clusters that meet quality thresholds.
2. Scores each cluster using a formula that balances significance, recency, and evergreen value.
3. Selects the top N by score.
4. Re-sorts the selected set by newest source time for display.

The result is: **quality determines what's included, recency determines the order you see it.**

## Significance Score

Each cluster (a group of related articles) gets a significance score:

```
significance = member_count + (max_curator_confidence x 3) + (avg_salience x 3)
```

| Component | Range | Weight | What it means |
|-----------|-------|--------|---------------|
| `member_count` | 1+ | x1 | Number of articles in the cluster. More coverage = more significant. |
| `max_curator_confidence` | 0.0-1.0 | x3 | Highest source authority score among cluster members. |
| `avg_salience` | 0.0-1.0 | x3 | Average newsworthiness of articles in the cluster. |

A major story with 10 articles from authoritative sources might score ~16. A 2-article cluster from unknown sources might score ~3.

## Evergreen Score

Each article gets an evergreen score (0.0-1.0) during synthesis:

- **0.0** — breaking/ephemeral (will be irrelevant tomorrow)
- **0.3** — default (single news cycle)
- **1.0** — reference/timeless (useful in 6 months)

The cluster's evergreen score is the average across its members. This stretches or compresses how fast a cluster decays out of the feed.

## Default Ranking Formula ("top")

The default rank mode combines significance with time-based boost and decay:

```
rank = significance
         x (1.0 + 0.5 x EXP(-age / 1800))
         / (1.0 + age / (43200 x (1.0 + evergreen x 4.0)))
```

Where `age` is seconds since the cluster's newest source was published.

### Recency Boost (numerator)

`1.0 + 0.5 x EXP(-age / 1800)`

- Brand new content gets up to **+50%** boost.
- Tapers over ~30 minutes.
- After an hour, the boost is negligible.

### Evergreen-Aware Decay (denominator)

`1.0 + age / (43200 x (1.0 + evergreen x 4.0))`

- Baseline decay half-life: **12 hours** (43200 seconds).
- High evergreen (1.0) stretches this to **60 hours** (5x slower decay).
- Low evergreen (0.0) keeps the 12-hour baseline.
- Breaking news fades in hours. Reference content persists for days.

## Quality Filters

A cluster must pass all of these to appear in results:

- `status = active`
- **Not superseded** — clusters that have been overtaken by newer reporting are suppressed (see Staleness Triage below)
- Has a non-empty headline
- At least **2 articles** in the cluster
- At least one member with **synthesis > 500 characters**
- `significance_score >= 5.5` (first page), `>= 3.5` (page 2+), `>= 2.0` (deep scroll)
- Not dismissed by the user in the last 48 hours (authenticated users only)

## Post-Selection Re-sort

After the ranking formula selects the top N clusters, they are **re-sorted by `latest_source_at` descending** for display. This means:

- The ranking formula controls **what makes the cut**.
- The display order is **newest first** within the selected set.

## Rank Modes

The default mode is `top`. Feed sessions can switch mode via `PATCH /api/v1/feed-sessions/{id}/controls` with `rank_mode`:

| Mode | Slug | Behavior |
|------|------|----------|
| **Boost + Decay** | `top` (default) | Full formula above — significance weighted by recency and evergreen |
| **Decay Only** | `latest` | Same formula without the recency boost — favors sustained stories |
| **Chronological** | `chronological` | Pure `latest_source_at` descending — no quality weighting |
| **Significance** | `significance` | Pure `significance_score` descending — ignores time |
| **Evergreen** | `evergreen` | `avg_evergreen` descending, then significance — surfaces reference content |

## Interest Filtering

When a feed session has interest controls applied (e.g. `science`, `world`), the query adds a topic filter: the cluster must have at least one topic matching the selected interests. The same ranking formula applies within the filtered set.

## Staleness Triage

In addition to time-based decay, an LLM evaluates active clusters every 3 hours to detect **supersession** — when a story has been overtaken by newer developments (e.g. "negotiations ongoing" → "war declared", "queen hospitalized" → "queen dies").

The triage evaluates each candidate cluster against:
- **Recent facts** on the cluster's topics (the narrative arc since the cluster was created)
- **Sibling clusters** (other active clusters on the same topics)

Three possible verdicts:

| Verdict | Action | Example |
|---------|--------|---------|
| **active** | No action — cluster stays in the stream | Story is still the current telling |
| **superseded** | Suppressed from stream, annotated with `superseded_by` pointing to the successor cluster | "Queen hospitalized" superseded by "Queen dies" cluster |
| **dissolve** | Cluster dissolved, members freed for re-clustering | Story is outdated but no successor cluster exists yet |

Superseded clusters remain accessible via permalinks — following the link redirects to the successor cluster. This means old shared links still work and point to the current version of the story.

Dissolution is rare and only used when the LLM can confirm the story moved on but cannot identify which cluster now carries it. The freed articles re-cluster in the next pipeline cycle, potentially joining or forming the correct successor.

## What This Means for Agents

- **Default delivery is opinionated.** High-significance, recent content surfaces first. You don't need to implement your own ranking.
- **Stale content is actively removed.** Superseded stories drop from the stream automatically — you don't need to track story arcs or filter outdated content yourself.
- **Controls adjust, don't replace.** Applying interest filters narrows the pool; changing rank mode changes the weighting. The quality floor stays.
- **Freshness is built in.** You don't need to filter by date — the decay formula handles it. Old content drops out naturally unless it's evergreen.
- **The significance floor drops with pagination.** Early pages are high-quality only. Deep scrolling surfaces more niche content.
- **Permalinks follow supersession chains.** If you link to a cluster that later gets superseded, the link resolves to the current successor.
