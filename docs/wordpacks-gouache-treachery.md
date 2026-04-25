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

- Source meta-prompt: `PE/image-gouache-V1.1-template.md` for current batches; earlier seed images used the original gouache template before the V1.1 refinement.
- Target aspect ratio: **7:10 portrait** (`1050x1500` requested), with a few-pixel tolerance accepted after generation.
- Prompting should keep the lower 20 percent calmer for the card title overlay.
- Outputs are checked with `identify`; if an image lands outside the 7:10 ratio tolerance, remove it and regenerate rather than locally cropping/resizing.
- No local Python postprocessing is part of this pass.
- Batched generation/validation workflow: `workflows/asset-image-gen-v1.md`.

## Progress log

- 2026-04-24: Created the asset set metadata and directories; started generating gouache fallback and card-specific art; current workflow uses ten-image validation checkpoints.

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

### 2026-04-24 checkpoint 4

Generated and validated the next ten means-card images with ImageMagick `identify`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/024-drown.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/025-dumbbell.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/026-e-bike.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/027-electric-baton.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/028-electric-current.png` | 1050x1498 | 0.700935 |
| `assets/gouache-treachery/means/029-explosives.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/030-folding-chair.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/031-gunpowder.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/032-hammer.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/033-hook.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`.

### 2026-04-24 checkpoint 5

Generated and validated the next ten means-card images with ImageMagick `identify`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/034-ice-skates.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/035-illegal-drug.png` | 1049x1500 | 0.699333 |
| `assets/gouache-treachery/means/036-injection.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/037-kerosene.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/038-kick.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/039-knife-and-fork.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/040-lighter.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/041-liquid-drug.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/042-locked-room.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/043-machete.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`.

### 2026-04-25 partial checkpoint 6

Generation was interrupted by operator request after five images. The five copied project assets were validated with ImageMagick `identify` and retained as a natural endpoint:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/044-machine.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/045-mad-dog.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/046-match.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/047-mercury.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/048-metal-chain.png` | 1049x1499 | 0.699800 |

All five are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/means/049-metal-wire.png`.

### 2026-04-25 checkpoint 7

Generated and validated the next ten means-card images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/049-metal-wire.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/050-overdose.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/051-packing-tape.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/052-pesticide.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/053-pill.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/054-pillow.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/055-pistol.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/056-plague.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/057-plastic-bag.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/058-poisonous-gas.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/means/059-poisonous-needle.png`.

### 2026-04-25 checkpoint 8

Generated and validated the next ten means-card images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/059-poisonous-needle.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/060-potted-plant.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/061-powder-drug.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/062-punch.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/063-push.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/064-radiation.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/065-razor-blade.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/066-rope.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/067-scarf.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/068-scissors.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/means/069-seafood.png`.

### 2026-04-25 checkpoint 9

Generated and validated the next ten means-card images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/069-sculpture.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/070-smoke.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/071-sniper.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/072-starvation.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/073-steel-tube.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/074-stone.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/075-sulfuric-acid.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/076-surgery.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/077-throat-slit.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/078-towel.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/means/079-trophy.png`.

### 2026-04-25 checkpoint 10

Validated the pre-existing next pending asset and generated/validated the next ten means-card images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/079-trophy.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/080-trowel.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/081-unarmed.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/082-venomous-scorpion.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/083-venomous-snake.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/084-video-game-console.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/085-virus.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/086-whip.png` | 1049x1500 | 0.699333 |
| `assets/gouache-treachery/means/087-wine.png` | 1049x1500 | 0.699333 |
| `assets/gouache-treachery/means/088-wire.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/means/089-work.png` | 1049x1499 | 0.699800 |

All listed images are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/means/090-wrench.png`.

### 2026-04-25 checkpoint 11

Generated and validated one batch of ten images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/means/090-wrench.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/004-apple.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/005-badge.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/006-bandage.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/007-banknote.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/008-bell.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/009-betting-chips.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/010-blood.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/011-bone.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/012-book.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/clues/013-bottle.png`.

### 2026-04-25 checkpoint 12

Generated and validated one batch of ten clues-card images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/clues/013-bracelet.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/014-bread.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/015-briefs.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/016-broom.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/017-bullet.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/018-button.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/019-cake.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/020-calender.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/021-candy.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/022-carton.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/clues/023-cassette-tape.png`.

### 2026-04-25 checkpoint 13

Generated and validated one batch of ten clues-card images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/clues/023-cassette-tape.png` | 1050x1498 | 0.700935 |
| `assets/gouache-treachery/clues/024-cat.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/025-certificate.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/026-chalk.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/027-cigar.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/028-cigarette-ash.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/029-cigarette-butt.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/030-cleaning-cloth.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/031-cockroach.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/032-coffee.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/clues/033-coins.png`.

### 2026-04-25 checkpoint 14

Validated seven pre-existing clues-card images, then generated and validated one batch of ten additional clues-card images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/clues/033-coins.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/034-comics.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/035-computer.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/036-computer-disk.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/037-computer-mouse.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/038-confidential-letter.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/039-cosmetic-mask.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/040-cotton.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/041-cup.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/042-curtains.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/043-dentures.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/044-diamond.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/045-diary.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/046-dice.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/047-dictionary.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/048-dirt.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/049-documents.png` | 1049x1499 | 0.699800 |

All listed images are within the accepted few-pixel 7:10 tolerance. Prompt briefs for the generated batch are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/clues/050-dust.png`.

### 2026-04-25 checkpoint 15

Generated and validated one batch of ten clues-card images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/clues/050-dog-fur.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/051-dust.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/052-earrings.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/053-eggs.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/054-electric-circuit.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/055-envelope.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/056-exam-paper.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/057-express-courier.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/058-fan.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/059-fax.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/clues/060-fiber-optics.png`.

### 2026-04-25 checkpoint 16

Generated and validated one batch of ten clues-card images with ImageMagick `identify` using `PE/image-gouache-V1.1-template.md`:

| Path | Dimensions | Ratio |
| --- | ---: | ---: |
| `assets/gouache-treachery/clues/060-fiber-optics.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/061-fingernails.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/062-flashlight.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/063-flip-flop.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/064-flute.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/065-flyer.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/066-food-ingredients.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/067-gear.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/068-gift.png` | 1049x1499 | 0.699800 |
| `assets/gouache-treachery/clues/069-gloves.png` | 1049x1499 | 0.699800 |

All ten are within the accepted few-pixel 7:10 tolerance. Prompt briefs are recorded in `PE/gouache-treachery-prompts.md`. The next pending asset is `assets/gouache-treachery/clues/070-glue.png`.
