import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { combineLatest, firstValueFrom, Observable, of } from 'rxjs';
import { catchError, shareReplay, switchMap } from 'rxjs/operators';
import { TgCrimePackResource, TgCurrentPackResources, TgForensicCard, TgHintPackResource, TgWordpackCatalog } from './../models/models';
import { GameApiService } from '../game/game-api.service';

@Injectable({
  providedIn: 'root',
})
export class CardApiService {
  public catalog$: Observable<TgWordpackCatalog>;
  public cards$: Observable<TgCurrentPackResources>;

  constructor(
    private http: HttpClient,
    private gameApi: GameApiService,
  ) {
    this.catalog$ = this.http.get<TgWordpackCatalog>('/api/wordpacks/catalog').pipe(shareReplay(1));
    this.cards$ = this.gameApi.game$.pipe(
      switchMap((game) => {
        if (!game) {
          return of(null);
        }
        return combineLatest([
          this.http.get<TgCrimePackResource>(
            `/api/wordpacks/crime/${encodeURIComponent(game.crimePackId)}?language=${encodeURIComponent(game.crimePackLanguage || '')}&assetSetId=${encodeURIComponent(
              game.crimePackAssetSetId || '',
            )}`,
          ),
          this.http.get<TgHintPackResource>(
            `/api/wordpacks/hint/${encodeURIComponent(game.hintPackId)}?language=${encodeURIComponent(game.hintPackLanguage || '')}`,
          ),
        ]);
      }),
      catchError(() => of(null)),
      switchMap((resources) => {
        if (!resources) {
          return of(null);
        }
        const [crimePack, hintPack] = resources;
        return of({
          crimePack,
          hintPack,
          clueCards: crimePack.clueCards,
          meansCards: crimePack.meansCards,
          forensicCards: hintPack.forensicCards,
        } as TgCurrentPackResources);
      }),
      shareReplay(1),
    );
  }

  async getCatalogSnapshot(): Promise<TgWordpackCatalog> {
    return firstValueFrom(this.catalog$);
  }

  async getCardsSnapshot(): Promise<TgCurrentPackResources> {
    return firstValueFrom(this.cards$);
  }

  async getCauseCard(cardId: string, selectedChoiceId: string): Promise<TgForensicCard> {
    const cards = await this.getCardsSnapshot();
    const base = cards?.forensicCards?.causeCards?.find((card) => card.cardId === cardId);
    return {
      ...base,
      selectedChoiceId,
      selectedChoice: base?.choices?.[base.choiceIds?.indexOf(selectedChoiceId)] || '',
    } as TgForensicCard;
  }
}
