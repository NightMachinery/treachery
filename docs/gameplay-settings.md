# Treachery gameplay settings and hidden roles

This document covers the pre-start lobby settings and the optional hidden-role flow.

## Lobby settings

Before the game starts, the creator can configure:

- means cards per suspect
- evidence/clue cards per suspect
- a checked-by-default toggle that makes evidence count follow means count
- accomplice count (`0-10`)
- witness count (`0-10`)
- how many witnesses the murderer team must identify to win the witness reversal

Defaults:

- means cards per suspect: `4`
- evidence/clue cards per suspect: `4`
- evidence follows means: `true`
- accomplices: `0`
- witnesses: `0`
- witnesses to find: `0`

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
- the creator can also show the same prompt to any accomplice on the murderer team
- creator-issued prompts are dismissible
- all prompts are minimizable
- the first submitted witness selection ends the game immediately

If every selected player is actually a witness, the murderer team wins. Otherwise the investigator team wins.

## Finished-game reveal

Once the game ends, the UI shows:

- winner/result message
- the murderer and murder cards
- a full role reveal for everyone in the room
