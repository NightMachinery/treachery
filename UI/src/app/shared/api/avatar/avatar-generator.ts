import { h } from 'preact';
import renderToString from 'preact-render-to-string';
import NiceAvatar, { type NiceAvatarProps } from '@nice-avatar-svg/preact';

const BACKGROUND_COLORS = ['#E0DDFF', '#D2EFF3', '#FFEBA4', '#9287FF', '#6BD9E9'] as const;
const SKIN_COLORS = ['#F9C9B6', '#E8B196', '#AC6651'] as const;
const HAIR_COLORS = ['#171921', '#3A2618', '#6E442C', '#9287FF'] as const;
const SHIRT_COLORS = ['#F4D150', '#E0DDFF', '#D2EFF3', '#9287FF'] as const;

const EAR_SIZES: NiceAvatarProps['earSize'][] = ['small', 'big'];
const HAIR_STYLES: NiceAvatarProps['hairStyle'][] = ['dannyPhantom', 'dougFunny', 'fonze', 'full', 'mrT', 'pixie', 'turban'];
const NOSE_STYLES: NiceAvatarProps['noseStyle'][] = ['curve', 'pointed', 'round'];
const MOUTH_STYLES: NiceAvatarProps['mouthStyle'][] = ['frown', 'laughing', 'nervous', 'pucker', 'sad', 'smile', 'smirk', 'surprised'];
const SHIRT_STYLES: NiceAvatarProps['shirtStyle'][] = ['collared', 'crew', 'open'];
const EYES_STYLES: NiceAvatarProps['eyesStyle'][] = ['base', 'round', 'shadow', 'smiling'];
const EYEBROW_STYLES: NiceAvatarProps['eyebrowsStyle'][] = ['down', 'eyelashesDown', 'eyelashesUp', 'up'];
const GLASSES_STYLES: Array<NiceAvatarProps['glassesStyle'] | undefined> = [undefined, 'round', 'square'];
const FACIAL_HAIR_STYLES: Array<NiceAvatarProps['facialHairStyle'] | undefined> = [undefined, 'beard', 'scruff'];
const EARRING_STYLES: Array<NiceAvatarProps['earRing'] | undefined> = [undefined, 'loop'];

function normalizePart(value?: string) {
  return String(value || '').trim();
}

export function buildAvatarSeed(uid?: string, displayName?: string) {
  const normalizedUid = normalizePart(uid);
  const normalizedName = normalizePart(displayName);

  if (normalizedUid && normalizedName) {
    return `${normalizedUid}:${normalizedName}`;
  }
  if (normalizedUid) {
    return normalizedUid;
  }
  if (normalizedName) {
    return normalizedName;
  }
  return 'anonymous';
}

export function getAvatarFallbackColor(uid?: string, displayName?: string) {
  const config = buildAvatarConfig(uid, displayName);
  return config.bgColor || BACKGROUND_COLORS[0];
}

export function renderAvatarDataUri(uid?: string, displayName?: string) {
  const svg = renderToString(h(NiceAvatar, buildAvatarConfig(uid, displayName)));
  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
}

function buildAvatarConfig(uid?: string, displayName?: string): NiceAvatarProps {
  const seed = buildAvatarSeed(uid, displayName);

  return {
    shape: 'circle',
    bgColor: pick(seed, 'background', BACKGROUND_COLORS),
    skinColor: pick(seed, 'skin', SKIN_COLORS),
    hairColor: pick(seed, 'hair', HAIR_COLORS),
    shirtColor: pick(seed, 'shirt', SHIRT_COLORS),
    earSize: pick(seed, 'ears', EAR_SIZES),
    hairStyle: pick(seed, 'hair-style', HAIR_STYLES),
    noseStyle: pick(seed, 'nose-style', NOSE_STYLES),
    mouthStyle: pick(seed, 'mouth-style', MOUTH_STYLES),
    shirtStyle: pick(seed, 'shirt-style', SHIRT_STYLES),
    eyesStyle: pick(seed, 'eyes-style', EYES_STYLES),
    eyebrowsStyle: pick(seed, 'eyebrows-style', EYEBROW_STYLES),
    glassesStyle: pick(seed, 'glasses-style', GLASSES_STYLES),
    facialHairStyle: pick(seed, 'facial-hair-style', FACIAL_HAIR_STYLES),
    earRing: pick(seed, 'earring-style', EARRING_STYLES),
  };
}

function pick<T>(seed: string, feature: string, values: readonly T[]): T {
  return values[hashNumber(`${seed}:${feature}`) % values.length];
}

function hashNumber(value: string) {
  let hash = 0x811c9dc5;

  for (let i = 0; i < value.length; i++) {
    hash ^= value.charCodeAt(i);
    hash = Math.imul(hash, 0x01000193);
  }

  return hash >>> 0;
}
