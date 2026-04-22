import { Component, OnInit } from '@angular/core';
import { Observable, of } from 'rxjs';
import { map, switchMap } from 'rxjs/operators';
import { GameApiService } from 'src/app/shared/api/game/game-api.service';
import { findClueCard, findMeansCard } from 'src/app/shared/utils/find';
import { ForensicApiService } from './../../../shared/api/forensic/forensic-api.service';
import { TgMurdererInfo } from './../../../shared/api/models/models';

@Component({
  selector: 'tg-murderer-info',
  standalone: false,
  templateUrl: './murderer-info.component.html',
  styleUrls: ['./murderer-info.component.scss']
})
export class MurdererInfoComponent implements OnInit {
  murderer$: Observable<TgMurdererInfo>;

  constructor(public forensicApi: ForensicApiService, public gameApi: GameApiService) {
    this.murderer$ = forensicApi.forensicPrivateData$.pipe(
      switchMap(forensicPrivateData => {
        if (forensicPrivateData && forensicPrivateData.murderer) {
          return of({
            murderer: forensicPrivateData.murderer,
            clueCard: findClueCard(forensicPrivateData.murderer, forensicPrivateData.murdererClueCardId),
            meansCard: findMeansCard(forensicPrivateData.murderer, forensicPrivateData.murdererMeansCardId)
          });
        }

        return this.gameApi.game$.pipe(
          switchMap(game => {
            if (game && game.murdererUid) {
              return this.gameApi.players$.pipe(
                map(players => {
                  const murderer = players.find(p => p.uid === game.murdererUid);
                  return {
                    murderer,
                    clueCard: findClueCard(murderer, game.murdererClueCardId),
                    meansCard: findMeansCard(murderer, game.murdererMeansCardId)
                  };
                })
              );
            }
            return of(null);
          })
        );
      })
    );
  }

  ngOnInit(): void {}
}
