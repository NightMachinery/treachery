export interface TgForensicCard {
  cardName: string;
  choices: string[];
  selectedChoice: string;
  replaced: boolean;
}

export interface TgCard {
  altImgUrl: string;
  guessedBy: string[];
  imgUrl: string;
  name: string;
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
  meansCardName: string;
  clueCardName: string;
  correct: boolean;
  createdTimestamp?: string;
}

export interface TgPartialGuess {
  murdererUid: string;
  meansCardName: string;
  clueCardName: string;
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
  FORENSIC = 'forensic'
}

export interface TgMessage {
  playerUid: string;
  message: string;
  timestamp: string;
  type: TgMessageType;
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
  murdererClueCardName: string;
  murdererMeansCardName: string;
  murdererUid: string;
  startedOn: string;
  finished: boolean;
  meansCardsPerPlayer: number;
  clueCardsPerPlayer: number;
  linkClueCountToMeans: boolean;
  accompliceCount: number;
  witnessCount: number;
  witnessesToFind: number;
  pendingWitnessSelection: boolean;
  winner: TgWinner;
  finishedReason: string;
  resultMessage: string;
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
  murdererClueCardName: string;
  murdererMeansCardName: string;
}

export interface TgPlayerPrivateData {
  isMurderer: boolean;
  clueCardName: string;
  meansCardName: string;
  role: TgSecretRole;
  knownMurdererTeam: TgKnownRolePlayer[];
  knownMurdererClueCardName: string;
  knownMurdererMeansCardName: string;
  activeWitnessSelectionPrompt: TgWitnessSelectionPromptState;
}

interface TgForensicCardResource {
  causeCards: TgForensicCard[];
  locationCards: TgForensicCard[];
  otherCards: TgForensicCard[];
}

export interface TgCardResources {
  clueCards: TgCard[];
  meansCards: TgCard[];
  forensicCards: TgForensicCardResource;
}

export interface TgMurdererInfo {
  murderer: TgPlayer;
  clueCard: TgCard;
  meansCard: TgCard;
}

export interface TgWitnessPromptTarget {
  uid: string;
  name: string;
  role: TgSecretRole;
  hasActivePrompt: boolean;
  dismissible: boolean;
  creatorInitiated: boolean;
}

export interface TgModeratorPrivateData {
  witnessPromptTargets: TgWitnessPromptTarget[];
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
  accompliceCount: number;
  witnessCount: number;
  witnessesToFind: number;
}
