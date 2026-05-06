export interface TgForensicCard {
  cardId: string;
  cardName: string;
  choices: string[];
  choiceIds: string[];
  selectedChoiceId: string;
  selectedChoice: string;
  replaced: boolean;
}

export interface TgCard {
  id: string;
  altImgUrl: string;
  guessedBy: string[];
  hasImage: boolean;
  imgUrl: string;
  name: string;
  aspectRatio?: string;
  width?: number;
  height?: number;
}

export type TgParticipantRole = 'player' | 'observer';
export type TgSecretRole = 'investigator' | 'murderer' | 'accomplice' | 'witness';
export type TgWinner = 'none' | 'investigatorTeam' | 'murdererTeam';

export interface TgParticipant {
  name: string;
  uid: string;
  role: TgParticipantRole;
  isCreator: boolean;
  isScientist: boolean;
  isMarkedScientist: boolean;
  isBot?: boolean;
}

export interface TgViewer {
  uid: string;
  name: string;
  role: TgParticipantRole;
  isParticipant: boolean;
  isCreator: boolean;
  isScientist: boolean;
}

export interface TgGuess {
  guessedByUid: string;
  murdererUid: string;
  meansCardId: string;
  meansCardName: string;
  clueCardId: string;
  clueCardName: string;
  correct: boolean;
  createdTimestamp?: string;
}

export interface TgPartialGuess {
  murdererUid: string;
  meansCardId: string;
  meansCardName?: string;
  clueCardId: string;
  clueCardName?: string;
}

export interface TgPlayer {
  name: string;
  uid: string;
  clueCards: TgCard[];
  meansCards: TgCard[];
  guessed: boolean;
}

export enum TgMessageType {
  CHAT = 'chat',
  GUESS = 'guess',
  FORENSIC = 'forensic',
}

export interface TgMessage {
  playerUid: string;
  message: string;
  timestamp: string;
  type: TgMessageType;
}

export interface TgRoomTimer {
  durationSeconds: number;
  expiresAt: string;
  pausedRemainingSeconds: number;
  runId: number;
}

export interface TgGame {
  creatorUid: string;
  scientistUid: string;
  markedScientistUid: string;
  causeCard: TgForensicCard;
  locationCard: TgForensicCard;
  otherCards: TgForensicCard[];
  gameId: string;
  createdTimestamp: string;
  startedTimestamp: string;
  murdererSelected: boolean;
  murdererCardsSelected: boolean;
  murdererClueCardId: string;
  murdererClueCardName: string;
  murdererMeansCardId: string;
  murdererMeansCardName: string;
  murdererUid: string;
  startedOn: string;
  finished: boolean;
  meansCardsPerPlayer: number;
  clueCardsPerPlayer: number;
  linkClueCountToMeans: boolean;
  meansCluesTextOnly: boolean;
  randomMurdererCardSelection: boolean;
  showAllRolesToScientist: boolean;
  crimePackId: string;
  crimePackLanguage: string;
  crimePackAssetSetId: string;
  hintPackId: string;
  hintPackLanguage: string;
  accompliceCount: number;
  witnessCount: number;
  witnessesToFind: number;
  pendingWitnessSelection: boolean;
  winner: TgWinner;
  finishedReason: string;
  resultMessage: string;
  roomTimer?: TgRoomTimer;
}

export interface TgKnownRolePlayer {
  uid: string;
  name: string;
  role: TgSecretRole;
}

export interface TgWitnessSelectionPromptState {
  requiredSelections: number;
  dismissible: boolean;
  creatorInitiated: boolean;
  active: boolean;
}

export interface TgForensicPrivateData {
  murderer: TgPlayer;
  murdererClueCardId: string;
  murdererClueCardName: string;
  murdererMeansCardId: string;
  murdererMeansCardName: string;
}

export interface TgPlayerPrivateData {
  isMurderer: boolean;
  clueCardId: string;
  clueCardName: string;
  meansCardId: string;
  meansCardName: string;
  role: TgSecretRole;
  knownMurdererTeam: TgKnownRolePlayer[];
  knownMurdererClueCardId: string;
  knownMurdererClueCardName: string;
  knownMurdererMeansCardId: string;
  knownMurdererMeansCardName: string;
  activeWitnessSelectionPrompt: TgWitnessSelectionPromptState;
}

export interface TgForensicCardResource {
  causeCards: TgForensicCard[];
  locationCards: TgForensicCard[];
  otherCards: TgForensicCard[];
}

export interface TgCrimePackResource {
  packId: string;
  packName: string;
  language: string;
  assetSetId: string;
  defaultAssetSetId: string;
  hasAnyImages: boolean;
  languages: TgPackLanguageOption[];
  assetSets: TgPackAssetSetOption[];
  clueCards: TgCard[];
  meansCards: TgCard[];
}

export interface TgHintPackResource {
  packId: string;
  packName: string;
  language: string;
  languages: TgPackLanguageOption[];
  forensicCards: TgForensicCardResource;
}

export interface TgCurrentPackResources {
  crimePack: TgCrimePackResource;
  hintPack: TgHintPackResource;
  clueCards: TgCard[];
  meansCards: TgCard[];
  forensicCards: TgForensicCardResource;
}

export interface TgPackLanguageOption {
  id: string;
  name: string;
}

export interface TgPackAssetSetOption {
  id: string;
  name: string;
  hasAnyImages: boolean;
  aspectRatio: string;
  width: number;
  height: number;
}

export interface TgCrimePackCatalogEntry {
  id: string;
  name: string;
  defaultLanguage: string;
  defaultAssetSetId: string;
  fallbackAssetSetId?: string;
  meansCount: number;
  clueCount: number;
  hasAnyImages: boolean;
  languages: TgPackLanguageOption[];
  assetSets: TgPackAssetSetOption[];
}

export interface TgHintPackCatalogEntry {
  id: string;
  name: string;
  defaultLanguage: string;
  causeCount: number;
  locationCount: number;
  otherCount: number;
  languages: TgPackLanguageOption[];
}

export interface TgWordpackCatalog {
  crimePacks: TgCrimePackCatalogEntry[];
  hintPacks: TgHintPackCatalogEntry[];
}

export interface TgMurdererInfo {
  murderer: TgPlayer;
  clueCard: TgCard;
  meansCard: TgCard;
}

export interface TgWitnessPromptCandidate {
  uid: string;
  name: string;
}

export interface TgModeratorPrivateData {
  witnessPromptCandidates: TgWitnessPromptCandidate[];
}

export interface TgRoleRevealEntry {
  uid: string;
  name: string;
  role: string;
}

export interface TgGameSnapshot {
  game: TgGame;
  participants: TgParticipant[];
  players: TgPlayer[];
  guesses: TgGuess[];
  messages: TgMessage[];
  viewer: TgViewer;
  playerPrivateData: TgPlayerPrivateData;
  forensicPrivateData: TgForensicPrivateData;
  moderatorPrivateData: TgModeratorPrivateData;
  roleReveal: TgRoleRevealEntry[];
  serverTimestamp: string;
}

export interface TgAuthUser {
  uid: string;
  token: string;
  displayName: string;
}

export interface TgGameSettingsInput {
  meansCardsPerPlayer: number;
  clueCardsPerPlayer: number;
  linkClueCountToMeans: boolean;
  meansCluesTextOnly: boolean;
  randomMurdererCardSelection: boolean;
  showAllRolesToScientist: boolean;
  crimePackId: string;
  crimePackLanguage: string;
  crimePackAssetSetId: string;
  hintPackId: string;
  hintPackLanguage: string;
  accompliceCount: number;
  witnessCount: number;
  witnessesToFind: number;
}

export interface TgGameRoomModsInput {
  meansCluesTextOnly?: boolean;
  crimePackLanguage?: string;
  crimePackAssetSetId?: string;
  hintPackLanguage?: string;
}
