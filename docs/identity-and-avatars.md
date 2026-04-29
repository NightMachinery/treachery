# Treachery identity and avatar fallback

Treachery stores a local anonymous auth identity per browser and keeps the chosen display name attached to that identity.

## Generated fallback avatars

- Participants without a custom avatar use a generated SVG avatar, including observers.
- The generator is deterministic, so the same identity data always renders the same avatar.
- The UI now renders those SVGs through the normal pnpm package `@nice-avatar-svg/preact`, with a small local wrapper that hashes the seed into a stable config.
- The seed is built with the user ID prepended to the display name:
  - `"<uid>:<displayName>"` when both exist
  - `uid` when only the user ID exists
  - `displayName` when only the display name exists
  - `anonymous` when neither exists

Why this matters:

- two different users who both pick `Alex` still get different avatars because their IDs differ
- if a player renames themselves, their avatar can change because the display name is part of the seed

## Current scope

- This repo currently uses generated fallback avatars only.
- Observers use the same generated fallback as players in participant lists, chat attribution, and the navbar identity display.
- There is no persisted custom-avatar upload/storage flow yet.
- The fallback is entirely local to the UI and does not require any backend or schema changes.
