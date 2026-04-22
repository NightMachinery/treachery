import { Component, Input } from '@angular/core';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { GameApiService } from '../../api/game/game-api.service';
import { TgCard } from '../../api/models/models';

@Component({
  selector: 'app-card',
  standalone: false,
  templateUrl: './card.component.html',
  styleUrls: ['./card.component.scss']
})
export class CardComponent {
  @Input() card: TgCard;
  @Input() active = false;
  @Input() means: boolean;
  @Input() subdued = false;
  textOnlyMode$: Observable<boolean>;

  constructor(public gameApi: GameApiService) {
    this.textOnlyMode$ = this.gameApi.game$.pipe(map(game => !!(game && game.meansCluesTextOnly)));
  }

  get shouldShowTextOnly() {
    return !this.card?.hasImage;
  }

  onImageError(event: Event) {
    const img = event.target as HTMLImageElement;
    if (!img || !this.card) {
      return;
    }

    const fallbackAttempt = img.dataset.fallbackAttempt || '';
    if (!fallbackAttempt && this.card.altImgUrl && img.src !== this.card.altImgUrl) {
      img.dataset.fallbackAttempt = 'alt';
      img.src = this.card.altImgUrl;
      return;
    }

    this.card = {
      ...this.card,
      hasImage: false,
      imgUrl: '',
      altImgUrl: ''
    };
  }
}
