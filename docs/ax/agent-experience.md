# Agent Experience

## Intended Feeling

An agent should feel like it has a clean companion system, not a maze of product internals.

The main intents should feel obvious:

1. send content now
2. send more
3. replay what already happened
4. start radio
5. submit content
6. send feedback

## Default Agent Flow

### For feed delivery

1. Call `POST /api/v1/feed-sessions` with minimal assumptions.
2. Surface the returned controls back to the person in plain language.
3. If the person leans one way, update local memory first.
4. Only sync durable preferences to Chowdahh if the person confirms they want that to stick.
5. Use `send more` instead of starting a new session unless the intent changed.

### For replay/history

1. Call `GET /api/v1/replay`.
2. Explain the result as card history, not as analytics jargon.
3. Use stats only when the person wants totals or summaries.

### For saving or sharing

1. Record the intent in the conversation.
2. Call `POST /api/v1/signals` after the action actually happens.
3. Offer stats or replay later using the recorded product state.

### For radio

1. Call `POST /api/v1/radio-sessions` with a `mode` and `duration_minutes`.
2. Read `guidance.next_best_actions` for available controls (play, skip, pause, stop).
3. Control the session via `PATCH /api/v1/radio-sessions/{id}` with an `action`.
4. There are no per-card audio endpoints. Radio is session-based — the server builds a queue from current content.
5. Keep radio language separate from replay/history language.

### For submitting content

1. Confirm ownership or authority to submit.
2. Confirm preservation vs synthesis.
3. Submit via `items` or `collections`.
4. Offer later retrieval once accepted.

### For feedback

1. Classify whether this is a content request, bug report, feature request, or quality report.
2. Send it through one feedback surface.

## AX Requirements

- Agents should not need to understand internal ranking stages.
- Every response should contain enough explanation for a human-facing restatement.
- Controls and lenses should come back with counts or confidence so the agent can discuss them honestly.
- Status transitions should be explicit: `queued`, `processing`, `ready`, `failed`.
- Feed sessions should be resumable enough that `send more` feels natural.

## Preferred API Tone

- concise
- inspectable
- attribution-rich
- no magic wording like "we found something perfect for you"

Instead:

> "Here are 5 cards. I kept it balanced and current, and I can tilt it toward science, good news, or local topics."

## What The Agent Should Never Need To Guess

- whether content was preserved or rewritten
- whether an item came from an original source or Chowdahh synthesis
- whether a control chip is real
- whether a preference is local-only or synced server-side
