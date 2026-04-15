# Core Model

This repo uses a smaller, intent-shaped model.

## 1. `person`

The human Chowdahh is serving.

Holds durable product state such as:

- followed topics
- avoided topics
- tone preferences
- default delivery settings

## 2. `feed_session`

The main unit of delivery.

A feed session exists so an agent can:

- send content now
- send more
- apply or remove controls
- explain why the current set was chosen

## 3. `card`

The atomic thing the person sees.

A card may point to:

- a topic
- a source-backed item
- a synthesized view
- a submitted collection entry

## 4. `control_state`

The machine-usable mirror of the product toggle interface.

It includes:

- available controls
- currently selected controls
- counts or confidence
- safe defaults

## 5. `replay_event`

The history record of what happened to cards.

Examples:

- seen
- open
- save
- share
- dismiss

Replay is built from these events.

## 6. `radio_session`

Separate from replay.

Represents audio or voice-first delivery mode, with its own controls and progress.

## 7. `submission`

Represents content brought into Chowdahh by a person or agent.

Can be:

- a single item
- a collection
- a corpus

## 8. `feedback`

Represents a person- or agent-originated request back into the system.

Types:

- content request
- bug report
- feature request
- quality report
