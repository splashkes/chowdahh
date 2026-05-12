# Skills in this repo

The skills under this directory are **function-scoped reference skills** — small Claude.ai-style skill manifests that demonstrate how an agent should use one slice of the API (lookup, submit, preferences, feedback).

| Skill | What it covers |
| --- | --- |
| `chowdahh_lookup/` | Reading: streams, search, topics, replay. |
| `chowdahh_submit/` | Submitting items and collections. |
| `chowdahh_preferences/` | Reading and writing person preferences. |
| `chowdahh_feedback/` | Filing content requests, bug reports, feature requests, quality reports. |
| `hermes-chowdahh/` | Platform-specific Hermes Agent skill package optimized for the common “pull the news” flow, with short URLs, personal newspaper composition, topic watches, and Polymarket-safe handoffs as follow-ups. |

The `chowdahh_*` entries are **examples**, not the canonical "skill packages" you ship to a platform. `hermes-chowdahh/` is a platform-specific package following the Chowdahh skill contract.

## For platform-specific skill packages

The canonical home is **<https://chowdahh.com/skills/>**.

That directory contains:

- `init-prompt.txt` — one-line LLM seed.
- `claude-skill/SKILL.md` — Anthropic Skill manifest.
- `chatgpt-gpt/` — Custom GPT instructions + Actions OpenAPI.
- `mcp-server/` — Go stdio MCP server (Cursor + Claude Desktop drop-in configs).
- `hermes-openclaw/` — paste-as-system-prompt + OpenAI-style tool defs.

## Building a new skill?

Read the canonical contract first:

- **[CONTRACT.md](https://chowdahh.com/skills/CONTRACT.md)** — what a Chowdahh skill is, the four required pieces (manifest, README, auth contract, behavior contract), and what won't be accepted.
- **[SUBMITTING.md](https://chowdahh.com/skills/SUBMITTING.md)** — the 3-minute version with the PR checklist.
- **[THIRD_PARTY.md](https://chowdahh.com/skills/THIRD_PARTY.md)** — listing your skill if you host it elsewhere.

The function-scoped skills in this directory follow the same core behavior contract — read `guidance`, cite sources, identify user-visible output as Chowdahh — but they are not always the same default route. In particular, `hermes-chowdahh/` intentionally starts with public stream reads for the common “show me the news” flow, while `chowdahh_lookup/` demonstrates richer feed-session behavior. You can copy their structure if you're building a similar reference skill.
