# Wordpacks

Treachery now loads gameplay content from filesystem-backed wordpacks under `./wordpacks/`.

## Pack split

- `wordpacks/crime/<pack>/` — **CrimePack**
  - means cards
  - clue cards
  - image asset sets
- `wordpacks/hint/<pack>/` — **HintPack**
  - forensic/cause/location/other hint cards

Rooms choose one CrimePack and one HintPack before the game starts.

## File formats

The format stays intentionally simple:

- newline-separated text for flat ordered label lists
- `JSON5` for metadata and nested hint-card structures
- `JSONL` is supported by the loader when record-per-line data is cleaner than parallel text files

## Default packs

- CrimePack: `wordpacks/crime/treachery`
- HintPack: `wordpacks/hint/treachery-hints`

Both ship with:

- English (`en`)
- Persian (`fa`)
- Treachery also includes a selectable in-progress art pilot asset set: `storybook-cel`

## CrimePack layout

Example:

```text
wordpacks/crime/treachery/
  pack.json5
  means/
    ids.txt
    languages/
      en.txt
      fa.txt
  clues/
    ids.txt
    languages/
      en.txt
      fa.txt
  assets/
    treachery/
      means/
        001-alcohol.jpg
        default.jpg
      clues/
        001-air-conditioning.jpg
        default.jpg
```

Image lookup falls back from the selected asset set to the pack fallback asset set, then to deck-default art, and finally to a no-image text card.

## Card image framing

- Gameplay cards currently render images with centered cover-crop:
  - `object-fit: cover`
  - `object-position: center`
- The new `storybook-cel` pilot art is mastered at **7:10 portrait** (`1400x2000`) to match the current card slot best across layouts.
- Because the card title overlay sits near the bottom edge, important subject detail should stay centered and slightly above center.

## Storybook Cel pilot

- CrimePack: `treachery`
- Asset set id: `storybook-cel`
- Display name: `Storybook Cel`
- Default/fallback behavior is unchanged:
  - default asset set stays `treachery`
  - fallback asset set stays `treachery`
- Current generated pilot cards:
  - means: `006-bamboo-tip`, `007-bat`, `025-dumbbell`, `029-explosives`, `037-kerosene`, `042-locked-room`, `055-pistol`, `057-plastic-bag`, `064-radiation`, `070-smoke`, `081-unarmed`, `088-wire`
  - clues: `015-briefs`, `019-cake`, `065-flyer`, `074-handcuffs`, `093-juice`
- Cards not yet present in `storybook-cel` fall back to the existing `treachery` art automatically.
- See `docs/wordpacks-storybook-cel-pilot.md` for the pilot art direction and prompt template.

## HintPack layout

Example:

```text
wordpacks/hint/treachery-hints/
  pack.json5
  languages/
    en.json5
    fa.json5
```

Hint language files must keep the same card IDs and choice IDs across languages.
