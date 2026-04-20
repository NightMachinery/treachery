import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { shareReplay } from 'rxjs/operators';
import { TgCardResources, TgForensicCard } from './../models/models';

@Injectable({
  providedIn: 'root'
})
export class CardApiService {
  public cards$: Observable<TgCardResources>;

  constructor(private http: HttpClient) {
    this.cards$ = this.http.get<TgCardResources>('assets/cards.json').pipe(shareReplay(1));
  }

  async getCardsSnapshot(): Promise<TgCardResources> {
    return this.cards$.toPromise<TgCardResources>();
  }

  async getCauseCard(cardName: string, selectedChoice: string): Promise<TgForensicCard> {
    const cards = await this.getCardsSnapshot();

    return {
      ...cards.forensicCards.causeCards.find(card => card.cardName === cardName),
      selectedChoice
    } as TgForensicCard;
  }
}
