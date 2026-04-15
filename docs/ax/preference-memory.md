# Preference Memory

## Split The State

### Keep in local agent memory

- emotional context
- temporary missions
- family or team context
- phrasing preferences
- unconfirmed dislikes
- high-context notes like "avoid heavy war coverage in the morning"

### Sync to Chowdahh

- topics followed
- topics muted
- default delivery mode
- default budget
- source preferences
- tone lenses the person explicitly wants to persist

## Suggested Conversation Pattern

The agent should ask narrow questions with visible effect.

### Step 1: topic appetite

> "What do you want more of by default: science, world, business, culture, local, or good news?"

### Step 2: topic avoidance

> "Anything you want de-emphasized unless it becomes important?"

### Step 3: delivery style

> "Should I usually keep Chowdahh brief, balanced, and source-heavy, or more narrative?"

### Step 4: controls boundary

> "Should I usually choose the control chips for you, or ask before I tilt the feed?"

### Step 5: persistence boundary

> "Do you want me to remember that just for our conversations, or save it into Chowdahh too?"

That last question is important. It prevents accidental preference sync.

## Recommended Local Memory Format

```json
{
  "chowdahh_profile": {
    "default_goal": "brief useful update",
    "morning_bias": ["uplifting", "science"],
    "avoid_until_requested": ["celebrity-gossip", "graphic-crime"],
    "source_trust_notes": ["prefers original reporting"]
  }
}
```

## Recommended Chowdahh Profile Format

```json
{
  "topics_followed": ["science", "canada"],
  "topics_avoided": ["celebrity-gossip"],
  "tone_preferences": ["uplifting", "grounded"],
  "delivery_preferences": {
    "default_budget_minutes": 8,
    "default_delivery_mode": "brief"
  }
}
```

## Practical Rule

If a preference is:

- subtle
- situational
- private
- emotionally loaded

keep it local until the person explicitly asks for product-level persistence.
