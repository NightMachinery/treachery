# Gouache Treachery asset set

`gouache-treachery` is a selectable Treachery CrimePack image asset set for gallery-grade gouache card art.

## Pack wiring

- CrimePack: `treachery`
- Asset set id: `gouache-treachery`
- Display name: `Gouache Treachery`
- Default CrimePack asset set remains `treachery`.
- Pack fallback asset set remains `treachery`.
- The gouache set has its own deck-level fallback cards instead of reusing another non-default art set:
  - means: `wordpacks/crime/treachery/assets/gouache-treachery/means/default.png`
  - clues: `wordpacks/crime/treachery/assets/gouache-treachery/clues/default.png`

## Image generation direction

- Source meta-prompt: `PE/image-gouache-1-template.md`.
- Target aspect ratio: **7:10 portrait** (`1050x1500` requested), with a few-pixel tolerance accepted after generation.
- Prompting should keep the lower 20 percent calmer for the card title overlay.
- Outputs are checked with `identify`; if an image lands outside the 7:10 ratio tolerance, remove it and regenerate rather than locally cropping/resizing.
- No local Python postprocessing is part of this pass.
- Batched generation/validation workflow: `workflows/asset-image-gen-v1.md`.

## Progress log

- 2026-04-24: Created the asset set metadata and directories; started generating gouache fallback and card-specific art in five-image checkpoints.

## Generated image checkpoints

### 2026-04-24 checkpoint 1

Generated and validated the first five images with ImageMagick `identify`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/default.png` | 1050x1498 | 0.700935 |
| `assets/gouache-treachery/clues/default.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/001-alcohol.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/001-air-conditioning.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/002-amoeba.png` | 1049x1499 | 0.699800 |

All are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`.

### 2026-04-24 checkpoint 2

Generated and validated the next ten means-card images with ImageMagick `identify`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/004-arson.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/005-axe.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/006-bamboo-tip.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/007-bat.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/008-belt.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/009-bite-and-tear.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/010-blender.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/011-blood-release.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/012-box-cutter.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/013-brick.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`.

### 2026-04-24 checkpoint 3

Generated and validated the next ten means-card images with ImageMagick `identify`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/014-bury.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/015-candlestick.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/016-chainsaw.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/017-chemicals.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/018-cleaver.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/019-crutch.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/020-dagger.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/021-dirty-water.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/022-dismember.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/023-drill.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`.
