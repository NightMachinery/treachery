# Asset image generation workflow v1

This workflow is for generated bitmap card-art asset packs such as `gouache-treachery`.
It optimizes for generation throughput by batching validation: generate **N=10** images, then validate aspect ratios for the whole batch at once.

## Batch contract

- Batch size: `N=10` target assets.
- Target output: 7:10 portrait, requested as `1050x1500`.
- Aspect ratio tolerance: accept a few-pixel mismatch from 7:10. As a practical check, accept images where `abs((width / height) - 0.7) <= 0.005`.
- Do not locally crop, resize, or otherwise postprocess generated images in this workflow.
- If an image fails aspect-ratio validation, remove only the copied project asset. Leave the original generated image under `$CODEX_HOME/generated_images/` intact unless explicitly told otherwise.
- A removed asset remains missing from the asset pack, so it automatically re-enters the pending set for a later batch.
- Commit and push after every batch that leaves 10 valid project images, or at a natural endpoint if fewer than 10 are intentionally being finalized.

## Inputs

For each asset in the batch, record:

- deck: `means` or `clues`
- card id or `default`
- card label, if card-specific
- destination path under the asset pack
- handcrafted prompt derived from the active meta-prompt, for example `PE/image-gouache-1-template.md`

Keep prompt notes in a project file such as `PE/<asset-pack>-prompts.md`, and update any relevant docs under `docs/` when the asset pack behavior or process changes.

## Build the next pending batch

1. List expected assets for the pack.
2. Remove any destination paths that already exist and passed validation in a prior checkpoint.
3. Take the next `N=10` missing assets in deterministic order.
4. Include any assets removed for bad aspect ratio in a future batch automatically because their destination files no longer exist.

Example pending-set helper for a Treachery-style pack:

```bash
asset_set="gouache-treachery"
pack="wordpacks/crime/treachery"

{
  printf '%s\n' "$pack/assets/$asset_set/means/default.png"
  printf '%s\n' "$pack/assets/$asset_set/clues/default.png"
  sed 's#^#'"$pack"'/assets/'"$asset_set"'/means/#; s#$#.png#' "$pack/means/ids.txt"
  sed 's#^#'"$pack"'/assets/'"$asset_set"'/clues/#; s#$#.png#' "$pack/clues/ids.txt"
} | while read -r path; do
  [ -e "$path" ] || printf '%s\n' "$path"
done | head -10
```

## Generate the batch

For each of the 10 pending assets:

1. Handcraft the prompt from the meta-prompt and the card label/role.
2. Request `1050x1500`, 7:10 portrait, full-bleed illustration-only output.
3. Copy the generated image from `$CODEX_HOME/generated_images/<session>/<image_id>.png` into the intended project path.
4. Do not run `identify` yet unless you need to debug a generation failure; keep generating until all 10 project files have been copied.

The built-in image tool may still require one generation call per distinct asset. The batching improvement is that validation and cleanup happen once per 10 copied outputs instead of interrupting generation after each image.

## Validate all 10 at once

Save the batch destination paths in a file, for example `tmp/asset-image-batch.txt`, then run:

```bash
mkdir -p tmp
# tmp/asset-image-batch.txt should contain one project asset path per line.

while read -r path; do
  identify -format '%w %h %[fx:w/h] %i\n' "$path"
done < tmp/asset-image-batch.txt
```

To remove invalid project copies automatically:

```bash
target=0.7
tolerance=0.005

while read -r path; do
  [ -e "$path" ] || continue
  read -r width height ratio name < <(identify -format '%w %h %[fx:w/h] %i\n' "$path")
  awk -v r="$ratio" -v t="$target" -v tol="$tolerance" 'BEGIN { d=r-t; if (d<0) d=-d; exit(d <= tol ? 0 : 1) }'
  if [ "$?" -ne 0 ]; then
    echo "Removing bad-ratio project copy: $path ($width x $height, ratio=$ratio)"
    rm -f "$path"
  fi
done < tmp/asset-image-batch.txt
```

After cleanup, count retained valid batch assets:

```bash
while read -r path; do
  [ -e "$path" ] && printf '%s\n' "$path"
done < tmp/asset-image-batch.txt | wc -l
```

If the count is less than 10, continue with a later pending batch. The removed images will be selected again because their destination files are missing.

## Update records

After validation:

1. Append concise prompt summaries to `PE/<asset-pack>-prompts.md`.
2. Update the asset-pack progress log in `docs/` with paths, dimensions, and ratios for retained images.
3. Keep `tmp/asset-image-batch.txt` uncommitted or delete it.

## Commit checkpoint

Commit only after a validation pass has retained 10 valid project images, or when intentionally stopping at a smaller natural endpoint.

Recommended checkpoint commands:

```bash
git status --short
find wordpacks/crime/treachery/assets/gouache-treachery -type f -print | sort \
  | xargs -r identify -format '%i %w %h %[fx:w/h]\n'
git add PE/<asset-pack>-prompts.md docs/ wordpacks/crime/treachery/assets/<asset-pack>/
git commit -m "Add <asset-pack> image batch <n>"
git push
```

If invalid project copies were removed, do not commit those failed outputs. They are intentionally left out so the next pending-set pass can regenerate them.
