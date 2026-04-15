# Principles

## 1. Design for the agent's first question

Most agents will start with one of three asks:

- "Send content for this person right now."
- "Send more."
- "Show me what they already saw or acted on."

The repo is organized around these asks, not around internal pipeline stages.

## 2. Preserve honest toggles

If the interface exposes a lens like `good-news`, `local`, `deep-tech`, or `balanced`, the backend must actually rank or filter with it. No placebo toggles.

## 3. Mirror the control surface structurally

Agents need access to the same shape of controls the product has, but in a machine-usable form:

- control groups
- available chips
- defaults
- counts
- confidence

## 4. Keep the surface small

An agent does better with:

- one high-frequency feed-session start call
- one send-more call
- one control-update call
- one deep topic call
- one replay call
- one radio call
- one preference sync call
- one signal write call
- two submission calls
- one feedback call

than with twenty specialized endpoints it must learn through trial and error.

## 5. Separate memory from profile

The agent should keep nuanced private context in local memory:

- emotional sensitivities
- temporary missions
- family or team context
- phrasing preferences
- unconfirmed dislikes

Chowdahh should only hold confirmed, product-relevant preferences:

- topics followed or avoided
- delivery defaults
- source preferences

## 6. Preserve source integrity

Every surfaced item should retain:

- original source URL
- original creator or curator attribution
- transformation status
- explanation of what was synthesized, if anything

## 7. Treat ingest as a contract

Submission is not just upload. The submitter needs to state whether Chowdahh may:

- leave content untouched
- summarize it for discoverability
- synthesize it into new feed cards
- derive previews from media

## 8. Separate replay from radio

Replay means:

- prior cards
- prior signals
- prior interactions

Radio means:

- audio delivery mode
- queue/progress/state

They should never be collapsed into one concept.

## 9. Feedback is part of the loop

An agent should be able to ask:

- what did we save this week
- what did we share this month
- what did they already see
- can I request more content on this topic
- can I report a bug
- can I file a feature request

without needing a separate analytics product.
