# Player count

For a single game, this repo currently supports:

- **Minimum:** 4 lobby players to start (`server/app/store.go`), which becomes **1 forensic scientist + 3 suspects**.
- **Practical maximum:** **23 lobby players total** = **1 forensic scientist + 22 suspects**.

Why 23 is the cap for the default Treachery CrimePack:

- Each suspect is dealt **4 means cards** and **4 clue cards**.
- The default `wordpacks/crime/treachery` pack has **90 means cards** and **200 clue cards**.
- Means cards are the limiting factor: `floor(90 / 4) = 22` suspects, plus 1 scientist = **23 players**.

Notes:

- **Observers are separate** from players and do not count toward this number.
- The server validates against the currently selected **CrimePack**, so a different pack can raise or lower the maximum.
