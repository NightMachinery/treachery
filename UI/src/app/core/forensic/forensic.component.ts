import { Component, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Subscription } from 'rxjs';
import { take } from 'rxjs/operators';
import { ChatApiService } from 'src/app/shared/api/chat/chat-api.service';
import { CardApiService } from './../../shared/api/card/card-api.service';
import { AuthService } from './../../shared/api/auth/auth.service';
import { GameApiService } from '../../shared/api/game/game-api.service';
import { ForensicApiService } from './../../shared/api/forensic/forensic-api.service';
import { SnackBarService } from './../../shared/api/snack-bar/snack-bar.service';
import { TgCard, TgForensicCard, TgForensicPrivateData, TgGame } from './../../shared/api/models/models';

@Component({
  selector: 'app-forensic',
  templateUrl: './forensic.component.html',
  styleUrls: ['./forensic.component.scss']
})
export class ForensicComponent implements OnInit, OnDestroy {
  selectedCauseCardName: string;
  selectedLocationCardName: string;
  selectedCauseCardOption: string;
  selectedLocationCardOption: string;
  selectedOtherCardOption: string;
  replaceCardName: string;
  loading = true;
  private subscription = new Subscription();

  constructor(
    private route: ActivatedRoute,
    public cardApi: CardApiService,
    public forensicApi: ForensicApiService,
    public gameApi: GameApiService,
    public auth: AuthService,
    public chatApi: ChatApiService,
    private snack: SnackBarService
  ) {}

  ngOnInit() {
    setTimeout(() => {
      this.loading = false;
    }, 10000);
    this.subscription.add(
      this.route.params.subscribe(async ({ gameId }) => {
        const roomAuth = this.route.snapshot.queryParamMap.get('roomAuth') || null;
        this.gameApi.setGameContext(gameId, roomAuth);
        await this.gameApi.refreshSnapshot();
      })
    );
    this.subscription.add(
      this.route.queryParams.subscribe(params => {
        const roomAuth = params.roomAuth || null;
        if (this.gameApi.gameId$.value) {
          this.gameApi.setGameContext(this.gameApi.gameId$.value, roomAuth);
        }
      })
    );
    this.subscription.add(
      this.gameApi.snapshot$.subscribe(snapshot => {
        if (!snapshot || !snapshot.game) {
          return;
        }
        this.syncRoute(snapshot).catch(error => console.warn('Failed to sync forensic route', error));
      })
    );
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
  }

  getChosenMeansCard(privateObj: TgForensicPrivateData) {
    return privateObj.murderer.clueCards.find((card: TgCard) => card.name === privateObj.murdererClueCardName);
  }

  getChosenClueCard(privateObj: TgForensicPrivateData) {
    return privateObj.murderer.meansCards.find((card: TgCard) => card.name === privateObj.murdererMeansCardName);
  }

  startGame() {
    this.forensicApi.startGame();
  }

  causeCardClick(card: TgForensicCard) {
    this.selectedCauseCardName = card.cardName;
  }

  locationCardClick(card: TgForensicCard) {
    if (this.selectedLocationCardName !== card.cardName) {
      this.selectedLocationCardName = card.cardName;
      this.selectedLocationCardOption = card.choices[0];
    }
  }

  nextCard(game: TgGame) {
    return game.otherCards.find(value => !value.selectedChoice);
  }

  async selectCauseCard() {
    this.gameApi.selectForensicCauseCard(await this.cardApi.getCauseCard(this.selectedCauseCardName, this.selectedCauseCardOption));
    this.selectedCauseCardName = null;
    this.selectedCauseCardOption = null;
  }

  async selectLocationCard() {
    this.gameApi.selectForensicLocationCard(
      await this.cardApi.getLocationCard(this.selectedLocationCardName, this.selectedLocationCardOption)
    );
    this.selectedLocationCardName = null;
    this.selectedLocationCardOption = null;
  }

  async selectNextOtherCard() {
    this.gameApi.game$.pipe(take(1)).subscribe(game => {
      if (this.toReplace(game) && !this.replaceCardName) {
        this.snack.error('Please select a card to replace first!');
        return;
      }
      const nextCard = this.nextCard(game);
      this.gameApi.selectNextForensicOtherCard(
        {
          ...nextCard,
          selectedChoice: this.selectedOtherCardOption
        },
        this.replaceCardName
      );
      this.selectedOtherCardOption = null;
      this.replaceCardName = null;
    });
  }

  selectReplaceCard = (card: TgForensicCard) => {
    this.replaceCardName = card.cardName;
  };

  toReplace(game: TgGame) {
    return game.otherCards.filter(card => card.selectedChoice).length >= 4 && game.otherCards.filter(card => card.replaced).length < 2;
  }

  canSelectNextOtherCard(game: TgGame) {
    return !!this.selectedOtherCardOption && (!this.toReplace(game) || !!this.replaceCardName);
  }

  showNextOtherCard(game: TgGame) {
    return game.causeCard && game.locationCard && game.otherCards.filter(card => card.selectedChoice).length < 6;
  }

  waitingToEnd(game: TgGame) {
    return game.causeCard && game.locationCard && game.otherCards.filter(card => card.selectedChoice).length >= 6;
  }

  endGame() {
    this.forensicApi.endGame();
  }

  async copyMigrateLink() {
    const link = await this.gameApi.createMigrateLink();
    if (!link) {
      this.snack.error('Could not create a migrate-device link.');
      return;
    }
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(link);
    } else {
      window.prompt('Copy this migrate-device link', link);
    }
  }

  private async syncRoute(snapshot) {
    if (!snapshot.game.startedOn) {
      await this.gameApi.navigateTo('join', snapshot.game.gameId);
      return;
    }
    if (!snapshot.viewer.isScientist) {
      if (snapshot.viewer.isParticipant && snapshot.viewer.role === 'player') {
        await this.gameApi.navigateTo('play', snapshot.game.gameId);
      } else {
        await this.gameApi.navigateTo('observe', snapshot.game.gameId);
      }
    }
  }
}
