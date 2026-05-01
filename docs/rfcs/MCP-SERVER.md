# RFC: Chowdahh MCP Server v1

**Status:** Draft for review
**Owner:** TBD
**Depends on:** PR `audit/openapi-v0.3-reconciliation` (must merge first)

---

## Summary

This RFC proposes a Model Context Protocol (MCP) server for Chowdahh, hosted at `mcp.chowdahh.com/v1/mcp` over HTTP Streamable. The server exposes 25 tools covering Chowdahh's discovery, session, signal, submission, and feedback surfaces. Tools are generated mechanically from the v0.3.0-reconciled OpenAPI spec via a set of `x-mcp-*` extensions added in that audit.

The design is **stateless edge proxy** — every MCP tool call translates to a single REST API call against `chowdahh.com/api/v1/`. The MCP server adds no new persistence, no new auth tokens, and no new business logic.

The design is **phased** — 8 read-only tools in Phase 1, 8 session/control tools in Phase 2, 9 mutation tools in Phase 3. Phase 1 ships independently; Phases 2 and 3 are gated on resolving open questions and on landing prerequisite server-side fixes.

The design is **honest about what's not yet ready** — three of the 25 tools require new server endpoints, and several Phase 3 tools carry persistence caveats that should be either fixed server-side or surfaced in tool descriptions until they are.

---

## Motivation

Chowdahh's positioning is "the content layer for the agentic web." That posture demands a first-class MCP surface. AI agents already discover tools through MCP host catalogues (Claude Code, Cursor, Claude Desktop, ChatGPT-with-MCP); without an MCP server, Chowdahh is invisible to the discovery path that increasingly defines whether a content API gets used.

The competitive benchmark is Perigon's MCP server (`mcp.perigon.io`), which exposes 8 search-only tools. Chowdahh's surface is richer than Perigon's — sessions, signals, synthesis, guidance — and the MCP server should reflect that richness rather than collapse to search parity.

---

## Design

### Architecture

```
MCP Host (Claude Code, Cursor, etc.)
        │  HTTP Streamable / JSON-RPC 2.0
        ▼
mcp.chowdahh.com/v1/mcp  (Cloudflare Worker, stateless)
        │  HTTPS, Authorization: Bearer ch_*
        ▼
chowdahh.com/api/v1/  (existing REST)
```

The Worker is a thin translator. Tool invocations become REST calls. Response envelopes (`{data, guidance, meta}`) become MCP tool results with `structuredContent` and `_meta` fields. Authentication is delegated to the REST backend by forwarding the Bearer token verbatim.

### Tool result shape

Each tool call produces an MCP tool result of the form:

```json
{
  "content": [
    { "type": "text", "text": "<rendered text including 'Next best actions:' block when present>" }
  ],
  "structuredContent": <Envelope.data>,
  "_meta": {
    "guidance": <Envelope.guidance>,
    "request_id": "<from Envelope.meta.request_id>",
    "next_cursor": "<when paginated>",
    "reasoning": "<when underlying response carries Reasoning fields>"
  }
}
```

The `structuredContent` field carries the typed `data` payload from the REST response. The `_meta.guidance` field carries the full guidance block — `next_best_actions`, `account_state`, `capability_hints` — for AI agents that route on it. The `_meta.reasoning` field carries algorithmic decision metadata (confidence, rule version, triggered rules) when the underlying response includes it.

This dual surfacing is deliberate: text content is for the LLM's natural reading, structured fields are for routing logic.

### Tool naming

`verb_object` snake_case, action-first: `start_feed_session`, `update_preferences`, `submit_item`. This convention mirrors the action vocabulary in `agent.txt` and is more expressive than Perigon's `search_<entity>` mono-pattern, which would not extend to mutations.

### Authentication and the two-principal model

Two senses of "agent" appear in this surface and must not be conflated:

- **Human agent** — the principal of the bearer token. A Person (`ch_person_*`) or Curator (`ch_cur_*`) who owns resources and accrues signal history. **Authorization keys off this.**
- **AI agent** — the MCP host runtime invoking tools. **Approval friction, telemetry, and rate-shaping key off this**, via the `X-Chowdahh-Agent-Id` and `X-Chowdahh-Agent-Name` headers (already documented in the spec).

The MCP server forwards the bearer token unmodified to the REST API and populates the AI-agent-runtime headers from the host's identifier when available. Token issuance remains the responsibility of the existing sign-in flow.

### Session state

Feed sessions and radio sessions are stateful on the REST server (Redis, 12-hour TTL). The MCP surface keeps the protocol-level interaction stateless: tools return opaque session IDs, and AI agents pass them back on subsequent calls. Separate `start_*` and `continue_*` tools rather than collapsed alternatives.

The MCP server itself maintains no per-connection state. If future telemetry needs change this, the storage choice should favor queryability (D1 if Cloudflare-hosted) over opaque KV.

### Mutation friction

The `seen` signal is **not exposed as a tool**. Per the contract spec, `seen` is server-emitted as cards are delivered; it's not the AI agent's intent. `record_signals` covers the intentful subset (`save`, `share`, `dismiss`, `open`, `source_open`).

Other mutations get MCP tool annotations:
- `destructiveHint: false` is the default for additive operations (signals, feedback)
- `destructiveHint: true` for `update_preferences` (overwrites prior state) and `dismiss_feed_session`
- High-risk tools (`submit_item`, `submit_collection`) carry no `idempotentHint: true` (no auto-approve) — the MCP host should require per-call approval

The risk levels come from each operation's `x-mcp-mutation-risk` extension in the reconciled spec.

### Transparency surfacing

Wherever the underlying API response carries algorithmic-decision metadata — staleness verdicts, supersession pointers, significance breakdowns, search relevance, control-option confidence — the MCP server propagates it to `_meta.reasoning`. This is opt-in at the server level: when the API doesn't emit reasoning, the field is absent. The MCP server does **not** synthesize reasoning the underlying API did not provide.

This surfacing is the lever that distinguishes this product from a black-box search API. Customers building on the MCP server can route on `_meta.reasoning.confidence`, filter on `_meta.reasoning.rule_version`, or A/B-test ranking versions — all without parsing prose.

### Skills

Existing skills in `skills/*/SKILL.md` are kept, not replaced. MCP tool descriptions (in the spec) cover *what* each tool does; skills cover *how* AI agents should sequence tool calls and frame interactions with humans. The two layers are complementary, not redundant.

A small follow-up workstream updates each skill to enumerate the MCP tool names it suggests; that work is out of scope for this RFC.

---

## Tool catalog

| # | Tool name | Source | Phase | Risk | Owns? | Notes |
|---|---|---|---|---|---|---|
| 1 | `discover_streams` | `discoverStreams` | 1 | none | — | Catalog of public lanes |
| 2 | `get_stream` | `getPublicStream` | 1 | none | — | Anonymous-friendly |
| 3 | `get_categories` | `getCategories` | 1 | none | — | Anonymous-friendly |
| 4 | `get_topic` | `getTopic` | 1 | none | — | See open question 2 (topic ID model) |
| 5 | `search` | `search` | 1 | none | — | Mixed result types |
| 6 | `lookup_source` | `search?scope=sources` | 1 | none | — | See open question 8 (dedicated endpoint?) |
| 7 | `lookup_curator` | `getCurator` | 1 | none | — | |
| 8 | `get_card` | **needs new endpoint** | 1 | none | — | See open question 8 |
| 9 | `start_feed_session` | `startFeedSession` | 2 | low | yes | Returns opaque `session_id` |
| 10 | `get_feed_session` | `getFeedSession` | 2 | none | yes | Session-owner check |
| 11 | `continue_feed_session` | `continueFeedSession` | 2 | low | yes | Same session, more cards |
| 12 | `update_feed_controls` | `updateFeedSessionControls` | 2 | low | yes | See open question 3 (payload shape) |
| 13 | `dismiss_feed_session` | **needs new endpoint** | 2 | low | yes | See open question 8 |
| 14 | `start_radio_session` | `startRadioSession` | 2 | low | yes | State `ready`; resume to begin |
| 15 | `get_radio_session` | `getRadioSession` | 2 | none | yes | |
| 16 | `update_radio_session` | `updateRadioSession` | 2 | low | yes | resume / pause / skip / stop |
| 17 | `record_signals` | `recordSignals` | 3 | low | yes | Batched; `seen` not exposed |
| 18 | `get_replay` | `getReplay` | 3 | none | yes | Person token required |
| 19 | `get_activity_stats` | `getActivityStats` | 3 | none | yes | |
| 20 | `get_preferences` | `getPreferences` | 3 | none | yes | Cross-person 403 |
| 21 | `update_preferences` | `updatePreferences` | 3 | medium | yes | Carries durability caveat (see prerequisites) |
| 22 | `submit_item` | `submitItem` | 3 | high | yes | Per-call approval expected |
| 23 | `submit_collection` | `submitCollection` | 3 | high | yes | Per-item outcomes in response |
| 24 | `get_submission` | `getSubmission` | 3 | none | yes | Submitter or admin |
| 25 | `submit_feedback` | `submitFeedback` | 3 | low | yes | |

**Phase totals:** 8 in Phase 1, 8 in Phase 2, 9 in Phase 3.

The `Owns?` column flags tools whose authorization is keyed off bearer-principal ownership of the affected resource (the human-agent sense). Approval friction (the AI-agent sense) is governed independently by the MCP host's policy, informed by the `Risk` column.

---

## Phasing

### Phase 1 — Read-only discovery (8 tools)

Ships once the v0.3.0 spec is merged. No prerequisites. Anonymous-friendly tools work without a token. Person tokens unlock the same tools with rate-limit ceiling differences. This phase achieves Perigon parity.

Three of the 8 Phase 1 tools depend on resolutions:
- `get_topic` works today but with topic-id-as-headline ambiguity (open question 2)
- `lookup_source` is implementable as `search?scope=sources` today; consider promoting to a dedicated endpoint (open question 8)
- `get_card` requires a new server endpoint (open question 8)

If all three blockers are deferred, Phase 1 ships with 5 tools and the other 3 join when ready.

### Phase 2 — Sessions and controls (8 tools)

Adds the differentiating delivery layer: feed sessions, radio sessions, control updates. This is what distinguishes this product from a search-only MCP. Session state flows through opaque IDs returned in tool results.

Two Phase 2 tools depend on resolutions:
- `update_feed_controls` works with either the live flat shape or the legacy structured shape (open question 3)
- `dismiss_feed_session` requires a new server endpoint (open question 8)

### Phase 3 — Mutations (9 tools)

Adds signals, preferences, submissions, feedback, and replay/stats. Highest-risk tools live here; MCP host approval policy mediates per-call confirmation.

Two Phase 3 tools carry user-facing caveats until prerequisites land:
- `update_preferences` should display the persistence caveat in its description (only `topics_followed` is durably stored today; other fields accepted but discarded)
- `record_signals` and `submit_collection` should display the silent-skip caveat (server may drop entries; inspect `recorded` / `accepted` / `skipped` fields)

These caveats are honest acknowledgements of current behavior, not design choices we like. Once the server fixes land, the caveats can be removed.

---

## Hosting and deployment

- **Domain:** `mcp.chowdahh.com/v1/mcp`. The version path matches the REST API's `/api/v1/` convention.
- **Runtime:** Cloudflare Workers. Stateless edge function. No Durable Objects, no KV, no D1 — every request is a single fetch to the REST API.
- **Install path** for MCP hosts:
  ```
  claude mcp add --transport http chowdahh https://mcp.chowdahh.com/v1/mcp \
    --header "Authorization: Bearer <YOUR_CHOWDAHH_TOKEN>"
  ```
- **Rate limits** inherit from REST. The `X-RateLimit-*` response headers (formalized in the v0.3.0 spec) are passed through to MCP callers as `_meta` fields on tool results.
- **Logging:** request_id from each REST response surfaces in `_meta.request_id` for tracing. No request body or token is logged at the MCP layer.

---

## Prerequisites

Before any phase ships, these server-side items should be addressed (or explicitly accepted as known caveats):

1. **Sign-in flow repair.** The current sign-in flow at chowdahh.com returns "Unable to send code — try again," meaning no person tokens can be created through the documented user-facing path. Any Phase 2 / 3 tool requiring a person token is unusable until this is fixed.
2. **Idempotency-Key implementation.** Documented as planned-not-implemented. Until wired up, AI agents that retry on transient failures may create duplicate submissions, signals, or feedback. The header is currently marked `deprecated` in the spec; flip to live when the server-side support lands.
3. **Preference durability.** Only `topics_followed` is persisted today. Either (a) extend storage to cover the other fields the schema accepts, or (b) leave the spec's `deprecated` annotations in place and accept the user-facing caveat in `update_preferences`.
4. **Signal silent-skip surfacing.** The reconciled spec adds `recorded` and `skipped[]` fields to the signal-batch response. The server should populate these accurately; otherwise AI agents have no way to recover from invalid signals.
5. **Collection submission per-item outcomes.** Same shape concern: the spec defines `accepted`, `skipped`, and `results[]`; the server should populate them.

For Phase 1, only item 1 (sign-in) matters, and only for token-gated rate-limit ceilings — the Phase 1 tool surface is anonymous-friendly.

---

## Open questions

These need decisions before the corresponding tools / phases can ship cleanly. Most are pre-existing product or product-spec questions that the audit surfaced; none were created by this RFC.

1. **Sign-in flow repair.** Timeline?
2. **Topic ID model.** Stable opaque IDs, or headlines? `get_topic` works either way but the tool description and skill guidance change accordingly.
3. **Controls payload shape.** The reconciled spec accepts both the flat live shape and the structured legacy shape. Which is canonical going forward?
4. **Idempotency.** Wire it up server-side, or remove from spec?
5. **Curator capabilities.** Do `ch_cur_*` tokens unlock anything beyond rate quotas (e.g., curator-owned collections, attribution, extended quotas)? If yes, a follow-up RFC adds curator-only tools. If no, the bearer-token taxonomy can be simplified.
6. **Staleness reasoning surfacing.** The spec defines `Reasoning`, `StalenessVerdict`, and `SupersededBy` schemas. Will the server emit them on cards and topics? This is the defining transparency feature.
7. **Spec ↔ narrative reconciliation.** Should the OpenAPI spec be promoted to source-of-truth and the narrative contract demoted to companion docs? Avoiding future drift requires picking one.
8. **Three new endpoints to consider:**
   - `GET /api/v1/cards/{card_id}` — for `get_card`. Currently the MCP RFC's only Phase 1 tool with no underlying endpoint.
   - `DELETE /api/v1/feed-sessions/{session_id}` — for `dismiss_feed_session`. Cleaner than letting the 12-hour TTL handle it.
   - Whether `lookup_source` warrants a dedicated `GET /api/v1/sources/{source_id}` endpoint or remains a `search?scope=sources` query.

---

## What this RFC deliberately doesn't propose

- **A curator-only tool surface.** Pending open question 5.
- **Synthesized reasoning when the underlying API doesn't emit it.** The MCP server reflects the API; it doesn't invent.
- **Replacing skills with MCP prompts.** Skills survive in non-MCP contexts and are the right vehicle for behavioral guidance.
- **A new persistence layer at the MCP edge.** Stateless proxy is the design.
- **Generated SDKs.** Out of scope for this RFC; the reconciled spec is now SDK-generation-ready, but actually generating SDKs is a separate workstream.

---

## Changes required if approved

1. New top-level directory `mcp-server/` containing the Cloudflare Worker.
2. `mcp-server/wrangler.toml` for Worker configuration.
3. `mcp-server/src/server.ts` (or `.js`) — entry point implementing the JSON-RPC 2.0 protocol over HTTP Streamable.
4. `mcp-server/src/tools/` — one file per tool group (Discovery, Sessions, Radio, Signals, Submissions, Preferences, Feedback). Each file's tools generated from the spec via `x-mcp-*` extensions.
5. `mcp-server/README.md` — install instructions, local-dev guide, deployment.
6. CI: add a workflow that builds and validates the Worker.
7. After ship: small follow-up updating `skills/*/SKILL.md` to enumerate MCP tool names per skill.

The reconciled OpenAPI spec already includes everything else needed: tool names (`x-mcp-tool-name`), inclusion flags (`x-mcp-include`), phase assignment (`x-mcp-phase`), risk annotations (`x-mcp-mutation-risk`), and principal sense (`x-mcp-principal`).

---

## Validation

When sign-in is restored, the implementation can be validated by:

- The MCP test harness (e.g., `mcp-inspector`) walking each tool with realistic inputs.
- Round-trip tests: every Phase 1 tool returns shaped data matching `structuredContent`'s declared schema.
- Negative tests: cross-person preference reads return MCP errors mapping to the REST 403.
- Manual: install the server in Claude Code, ask "what's in the science stream?", verify a sensible response with structured `_meta.guidance`.

Until sign-in works, validation is limited to Phase 1's anonymous tools.
