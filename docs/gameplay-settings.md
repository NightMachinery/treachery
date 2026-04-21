# Treachery gameplay settings and hidden roles

This document covers the pre-start lobby settings and the optional hidden-role flow.

## Lobby settings

Before the game starts, the creator can configure:

- means cards per suspect
- evidence/clue cards per suspect
- a checked-by-default toggle that makes evidence count follow means count
- a saved toggle that makes means/clue cards text-only and hides their images
- accomplice count (`0-10`)
- witness count (`0-10`)
- how many witnesses the murderer team must identify to win the witness reversal

Defaults:

- means cards per suspect: `4`
- evidence/clue cards per suspect: `4`
- evidence follows means: `true`
- means/clues text only: `false`
- accomplices: `0`
- witnesses: `0`
- witnesses to find: `0`

The lobby uses the existing **Save settings** flow, so pre-start changes to this toggle do not apply until saved.

## Means/clues display mode

- Means/clue cards support two room-wide display modes:
  - default image cards
  - text-only cards that hide the means/clue images
- After the game starts, the creator gets a live **Room mods** toggle for this setting on room screens.
- Mid-game room-mod changes apply immediately to everyone in the room, including observers and the forensic scientist.
- Existing suspect card images stay mounted when new forensic hints arrive, so revealing another hint does not reload the card art.
- Only means/clue cards are affected; forensic clue cards stay unchanged.
- Means/clue card labels now use browser auto-direction (`dir="auto"`) so RTL text renders correctly in both image and text-only modes.
- Means/clue cards no longer repeat the words `means` or `clue` on every card; the card content stays prominent while color accents still distinguish the two decks.

## In-game room layout

- Active game screens now group the play table into clear sections: room controls, forensic clues, suspect decks, guess composition, the viewer's private hand, and chat.
- Suspect decks now reflow into responsive cards instead of forcing a wide horizontal strip, so desktop and mobile layouts stay readable when comparing players.
- Chat is presented as a dedicated side panel on wide screens and stacks underneath the board on smaller screens.
- On mobile, the chat composer stays pinned to the bottom of the chat panel with safe-area padding so the text input remains visible.
- Creator action buttons such as **Migrate device** and **End game** are centered in their action groups, and **End game** now uses a dialog confirmation to reduce accidental taps.
- The accusation panel now shows guided next steps before a full accusation is ready, so the section stays useful instead of looking empty between selections.
- The same layout applies in both image-card mode and text-only means/clues mode.

## Shared room timer

- Once the game has started, the creator can launch a shared room timer from the active room view.
- The timer defaults to `40` seconds, but the creator can enter any positive whole-second duration.
- Timer controls are creator-only and support:
  - start a new timer
  - pause
  - resume
  - reset back to the last-set duration and immediately restart
  - clear
- Everyone in the room sees the same sticky timer banner on game screens, including mobile.
- When the timer expires, it stays visible at `0:00` until the creator restarts or clears it, with a visual expiry cue and a best-effort sound cue in supported browsers.

## Role assignment

When the creator starts the game:

1. one lobby player becomes the forensic scientist (marked player first, otherwise random)
2. the remaining suspects receive the configured clue/evidence and means counts
3. secret suspect roles are assigned randomly:
   - exactly 1 murderer
   - the configured number of accomplices
   - the configured number of witnesses
   - everyone else is an investigator

## Private information

- murderer: knows the full murderer team and selects the murder cards
- accomplice: knows the full murderer team, including the other accomplices, and sees the murder cards once selected
- witness: knows the murderer team identities
- forensic scientist: knows the murderer and the selected murder cards through the existing forensic private view

## Witness-selection resolution

If witnesses are disabled, a correct accusation ends the game immediately in favor of the investigator team.

If witnesses are enabled, a correct accusation opens a final witness-selection phase instead:

- the murderer automatically gets a non-dismissible witness-selection prompt
- the creator gets a blind suspect list so they can re-show the same prompt without learning who is on the murderer team
- only accomplices receive creator-issued prompts; selecting any other suspect quietly does nothing
- creator-issued accomplice prompts are dismissible
- all prompts are minimizable
- the first submitted witness selection ends the game immediately

If every selected player is actually a witness, the murderer team wins. Otherwise the investigator team wins.

## Finished-game reveal

Once the game ends, the UI shows:

- winner/result message
- the murderer and murder cards
- a full role reveal for everyone in the room
- a creator-only **Play again** confirmation that resets the same room back to the lobby while keeping the room URL, current participant roster, participant roles, and saved settings

## Guess restrictions

- Every suspect still gets only one submitted guess per game.
- Exact duplicate full guesses are rejected room-wide, so the same suspect + clue + means combination cannot be submitted twice.
