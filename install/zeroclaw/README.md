# ZeroClaw / OpenClaw Install

These instructions are written against the local ZeroClaw docs available in this environment, which document `zeroclaw skills install`, `zeroclaw skills audit`, and trusted local skill roots.

## Option 1: install individual skills from a local clone

```bash
git clone https://github.com/splashkes/chowdahh_recipes.git

zeroclaw skills audit /path/to/chowdahh_recipes/skills/chowdahh_lookup
zeroclaw skills install /path/to/chowdahh_recipes/skills/chowdahh_lookup

zeroclaw skills audit /path/to/chowdahh_recipes/skills/chowdahh_preferences
zeroclaw skills install /path/to/chowdahh_recipes/skills/chowdahh_preferences

zeroclaw skills audit /path/to/chowdahh_recipes/skills/chowdahh_feedback
zeroclaw skills install /path/to/chowdahh_recipes/skills/chowdahh_feedback

zeroclaw skills audit /path/to/chowdahh_recipes/skills/chowdahh_submit
zeroclaw skills install /path/to/chowdahh_recipes/skills/chowdahh_submit
```

## Option 2: keep skills in place with trusted roots

Add the repo root to trusted skill roots in `~/.zeroclaw/config.toml`:

```toml
[skills]
trusted_skill_roots = ["/absolute/path/to/chowdahh_recipes/skills"]
```

Then install from the local path as above, or manage symlinks inside `~/.zeroclaw/workspace/skills/` if you want a shared checkout.

## Why local-path install is the recommended first pass

This repo contains multiple skill directories, not one monolithic skill. Installing each skill separately is clearer and keeps the boundary between lookup, preference capture, feedback, and submission explicit.

## Recommended first skill order

1. `chowdahh_lookup`
2. `chowdahh_preferences`
3. `chowdahh_feedback`
4. `chowdahh_submit`

That mirrors the normal adoption path:

1. read
2. personalize
3. steer or correct
4. contribute
