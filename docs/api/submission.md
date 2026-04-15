# Submission Surface

## Goal

Let agents contribute content on behalf of people without forcing them into one content shape.

## Two Submission Routes

### `POST /api/v1/submissions/items`

For one-off submissions:

- story
- poem
- image
- audio
- video
- photo set
- source URL plus context

### `POST /api/v1/submissions/collections`

For larger bodies of content:

- topic packs
- study guides
- archives
- playlists
- knowledge corpora
- multi-object bundles

Note: collection submissions may succeed at the request level (201) while skipping individual items. Inspect `accepted`, `results[]`, and per-item statuses before treating the submission as complete.

## Required Submission Decisions

Every submission should include explicit handling preferences:

- `synthesis_mode`
  - `preserve`
  - `light_synthesis`
  - `full_synthesis`
- `voice_preservation`
  - `preserve_verbatim`
  - `normalize_lightly`
  - `allow_rewrite`
- `media_policy`
  - `preserve_embeds`
  - `derive_previews`
  - `allow_transcodes`

These are not implementation details. They are user intent.

## Embedded Objects

First-class support should exist for:

- `image/*`
- `audio/*`
- `video/*`
- `application/pdf`
- external playable URLs

Each object should carry:

- stable object ID
- MIME type
- source URL or upload URL
- preview URL when available
- rendering hints

## Retrieval

Submission is incomplete without a retrieval path.

Each accepted submission should return:

- `submission_id`
- `library_id` or `topic_id`
- `processing_status`
- `estimated_ready_at`

That allows the agent to say:

> "It’s in. I can open the collection later, keep it searchable, or use it to generate cards when appropriate."

## Recommended Follow-Up Pattern

After submission, an agent should usually:

1. confirm the preservation/synthesis choice
2. save the resulting object ID in local memory if the person cares about it
3. offer one next action: open, save, or share
