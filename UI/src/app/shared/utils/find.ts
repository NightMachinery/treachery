import { TgCard, TgPlayer } from 'src/app/shared/api/models/models';

export function findClueCard(player: TgPlayer, cardName: string): TgCard {
  return player && player.clueCards ? player.clueCards.find(card => card.name === cardName) : null;
}

export function findMeansCard(player: TgPlayer, cardName: string): TgCard {
  return player && player.meansCards ? player.meansCards.find(card => card.name === cardName) : null;
}
