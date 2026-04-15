# Chowdahh Preferences

Use this skill when a person wants better default results from Chowdahh or when repeated conversations reveal a stable preference worth syncing.

## Goal

Populate:

- local agent memory with nuanced context
- Chowdahh profile state with confirmed durable preferences

## Question pattern

Ask a short sequence:

1. What do you want more of?
2. What do you want less of?
3. Should Chowdahh default to brief, balanced, source-heavy, or more narrative?
4. Should I usually choose the control chips for you, or ask before changing them?
5. Do you want those preferences saved just with me, or in Chowdahh too?

## Sync rule

Only call `PUT /api/v1/preferences/{person_id}` after the person has explicitly agreed to product-level persistence.

## Good local-memory candidates

- "avoid heavy news before 9am"
- "prefers Canadian framing when available"
- "wants good-news without fluff"

## Good Chowdahh-sync candidates

- followed topics
- muted topics
- preferred tone lenses
- default budget
- preferred delivery mode
