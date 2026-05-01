# Public Stream Serving Cache Plan

## Goal

Make public Chowdahh stream reads fast and safe for many users, agents, scripts, and TUI clients without putting WSL on the live request path.

The current slow path is `GET /api/v1/streams/{slug}` for category streams. `top` and `latest` are usable, but interest streams can take tens of seconds because the request path runs expensive topic/interest SQL joins.

## Design Principle

Treat public stream delivery as a read-model serving problem, not a handler-level optimization.

The Facebook-style pattern is layered serving:

```text
Cloudflare edge cache
  -> EC2 Agent API
  -> EC2 Redis stream read models
  -> EC2 Postgres serving replica
  -> WSL Postgres primary and pipeline
```

WSL remains the content factory. EC2 owns all public serving caches. Public user traffic must not call WSL directly.

## EC2 vs WSL Split

### EC2 Owns Public Serving

EC2 should own:

- `cmd/agentapi`
- EC2 Redis cache for public API stream data
- stream ID manifests for public stream slugs
- short-lived anonymous response/data cache
- Cloudflare tunnel origin
- reads from the replicated EC2 Postgres copy

Request path:

```text
Cloudflare -> EC2 agentapi -> EC2 Redis -> EC2 Postgres replica
```

### WSL Owns Content Production

WSL should own:

- ingestion
- extraction
- synthesis
- topic classification
- clustering
- tile computation
- primary content writes

WSL outputs durable state through Postgres replication. It should not be queried by live public API requests.

## Public vs Authenticated Semantics

### Public Stream Endpoints

These endpoints are public-stream semantics:

- `GET /api/v1/streams`
- `GET /api/v1/streams/{slug}`

For now, `Authorization: Bearer ...` on these endpoints should affect only:

- rate limit tier
- guidance/account metadata
- future permission affordances

It should not alter the card list unless a separate personalized stream contract is introduced.

That means both anonymous and authenticated callers can share the same EC2 stream read model. Authenticated responses must still be wrapped in a fresh envelope.

### Personalized Endpoints

These endpoints are private/session semantics:

- `POST /api/v1/feed-sessions`
- `GET /api/v1/feed-sessions/{id}`
- `POST /api/v1/feed-sessions/{id}/more`
- `PATCH /api/v1/feed-sessions/{id}/controls`
- `GET /api/v1/replay`
- `GET/PUT /api/v1/preferences/{person_id}`

They must not be Cloudflare-cached. They may use EC2 Redis private/session caches.

## Cacheability Rules

### Safe For Shared EC2 Cache

Use shared EC2 read models for:

- anonymous `GET /api/v1/streams`
- anonymous `GET /api/v1/streams/{slug}`
- authenticated `GET /api/v1/streams`
- authenticated `GET /api/v1/streams/{slug}`, as long as stream content remains public-stream semantics

The shared cache stores only data needed to render the stream, not the final response envelope.

### Safe For Cloudflare Cache

Cloudflare may cache only anonymous public reads:

- method is `GET` or `HEAD`
- path is `/api/v1/streams` or `/api/v1/streams/*`
- no `Authorization` header
- no `key` query parameter

Bypass Cloudflare cache for:

- any `Authorization` header
- any `key` query parameter
- any non-GET/HEAD method
- feed sessions
- replay
- preferences
- signals
- feedback
- submissions
- radio session mutation

### Never Cache Publicly

Never Cloudflare-cache:

- `?key=` requests
- Bearer-authenticated responses
- `POST`, `PUT`, `PATCH`, or `DELETE`
- user preference/replay/session responses

## What To Cache

Do not cache full JSON envelopes.

Full envelopes contain request-specific data:

- `meta.request_id`
- rate-limit guidance
- auth/account state

Cache only:

```text
stream data
pagination metadata
ordered cluster IDs
optional hydrated card JSON
```

Then build a fresh envelope per request through `internal/agentapi/envelope.go`.

## Proposed Redis Keys

Stream discovery:

```text
agentapi:cache:v1:streams
```

Stream ID manifests:

```text
agentapi:stream-ids:v1:top
agentapi:stream-ids:v1:latest
agentapi:stream-ids:v1:science
agentapi:stream-ids:v1:world
agentapi:stream-ids:v1:tech
agentapi:stream-ids:v1:business
agentapi:stream-ids:v1:health
agentapi:stream-ids:v1:culture
agentapi:stream-ids:v1:sports
agentapi:stream-ids:v1:good-news
agentapi:stream-ids:v1:local
```

Optional page data cache:

```text
agentapi:stream-page:v1:{slug}:limit:{limit}:offset:{offset}
```

Private/session cache examples:

```text
agentapi:user-stream:v1:{person_id}:{slug}:prefs:{hash}:limit:{limit}:offset:{offset}
agentapi:feed-session:v1:{session_id}
```

## TTLs

Recommended initial TTLs:

| Cache | TTL |
|---|---:|
| `/api/v1/streams` discovery | 5 minutes |
| stream ID manifests | refreshed every 1 minute; expire after 5 minutes |
| anonymous first-page stream data | 30-60 seconds |
| anonymous later-page stream data | 60-120 seconds |
| authenticated public-stream data reuse | same shared stream read model; no shared envelope cache |
| private/session cache | 15-60 seconds or session TTL, depending on endpoint |

Use stale-if-error behavior where possible: serving a 1-5 minute stale public stream is better than timing out.

## Headers

For anonymous stream responses that are Cloudflare-cacheable:

```text
Cache-Control: public, max-age=15, s-maxage=60, stale-while-revalidate=120
Vary: Authorization
```

For Bearer-authenticated public stream responses:

```text
Cache-Control: private, max-age=15
Vary: Authorization
```

For `?key=` responses:

```text
Cache-Control: no-store, private
Pragma: no-cache
Referrer-Policy: no-referrer
Vary: Authorization
```

For personalized/session responses:

```text
Cache-Control: private, no-store
Vary: Authorization
```

## Implementation Plan

### Phase 1: Shared Data Cache For Public Streams

Add cache helpers in the main service repo:

- `internal/redis/api_cache.go`
- optional `internal/agentapi/cache.go`

Functions:

```go
GetJSON(ctx, key, dest)
SetJSON(ctx, key, value, ttl)
GetStringSlice(ctx, key)
SetStringSlice(ctx, key, values, ttl)
```

Wire `internal/agentapi/handlers/streams.go`:

1. Parse `slug`, `limit`, and `offset`.
2. Determine identity from context.
3. If path is public-stream semantics, try shared stream page/data cache.
4. On hit, wrap fresh guidance and request metadata.
5. On miss, fall back to existing DB path and write cache.

Acceptance:

- `GET /api/v1/streams/latest?limit=10` returns same card shape.
- `GET /api/v1/streams/tech?limit=10` returns under 1 second on warm cache.
- Bearer requests reuse stream data but get fresh rate-limit guidance.
- `?key=` responses are not cached.

### Phase 2: Stream ID Manifest Builder On EC2

Add an EC2-local builder in `cmd/agentapi` or scoring:

```text
every 60 seconds:
  build ordered IDs for each public stream slug
  write agentapi:stream-ids:v1:{slug}
  set 5-minute TTL
```

For the first implementation, it can use existing DB query functions off-path:

- `ListStreamClusters`
- `ListStreamClustersByInterests`

Later, replace category query internals with `tile_stream_cache` or a dedicated materialized stream table.

Acceptance:

- Warm manifest exists for every public slug.
- Request path reads IDs from Redis first.
- Cache builder failure does not break serving; stale manifests can be used until TTL.

### Phase 3: Hydrate By ID

Use existing server code:

- `internal/db/views.go::HydrateStreamClustersByIDs`

Request path:

1. read IDs from `agentapi:stream-ids:v1:{slug}`
2. slice by `offset` and `limit`
3. hydrate IDs
4. convert with existing wire card adapter
5. return page metadata

Acceptance:

- Category stream requests no longer run `ListStreamClustersByInterests` on cache hit.
- Pagination uses `offset` consistently.
- Card order matches manifest order.

### Phase 4: Cloudflare Edge Cache

After origin cache correctness is verified, add Cloudflare rules:

Cache:

```text
GET /api/v1/streams
GET /api/v1/streams/*
```

Bypass:

```text
http.request.headers["authorization"][exists]
or url.query contains "key="
or method not in {"GET", "HEAD"}
```

Acceptance:

- Anonymous repeated requests show Cloudflare cache hits.
- Bearer requests bypass Cloudflare but hit EC2 Redis read model.
- Query-key requests bypass Cloudflare and include no-store headers.

### Phase 5: Remove Slow Runtime Category Query From Hot Path

Once stream manifests are reliable, treat `ListStreamClustersByInterests` as builder-only or fallback-only.

Preferred long-term paths:

1. map public category slugs to precomputed stream manifests
2. use `tile_stream_cache` for topic/category matching
3. optionally create a dedicated `api_stream_cache` table if Redis-only rebuilds are too opaque

Acceptance:

- p95 `GET /api/v1/streams/{category}?limit=10` under 500ms warm.
- p99 under 1s warm.
- cold-cache fallback is observed but rare.

## Metrics

Add low-cardinality metrics:

```text
AgentAPI.StreamCache.Hit{slug,auth_mode}
AgentAPI.StreamCache.Miss{slug,auth_mode}
AgentAPI.StreamCache.BuildMS{slug}
AgentAPI.StreamCache.BuildRows{slug}
AgentAPI.StreamCache.StaleServe{slug}
AgentAPI.StreamCache.FallbackDB{slug}
AgentAPI.StreamCache.Error{slug,stage}
```

Avoid labels with query strings, tokens, user IDs, or raw request paths.

## Failure Behavior

Preferred fallback order:

1. fresh Redis stream manifest/page
2. stale Redis stream manifest/page
3. EC2 Postgres DB query
4. empty response with guidance only if DB fails or times out

Do not call WSL from the live request path.

## Security Notes

- Shared cache is allowed only for public-stream data.
- Authenticated public stream responses can reuse shared data but must have fresh envelopes.
- User-private responses must use private cache keys.
- Query-key requests must be `no-store`.
- Never include tokens, raw query strings, or user IDs in cache keys that might be logged externally.

## Rollout Checklist

1. Implement origin Redis data cache behind an env flag.
2. Deploy to EC2 only.
3. Verify warm/cold latency for all stream slugs.
4. Add stream manifest builder.
5. Switch handler to manifest hydration on hit.
6. Add headers for cacheable anonymous stream responses.
7. Add Cloudflare cache rule.
8. Monitor hit rate, fallback count, p95/p99 latency, 5xx, and rate-limit behavior.
9. Remove or demote runtime category SQL from the hot path.

## Non-Goals

- Do not cache personalized feed-session responses publicly.
- Do not route live traffic to WSL.
- Do not require exact real-time freshness for anonymous public streams.
- Do not solve all ranking semantics in this phase.

Freshness target for public streams is 30-120 seconds. That is acceptable for a public news stream and dramatically better than multi-second or timeout-prone category queries.
