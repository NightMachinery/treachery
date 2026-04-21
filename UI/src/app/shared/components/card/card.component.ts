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

  defaultImage() {
    return this.means ? 'assets/means-bg.jpg' : 'assets/clue-bg.jpg';
  }

  imageSrc() {
    return this.card && this.card.imgUrl ? this.card.imgUrl : this.defaultImage();
  }

  onImageError(event: Event) {
    const img = event.target as HTMLImageElement;
    if (!img) {
      return;
    }

    const fallbackAttempt = img.dataset.fallbackAttempt || '';
    if (!fallbackAttempt && this.card && this.card.altImgUrl && img.src !== this.card.altImgUrl) {
      img.dataset.fallbackAttempt = 'alt';
      img.src = this.card.altImgUrl;
      return;
    }

    if (img.src !== this.defaultImage()) {
      img.dataset.fallbackAttempt = 'default';
      img.src = this.defaultImage();
    }
  }
}
