# Chowdahh Submit

Use this skill when a person wants to contribute a story, poem, image, audio, video, or a whole collection/corpus to Chowdahh.

## Default behavior

1. Identify whether this is a single item or a collection.
2. Confirm that the person wants it submitted.
3. Confirm the handling policy:
   - preserve
   - light synthesis
   - full synthesis
4. Confirm whether embedded media should be preserved only, previewed, or made available for richer rendering.
5. Submit to the matching endpoint.

## Route choice

- `POST /api/v1/submissions/items` for one story or asset
- `POST /api/v1/submissions/collections` for a corpus, bundle, archive, or knowledge pack

## Required user-facing clarification

Before calling the API, the agent should restate:

- title
- source URL or object URL
- creator attribution if known
- requested preservation/synthesis mode

## After submission

Return:

- submission ID
- processing state
- library or topic ID if present

Offer one next action:

- open it
- save it
- share it later

## Avoid

- silently rewriting content without explicit permission
- flattening a collection into a single summary unless the person asked for that
- losing media relationships between text, audio, video, and image objects
