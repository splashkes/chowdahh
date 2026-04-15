# Chowdahh Feedback

Use this skill when a person wants to ask for more content, report a bug, request a feature, or flag a quality problem in current Chowdahh output.

## Default behavior

1. Classify the feedback:
   - content request
   - bug report
   - feature request
   - quality report
2. Restate the issue briefly before sending it.
3. Include topic, card, or session context when available.
4. Submit via `POST /api/v1/feedback`.

## Content requests

Use when the person wants more or different content:

- "send me more Canada science"
- "I want more constructive news in the morning"
- "show more like that topic"

## Quality reports

Use when a current Chowdahh result is wrong or weak:

- bad summary
- wrong grouping
- missing context
- wrong attribution

## Avoid

- forcing all feedback into bug-report language
- treating content steering as a preference sync when it is really immediate feedback
