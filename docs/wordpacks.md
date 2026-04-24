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
- Treachery also includes selectable supplemental art asset sets: `gouache-treachery` and `storybook-cel`

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
- The `storybook-cel` and `gouache-treachery` art is mastered at **7:10 portrait** (`1050x1500`, with a few-pixel tolerance) to match the current card slot best across layouts.
- Because the card title overlay sits near the bottom edge, important subject detail should stay centered and slightly above center.

## Gouache Treachery asset set

- CrimePack: `treachery`
- Asset set id: `gouache-treachery`
- Display name: `Gouache Treachery`
- Default/fallback behavior is unchanged:
  - default asset set stays `treachery`
  - pack fallback asset set stays `treachery`
- The pack has its own deck-level gouache fallback art:
  - means: `assets/gouache-treachery/means/default.png`
  - clues: `assets/gouache-treachery/clues/default.png`
- Card-specific art should be generated from `PE/image-gouache-1-template.md`, requested at 7:10 portrait (`1050x1500`), and validated with `identify` before committing.
- See `docs/wordpacks-gouache-treachery.md` for the asset-set notes and progress log.

## Storybook Cel asset set

- CrimePack: `treachery`
- Asset set id: `storybook-cel`
- Display name: `Storybook Cel`
- Default/fallback behavior is unchanged:
  - default asset set stays `treachery`
  - pack fallback asset set stays `treachery`
- The pack intentionally ships only deck-level fallback art now:
  - means: `assets/storybook-cel/means/default.png`
  - clues: `assets/storybook-cel/clues/default.png`
- Every Storybook Cel means/clue card uses its deck default first; card-specific `treachery` art remains available as the alternate fallback image.
- See `docs/wordpacks-storybook-cel-pilot.md` for the art direction and prompt template.

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
