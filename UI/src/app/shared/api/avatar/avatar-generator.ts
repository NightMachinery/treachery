const BACKGROUND_COLORS = ['#E0DDFF', '#D2EFF3', '#FFEBA4', '#9287FF', '#6BD9E9'] as const;
const SKIN_COLORS = ['#F9C9B6', '#E8B196', '#AC6651'] as const;
const HAIR_COLORS = ['#171921', '#3A2618', '#6E442C', '#5B4BC4'] as const;
const SHIRT_COLORS = ['#F4D150', '#E0DDFF', '#D2EFF3', '#9287FF'] as const;
const SHIRT_ACCENT_COLORS = ['#FFF6C2', '#F4EFFF', '#F3FBFD', '#DCD6FF'] as const;
const GLASSES_COLORS = ['#171921', '#3D4B6D'] as const;

const HAIR_STYLES = ['wave', 'bowl', 'part', 'spike', 'cap'] as const;
const EYE_STYLES = ['dot', 'round', 'smile'] as const;
const BROW_STYLES = ['up', 'flat', 'down', 'none'] as const;
const NOSE_STYLES = ['short', 'curve', 'pointed'] as const;
const MOUTH_STYLES = ['smile', 'laugh', 'smirk', 'open', 'frown'] as const;
const GLASSES_STYLES = ['none', 'round', 'square'] as const;
const FACIAL_HAIR_STYLES = ['none', 'scruff', 'beard'] as const;
const EARRING_STYLES = ['none', 'loop'] as const;
const EAR_SIZES = ['small', 'big'] as const;

type HairStyle = (typeof HAIR_STYLES)[number];
type EyeStyle = (typeof EYE_STYLES)[number];
type BrowStyle = (typeof BROW_STYLES)[number];
type NoseStyle = (typeof NOSE_STYLES)[number];
type MouthStyle = (typeof MOUTH_STYLES)[number];
type GlassesStyle = (typeof GLASSES_STYLES)[number];
type FacialHairStyle = (typeof FACIAL_HAIR_STYLES)[number];
type EarringStyle = (typeof EARRING_STYLES)[number];
type EarSize = (typeof EAR_SIZES)[number];

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
  const seed = buildAvatarSeed(uid, displayName);
  return pick(seed, 'fallback-color', BACKGROUND_COLORS);
}

export function renderAvatarDataUri(uid?: string, displayName?: string) {
  const seed = buildAvatarSeed(uid, displayName);
  const svg = renderAvatarSvg(seed);
  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
}

function renderAvatarSvg(seed: string) {
  const backgroundColor = pick(seed, 'background', BACKGROUND_COLORS);
  const skinColor = pick(seed, 'skin', SKIN_COLORS);
  const hairColor = pick(seed, 'hair', HAIR_COLORS);
  const shirtColor = pick(seed, 'shirt', SHIRT_COLORS);
  const shirtAccentColor = pick(seed, 'shirt-accent', SHIRT_ACCENT_COLORS);
  const glassesColor = pick(seed, 'glasses-color', GLASSES_COLORS);
  const hairStyle = pick(seed, 'hair-style', HAIR_STYLES);
  const eyeStyle = pick(seed, 'eye-style', EYE_STYLES);
  const browStyle = pick(seed, 'brow-style', BROW_STYLES);
  const noseStyle = pick(seed, 'nose-style', NOSE_STYLES);
  const mouthStyle = pick(seed, 'mouth-style', MOUTH_STYLES);
  const glassesStyle = pick(seed, 'glasses-style', GLASSES_STYLES);
  const facialHairStyle = pick(seed, 'facial-hair-style', FACIAL_HAIR_STYLES);
  const earringStyle = pick(seed, 'earring-style', EARRING_STYLES);
  const earSize = pick(seed, 'ear-size', EAR_SIZES);
  const blushOpacity = opacity(seed, 'blush', 0.08, 0.18, 2);
  const faceShadowOpacity = opacity(seed, 'face-shadow', 0.1, 0.24, 2);

  return `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-label="Generated avatar">
  <defs>
    <clipPath id="avatar-clip">
      <circle cx="32" cy="32" r="32" />
    </clipPath>
  </defs>
  <g clip-path="url(#avatar-clip)">
    <rect width="64" height="64" fill="${backgroundColor}" />
    <circle cx="18" cy="14" r="12" fill="rgba(255,255,255,0.18)" />
    <path d="M6 64c2-11 11-18 26-18s24 7 26 18H6Z" fill="${shirtColor}" />
    <path d="M14 64c3-8 9-13 18-13s15 5 18 13H14Z" fill="${shirtAccentColor}" opacity="0.82" />
    <rect x="28.4" y="34.3" width="7.2" height="9.7" rx="3.2" fill="${skinColor}" />
    ${renderEars(earSize, skinColor)}
    <ellipse cx="32" cy="28.3" rx="12.4" ry="13.1" fill="${skinColor}" />
    <ellipse cx="32" cy="30.4" rx="11.4" ry="12.2" fill="#000" opacity="${faceShadowOpacity}" />
    <ellipse cx="32" cy="28.1" rx="11.7" ry="12.3" fill="${skinColor}" />
    <ellipse cx="24.3" cy="32.5" rx="1.9" ry="1.2" fill="#F08C96" opacity="${blushOpacity}" />
    <ellipse cx="39.7" cy="32.5" rx="1.9" ry="1.2" fill="#F08C96" opacity="${blushOpacity}" />
    ${renderHair(hairStyle, hairColor)}
    ${renderEyebrows(browStyle)}
    ${renderEyes(eyeStyle)}
    ${renderGlasses(glassesStyle, glassesColor)}
    ${renderNose(noseStyle)}
    ${renderMouth(mouthStyle)}
    ${renderFacialHair(facialHairStyle, hairColor)}
    ${renderEarring(earringStyle)}
  </g>
</svg>`.trim();
}

function renderEars(earSize: EarSize, skinColor: string) {
  if (earSize === 'big') {
    return `
    <ellipse cx="18.4" cy="28.6" rx="2.5" ry="4.3" fill="${skinColor}" />
    <ellipse cx="45.6" cy="28.6" rx="2.5" ry="4.3" fill="${skinColor}" />`;
  }

  return `
    <ellipse cx="19.1" cy="28.6" rx="1.8" ry="3.3" fill="${skinColor}" />
    <ellipse cx="44.9" cy="28.6" rx="1.8" ry="3.3" fill="${skinColor}" />`;
}

function renderHair(style: HairStyle, color: string) {
  switch (style) {
    case 'bowl':
      return `
      <path d="M18.7 27.2c.1-9.6 6.2-15 13.3-15 7.2 0 13.1 5.7 13.3 15-2.7-1.9-5.8-2.6-9.1-2.6H27.8c-3.3 0-6.4.7-9.1 2.6Z" fill="${color}" />
      <path d="M19.7 24.1c2.5-5.9 6.9-9 12.3-9 5.5 0 10 3.1 12.5 9-2.3-1.1-4.9-1.6-7.8-1.6h-9.4c-2.7 0-5.3.6-7.6 1.6Z" fill="#fff" opacity="0.08" />`;
    case 'part':
      return `
      <path d="M19.2 28.2c.1-9.9 6.3-15.5 12.8-15.5 4 0 7.5 1.6 10 4.8 2.2 2.7 3.3 6.2 3.5 10.7-2.4-2.6-4.8-4.2-7-4.9-1.7-.6-3.8-.8-6.5-.8-5.3 0-9.9 2-12.8 5.7Z" fill="${color}" />
      <path d="M33.3 12.9c2.1 2.8 3 6.2 2.6 10.3" fill="none" stroke="#fff" stroke-linecap="round" stroke-width="1.2" opacity="0.18" />`;
    case 'spike':
      return `
      <path d="M18.4 28.2c.3-7.8 4.4-14.6 13.6-14.6 9.5 0 13 7 13.5 14.6-1.2-1.3-2.7-2.9-4.7-5.1l-2.1 2.9-3.3-5.7-3 5.1-2.4-4.1-3.8 4.8-2.4-3.3c-1.6 2-3.2 3.8-5.4 5.4Z" fill="${color}" />`;
    case 'cap':
      return `
      <path d="M19 27.9c.5-8.6 5.8-14.4 13-14.4s12.6 5.6 13 14.4c-2.7-1.6-6.1-2.4-10.2-2.4h-5.7c-4 0-7.4.8-10.1 2.4Z" fill="${color}" />
      <path d="M18.7 27.6c3.2-2.5 7-3.8 11.8-3.8h3.5c4.8 0 8.6 1.3 11.4 3.8v2.9c-2.9-2.1-6.5-3.2-10.9-3.2h-5.4c-4.1 0-7.6 1.1-10.4 3.2v-2.9Z" fill="#171921" opacity="0.28" />`;
    case 'wave':
    default:
      return `
      <path d="M18.6 28.4c.2-10.2 6.8-15.7 13.5-15.7 6.9 0 12.8 5.5 13.4 15.7-2-2.4-5-3.7-9-3.7-1.5 0-3 .1-4.4.4-1.1.2-2.2.4-3.2.4-2.3 0-4.3-.7-6-2.1-.5 1.6-1.8 3.3-4.3 5Z" fill="${color}" />`;
  }
}

function renderEyebrows(style: BrowStyle) {
  switch (style) {
    case 'up':
      return `
      <path d="M24 24.7c1.2-1 2.6-1.5 4.2-1.4" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.4" />
      <path d="M35.8 23.3c1.6-.1 3 .4 4.2 1.4" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.4" />`;
    case 'down':
      return `
      <path d="M23.7 23.5c1.4.2 2.8.8 3.9 1.8" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.4" />
      <path d="M36.4 25.3c1.1-1 2.4-1.6 3.9-1.8" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.4" />`;
    case 'none':
      return '';
    case 'flat':
    default:
      return `
      <path d="M23.8 24h4.2" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.4" />
      <path d="M36 24h4.2" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.4" />`;
  }
}

function renderEyes(style: EyeStyle) {
  switch (style) {
    case 'round':
      return `
      <circle cx="26.3" cy="28.9" r="1.7" fill="#171921" />
      <circle cx="37.7" cy="28.9" r="1.7" fill="#171921" />
      <circle cx="25.8" cy="28.3" r="0.45" fill="#fff" opacity="0.92" />
      <circle cx="37.2" cy="28.3" r="0.45" fill="#fff" opacity="0.92" />`;
    case 'smile':
      return `
      <path d="M24.6 28.8c.8 1.2 1.8 1.8 3.3 1.8 1.2 0 2.1-.4 2.8-1.1" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.4" />
      <path d="M33.3 29.5c.7.7 1.6 1.1 2.8 1.1 1.5 0 2.5-.6 3.3-1.8" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.4" />`;
    case 'dot':
    default:
      return `
      <ellipse cx="26.4" cy="28.6" rx="1.25" ry="1.55" fill="#171921" />
      <ellipse cx="37.6" cy="28.6" rx="1.25" ry="1.55" fill="#171921" />`;
  }
}

function renderGlasses(style: GlassesStyle, color: string) {
  switch (style) {
    case 'round':
      return `
      <circle cx="26.2" cy="29.2" r="3.5" fill="none" stroke="${color}" stroke-width="1.2" />
      <circle cx="37.8" cy="29.2" r="3.5" fill="none" stroke="${color}" stroke-width="1.2" />
      <path d="M29.7 29.2h4.6" fill="none" stroke="${color}" stroke-width="1.2" />`;
    case 'square':
      return `
      <rect x="22.8" y="25.8" width="6.8" height="6.4" rx="1.6" fill="none" stroke="${color}" stroke-width="1.2" />
      <rect x="34.4" y="25.8" width="6.8" height="6.4" rx="1.6" fill="none" stroke="${color}" stroke-width="1.2" />
      <path d="M29.6 29h4.8" fill="none" stroke="${color}" stroke-width="1.2" />`;
    case 'none':
    default:
      return '';
  }
}

function renderNose(style: NoseStyle) {
  switch (style) {
    case 'curve':
      return `<path d="M32.7 30.8c-.1 1.8.3 3.3 1.2 4.6-.5.5-1.2.8-1.9.9" fill="none" stroke="#8C5C49" stroke-linecap="round" stroke-width="1.1" />`;
    case 'pointed':
      return `<path d="M32.2 30.8c-.1 1.8.5 3.1 1.8 4.2l-2.8.6" fill="none" stroke="#8C5C49" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.1" />`;
    case 'short':
    default:
      return `<path d="M32.1 31.1v3.5" fill="none" stroke="#8C5C49" stroke-linecap="round" stroke-width="1.1" />`;
  }
}

function renderMouth(style: MouthStyle) {
  switch (style) {
    case 'laugh':
      return `
      <path d="M26.7 38.2c1.7 1.6 3.5 2.4 5.3 2.4s3.6-.8 5.3-2.4c-.4 2.3-2.5 4.6-5.3 4.6s-4.9-2.3-5.3-4.6Z" fill="#A33D4F" />
      <path d="M26.7 38.2c1.7 1.6 3.5 2.4 5.3 2.4s3.6-.8 5.3-2.4" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.3" />`;
    case 'smirk':
      return `<path d="M28.3 39.3c1.2 1 2.6 1.5 4.1 1.5 1.2 0 2.5-.4 4.2-1.3" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.3" />`;
    case 'open':
      return `<ellipse cx="32" cy="39.8" rx="2.8" ry="2.4" fill="#A33D4F" stroke="#171921" stroke-width="1.1" />`;
    case 'frown':
      return `<path d="M28.2 40.8c1.1-1.2 2.4-1.8 3.8-1.8s2.7.6 3.8 1.8" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.3" />`;
    case 'smile':
    default:
      return `<path d="M27.6 38.9c1.3 1.6 2.8 2.4 4.4 2.4s3.1-.8 4.4-2.4" fill="none" stroke="#171921" stroke-linecap="round" stroke-width="1.3" />`;
  }
}

function renderFacialHair(style: FacialHairStyle, color: string) {
  switch (style) {
    case 'scruff':
      return `<path d="M28.8 42.3c1 .7 2 .9 3.2.9 1.2 0 2.2-.2 3.2-.9" fill="none" stroke="${color}" stroke-linecap="round" stroke-width="1.1" opacity="0.8" />`;
    case 'beard':
      return `
      <path d="M26.1 39.7c.4 4.2 2.3 6.6 5.9 6.6s5.5-2.4 5.9-6.6" fill="${color}" opacity="0.86" />
      <path d="M26.1 39.7c.4 4.2 2.3 6.6 5.9 6.6s5.5-2.4 5.9-6.6" fill="none" stroke="${color}" stroke-width="0.9" opacity="0.92" />`;
    case 'none':
    default:
      return '';
  }
}

function renderEarring(style: EarringStyle) {
  if (style !== 'loop') {
    return '';
  }

  return `<circle cx="18.2" cy="31.9" r="1.1" fill="none" stroke="#D9B55A" stroke-width="0.9" />`;
}

function opacity(seed: string, feature: string, min: number, max: number, precision = 3) {
  const steps = Math.pow(10, precision);
  const value = min + (hashNumber(`${seed}:${feature}`) / 0xffffffff) * (max - min);
  return (Math.round(value * steps) / steps).toFixed(precision);
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
