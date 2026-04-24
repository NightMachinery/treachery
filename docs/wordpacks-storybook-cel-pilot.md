# Storybook Cel pilot asset set

This document records the current Treachery pilot art pack added as a new selectable CrimePack asset set.

## Asset set

- pack: `treachery`
- asset set id: `storybook-cel`
- asset set name: `Storybook Cel`
- status: in-progress pilot subset
- target image size: `1050x1500`
- accepted ratio-compatible size: `1400x2000`
- render assumption in the UI: centered `cover` crop

## Art direction

- storybook cel-animation look
- high contrast, bold silhouettes, and large geometric shapes
- flat saturated gouache-like color
- crisp black ink outlines
- hard-edged cel-shaded lighting
- one dominant subject per card
- keep the subject centered and slightly above center
- keep the bottom area visually simple because the card label overlays that region

## Prompt template

```text
Use case: illustration-story
Asset type: <means/clue> card art for a mobile game image pack
Primary request: <card label>
Subject: <one dominant readable subject>
Style/medium: Studio Ghibli animation cel. Maximize high contrast, bold silhouettes, and large, distinct geometric shapes for instant small-scale legibility. Render using flat, highly saturated gouache pigments, crisp black ink outlines, and sharp, hard-edged cel-shaded lighting.
Composition/framing: 7:10 portrait card illustration, subject centered and slightly above center, minimal background clutter, bottom area visually calm for overlaid label
Constraints: no text, no watermark, no border, highly legible at thumbnail size
```

Saved as repo template: `PE/image-ghibli-1-template.md`

## Current generated pilot cards

### Means

- `006-bamboo-tip`
- `007-bat`
- `025-dumbbell`
- `029-explosives`
- `037-kerosene`
- `042-locked-room`
- `055-pistol`
- `057-plastic-bag`
- `064-radiation`
- `070-smoke`
- `081-unarmed`
- `088-wire`

### Clues

- `015-briefs`
- `019-cake`
- `065-flyer`
- `074-handcuffs`
- `093-juice`

## Notes

- The existing `treachery` art remains the default selected asset set.
- Missing `storybook-cel` cards should first use Storybook Cel deck defaults, then fall back to the existing `treachery` asset set.
- Storybook Cel now has its own fallback default cards:
  - `assets/storybook-cel/means/default.png`
  - `assets/storybook-cel/clues/default.png`
- Aspect-ratio validation should be tolerant of a few pixels mismatch as long as the generated image remains effectively `7:10`.
