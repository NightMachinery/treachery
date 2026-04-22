import { TgCard, TgPlayer } from 'src/app/shared/api/models/models';

export function findClueCard(player: TgPlayer, cardId: string): TgCard {
  return player && player.clueCards ? player.clueCards.find(card => card.id === cardId) : null;
}

export function findMeansCard(player: TgPlayer, cardId: string): TgCard {
  return player && player.meansCards ? player.meansCards.find(card => card.id === cardId) : null;
}
