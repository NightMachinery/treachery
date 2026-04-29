import {
  AfterViewChecked,
  AfterViewInit,
  Component,
  ElementRef,
  Input,
  NgZone,
  OnChanges,
  OnDestroy,
  SimpleChanges,
  ViewChild,
} from '@angular/core';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { GameApiService } from '../../api/game/game-api.service';
import { TgCard } from '../../api/models/models';

@Component({
  selector: 'app-card',
  standalone: false,
  templateUrl: './card.component.html',
  styleUrls: ['./card.component.scss'],
})
export class CardComponent implements AfterViewInit, AfterViewChecked, OnChanges, OnDestroy {
  @Input() card: TgCard;
  @Input() active = false;
  @Input() means: boolean;
  @Input() subdued = false;
  @ViewChild('cardName') cardNameElement?: ElementRef<HTMLElement>;
  textOnlyMode$: Observable<boolean>;
  labelCompressed = false;

  private fitQueued = false;
  private lastFitSignature = '';
  private resizeObserver?: ResizeObserver;
  private observedLabel?: HTMLElement;

  constructor(
    public gameApi: GameApiService,
    private ngZone: NgZone,
  ) {
    this.textOnlyMode$ = this.gameApi.game$.pipe(map((game) => !!(game && game.meansCluesTextOnly)));
  }

  ngAfterViewInit() {
    this.observeLabelSize();
    this.scheduleLabelFit(true);
  }

  ngAfterViewChecked() {
    this.observeLabelSize();
    this.scheduleLabelFit();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes.card) {
      this.lastFitSignature = '';
      this.scheduleLabelFit(true);
    }
  }

  ngOnDestroy() {
    this.resizeObserver?.disconnect();
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
      altImgUrl: '',
    };
    this.lastFitSignature = '';
    this.scheduleLabelFit(true);
  }

  fitCardLabel() {
    const label = this.cardNameElement?.nativeElement;
    if (!label) {
      return;
    }

    const availableWidth = label.clientWidth;
    const availableHeight = label.clientHeight;
    const signature = `${this.card?.name || ''}|${availableWidth}x${availableHeight}`;
    if (signature === this.lastFitSignature) {
      return;
    }

    this.lastFitSignature = signature;
    label.style.setProperty('--tg-card-name-scale', '1');

    if (!availableWidth || !availableHeight || this.labelFits(label)) {
      this.labelCompressed = false;
      label.classList.remove('card-name--compressed');
      return;
    }

    let low = 0.58;
    let high = 1;
    for (let i = 0; i < 7; i++) {
      const mid = (low + high) / 2;
      label.style.setProperty('--tg-card-name-scale', mid.toFixed(3));
      if (this.labelFits(label)) {
        low = mid;
      } else {
        high = mid;
      }
    }

    const fittedScale = low;
    label.style.setProperty('--tg-card-name-scale', fittedScale.toFixed(3));
    this.labelCompressed = fittedScale < 0.985;
    label.classList.toggle('card-name--compressed', this.labelCompressed);
  }

  private labelFits(label: HTMLElement) {
    return label.scrollWidth <= label.clientWidth + 1 && label.scrollHeight <= label.clientHeight + 1;
  }

  private observeLabelSize() {
    const label = this.cardNameElement?.nativeElement;
    if (!label || typeof ResizeObserver === 'undefined' || this.observedLabel === label) {
      return;
    }

    this.resizeObserver?.disconnect();
    this.observedLabel = label;
    this.ngZone.runOutsideAngular(() => {
      this.resizeObserver = new ResizeObserver(() => {
        this.lastFitSignature = '';
        this.scheduleLabelFit(true);
      });
      this.resizeObserver.observe(label);
    });
  }

  private scheduleLabelFit(force = false) {
    if (force) {
      this.lastFitSignature = '';
    }
    if (this.fitQueued || !this.cardNameElement) {
      return;
    }

    this.fitQueued = true;
    this.ngZone.runOutsideAngular(() => {
      requestAnimationFrame(() => {
        this.fitQueued = false;
        this.fitCardLabel();
      });
    });
  }
}
