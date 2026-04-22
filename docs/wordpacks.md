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
