declare module '@nice-avatar-svg/preact' {
  import { FunctionComponent } from 'preact';

  export interface NiceAvatarProps {
    shape?: 'circle' | 'rounded' | 'square';
    bgColor?: string;
    hairColor?: string;
    shirtColor?: string;
    skinColor?: string;
    earSize?: 'small' | 'big';
    hairStyle?: 'dannyPhantom' | 'dougFunny' | 'fonze' | 'full' | 'mrT' | 'pixie' | 'turban';
    noseStyle?: 'curve' | 'pointed' | 'round';
    mouthStyle?: 'frown' | 'laughing' | 'nervous' | 'pucker' | 'sad' | 'smile' | 'smirk' | 'surprised';
    shirtStyle?: 'collared' | 'crew' | 'open';
    eyesStyle?: 'base' | 'round' | 'shadow' | 'smiling';
    eyebrowsStyle?: 'down' | 'eyelashesDown' | 'eyelashesUp' | 'up';
    glassesStyle?: 'round' | 'square';
    facialHairStyle?: 'beard' | 'scruff';
    earRing?: 'loop';
  }

  const NiceAvatar: FunctionComponent<NiceAvatarProps>;
  export default NiceAvatar;
}
