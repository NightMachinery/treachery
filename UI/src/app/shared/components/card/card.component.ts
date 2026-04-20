import { Component, Input } from '@angular/core';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { GameApiService } from '../../api/game/game-api.service';
import { TgCard } from '../../api/models/models';

@Component({
  selector: 'app-card',
  templateUrl: './card.component.html',
  styleUrls: ['./card.component.scss']
})
export class CardComponent {
  @Input() card: TgCard;
  @Input() active = false;
  @Input() means: boolean;
  textOnlyMode$: Observable<boolean>;

  constructor(public gameApi: GameApiService) {
    this.textOnlyMode$ = this.gameApi.game$.pipe(map(game => !!(game && game.meansCluesTextOnly)));
  }
}
