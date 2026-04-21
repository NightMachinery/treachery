# UI/UX polish notes

## April 21, 2026 pass

This pass focused on the main interaction pain points that showed up in live Chrome debugging:

- replaced the native browser display-name `window.prompt()` with an in-app dialog
- removed native-prompt fallback for migrate-device links and switched to programmatic copy fallback instead
- upgraded lobby and room-timer number inputs with explicit decrement/increment controls
- tightened the text-only means/clues cards so they are easier to scan at a glance
- centered and padded card labels more aggressively so text stays clear near rounded corners
- fixed forensic option buttons so long labels like `Suffocation` wrap/fit inside the selected chip instead of overflowing
- hid the empty accusation helper panel until a player actually starts composing a guess
- removed redundant per-row means/clues labels inside suspect decks, added stronger row styling, centered card wrapping, and improved image-card text readability with a translucent title plate
- widened suspect-deck image cards, prevented ugly mid-word wrapping, and converted the rows to compact centered grids to reduce wasted empty space
- made chat start collapsed by default and added an accusation history section after the player's private cards
- strengthened selected-card states and de-emphasized non-selected suspect cards once a guess target is chosen
- made clicking an already-selected suspect card toggle it off again and added a cancel action to the accusation panel
- made the accusation panel show partial selections, with placeholders for whichever means/clue card is still missing
- reduced ugly forensic-option wrapping by widening clue cards and using tighter, more balanced option label typography
- changed automatic bad-team victory logic so accomplices no longer block the endgame once every good-team guess has been spent
- polished the finished-state UI for both player and forensic views with clearer winner banners, round metrics, and a final accusation log
- redesigned the desktop shared room timer so the countdown, duration stepper, and moderator actions sit in a tighter two-column control surface instead of leaving a large empty slab
- upgraded the shared room timer again into a three-part control deck with a richer summary card, clearer duration editor, and a dedicated moderator action panel
- restyled the live room-mods panel into a proper settings card with a custom toggle, clearer status badge, and stronger hierarchy for witness prompt follow-ups
- normalized spacing tokens used across the UI so card/panel spacing is consistent

## Verified flows

Using the remote Chrome workflow in `docs/chrome/README.md`:

- desktop home page: creating a room now opens an in-app display-name dialog instead of a browser prompt
- desktop lobby: lobby settings steppers work and update values immediately
- desktop live game: room-timer stepper controls render and work for the moderator
- desktop live game: redesigned timer + room-mod control strip renders cleanly for the moderator in a seeded debug room
- mobile text-only game: text-only means/clues cards remain readable without console errors
- mobile live game: redesigned timer stack stays readable and touch-friendly in iPhone-sized emulation

## Validation notes

- Verified in Chrome MCP against seeded debug room `BEJI` on both desktop and mobile emulation.
- Hot reload succeeded and the browser console stayed free of new runtime/template errors during verification.
- I did **not** run a fresh Angular build on the VPS for this pass because the manual preflight failed the documented threshold for `/proc/pressure/io` full `avg10` (measured `5.15`, required `<= 4.0`), so starting a new Node build would have risked swap/disk thrash on this host.

## Follow-up ideas

- add visual density options for very large suspect decks
- consider a dedicated modal or toast for migrate-device links when copy fails entirely
- consider screenshot-based regression checks for the lobby, game, and forensic layouts
