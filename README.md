![Deployment](https://github.com/kindeep/treachery-angular/workflows/Firebase%20Deployment/badge.svg?branch=master)

# Treachery Web App

https://treacheryonline.web.app

A web app for a board game based on Deception: Murder in Hong Kong

# Gameplay

## Forensic Scientist

1. **Create a game**: Go to https://treacheryonline.web.app and click "Create Game"

![1](screenshots/forensic/0.png)

2. **Invite your friends to join the game**: Either copy the room link on the start screen, or share the game code which friends can enter on the home page. Anyone opening a started-game link now joins as a read-only observer unless they are already in the room.

![1](screenshots/forensic/1.png)

3. **Set up the lobby**: The creator is a normal player and moderator. Before the game starts, the creator can toggle any participant between **player** and **observer**, toggle-mark one player as the forensic scientist, configure how many **means** and **evidence/clue** cards each suspect receives, and optionally add **accomplices** and **witnesses**. Evidence can follow the means count automatically, or be set separately. Witnesses and accomplices default to `0`.

![1](screenshots/forensic/2.png)

4. **Start the game**: You now need at least four lobby players total so one can become the forensic scientist and three suspects remain. When the game starts, the selected forensic scientist stays visible in the room roster but does not receive cards, cannot guess, and cannot chat during the active game.

5. **Wait for murderer to select their cards**: The game distributes cards among the remaining suspect players randomly using the configured lobby counts. One suspect is selected to be the murderer. Optional accomplices and witnesses are also assigned randomly. The murderer is prompted to select one clue card and one means card from the cards they were dealt. These prompts can now be minimized while the player looks at the board.

![1](screenshots/forensic/3.png)

6. **Round 1 of clues**: You will be presented with cards that you can use to give clues to the Investigators to figure out who the murderer is. For each card, select one of the options and click "Select \_\_\_ card" to finalize your selection. The first card, the Cause of death card is fixed. For the second card, you get a choice among different Location cards. Choose whichever one you think would fit best. For the rest of the cards, they're selected randomly. In the first round, you should select a total of 6 cards. After selection is complete, the creator can start the built-in shared room timer (default `0:40`, but any positive whole-second duration works) for Investigators to discuss among themselves. Ideally, Investigators should be on a voice call, but you may also use the text chat function provided by the app. (You're not allowed to reveal who the murderer is, or any clues other than the cards you selected.). After the discussion timer is over, each Investigator gets 30 secs uninterrupted to speak their case for who they think it is and why. After this is over, we move on to round 2.

![1](screenshots/forensic/4.png)

7. **Round 2 & 3 of clues**: You will now randomly be dealt another card, which you can use to replace one of the existing cards. After making this replacement, the creator can restart, pause, resume, reset, or clear the shared room timer for the next discussion window. Repeat this for round 3.

![1](screenshots/forensic/5.png)

8. **Ending the game**: If no one correctly guesses who the murderer is before the end of round 3, the murderer team wins. If the investigators guess correctly and there are no witnesses in the lobby settings, the investigator team wins immediately. If witnesses are enabled, the game enters a final witness-selection step for the murderer team instead. The creator gets a blind suspect list they can use to quietly re-show the prompt without learning the murderer team, and any creator-issued accomplice prompt can be dismissed. The first submitted witness selection ends the game immediately, and the finished screen now reveals every hidden role.

![1](screenshots/forensic/6.png)

## Player

1. **Joining the game**: Either use the shared room link or enter the room code on the home page. The first time you join, create, or observe a room, the app asks for a display name and then saves it to your local auth session so you are not prompted again. Future joins reuse that saved identity automatically.

![1](screenshots/investigator/0.png)

2. **Lobby**: You can see every room participant in the lobby, including observers. Players can use chat in the lobby. The creator can switch participants between player and observer, mark one player as the forensic scientist, choose card counts, and configure optional accomplices and witnesses before the game starts. The lobby still needs at least 4 players before the game can start. After the game starts, the chosen forensic scientist is either the marked player or a random player if nobody was marked.

![1](screenshots/investigator/1.png)

## Investigator

1. **Start screen**: After the forensic scientist starts the game, you will see that the murderer is selecting their cards. You can see the suspect players and their cards here. One of the other suspects is the murderer and is selecting one Means card and one Clue card right now. Your objective along with the other investigators is to find out who the murderer is along with correctly guessing both their cards. Your forensic scientist will give out clues for you to figure this out.

![1](screenshots/investigator/2.png)

2. **Clues**: Once the murderer has selected their cards, the Forensic scientist will start selecting clues to point to who the murderer is. You may discuss amongst yourselves to what you think the clues mean and who they point to. The clues your Forensic Scientist select will appear at the top of your screen.

![1](screenshots/investigator/3.png)

3. **Making a guess**: At any point in the game, after the murderer has selected their cards, you may submit your guess to who the murderer is along with the clue card and means card they selected. To do this, select a clue card and a means card for one of the players by clicking the two cards. You will see a guess composer below the players list. If you're sure you want to guess, click "Make guess". Keep in mind that you only get one guess per game, so use your guess wisely.

![1](screenshots/investigator/4.png)

4. **Round 1**: Your forensic scientist will initially select 6 cards, here Round 1 begins and the Forensic Scientist gives you a set time to discuss amongst yourselves to who you think the murderer is. Use this time wisely. At the end of the discussion time, you will all get 30 seconds uninterrupted to present your case for who you think the murderer is, or why you're not the murderer. The shared room timer stays pinned near the top of supported game views so mobile players can keep it in view while scrolling. Round 1 ends here.

![1](screenshots/investigator/5.png)

5. **Round 2 & 3**: For Round 2, the forensic Scientist will replace one of the cards. Everything else about this round is same as the previous round, and the moderator can reuse the shared room timer for each new discussion window. Round 3 is similar to round 2, Forensic scientist will replace another card. You must submit your guess before the end of Round 3.

![1](screenshots/investigator/6.png)

## Murderer

1. **You are the murderer**: You will be prompted to select a clue card and a means card. The forensic scientist will give out clues based on these selections, so select wisely. Once the room starts, the forensic scientist does not get cards; only the suspect players do.

![1](screenshots/murderer/0.png)

2. **Gameplay**: Throughout the rounds, your role is the same as the Investigators. You have to pose as one, and make sure no one suspects you. You can make one guess just like everyone else. If by the end of Round 3, no one has guessed correctly or guesses have expired for everybody else, you win!

![1](screenshots/murderer/1.png)

## Optional accomplices and witnesses

- **Accomplices** know the full murderer team, including the murderer and the other accomplices.
- **Witnesses** know the murderer team, but the murderer team does not know which players are witnesses.
- If a correct accusation happens while witnesses are enabled, the murderer team gets one final witness-selection action before the case is resolved.

## Workflow

- Pushes to `master` go to production at https://treacheryonline.web.app

## Identity and device migration

- Each browser keeps a local auth token in local storage.
- Your saved display name is attached to that auth identity, so you only need to enter it once.
- Any room participant can use **Migrate device** to copy a room-specific link that reuses that room identity on another device without exposing the real auth token.
- The migration link is room-scoped and stays in the URL so refresh keeps the imported identity for that room.
- Opening a started-game room link as someone who is not already in the room joins as an observer automatically or via the observe prompt.
