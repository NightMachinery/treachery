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

export interface TgGameSnapshot {
  game: TgGame;
  participants: TgParticipant[];
  players: TgPlayer[];
  guesses: TgGuess[];
  messages: TgMessage[];
  viewer: TgViewer;
  playerPrivateData: TgPlayerPrivateData;
  forensicPrivateData: TgForensicPrivateData;
}

export interface TgAuthUser {
  uid: string;
  token: string;
  displayName: string;
}
