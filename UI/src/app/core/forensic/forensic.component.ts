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
import { TgCard, TgForensicCard, TgForensicPrivateData, TgGame, TgGuess, TgParticipant, TgPlayer } from './../../shared/api/models/models';
import { copyTextToClipboard } from '../../shared/utils/clipboard';

@Component({
  selector: 'app-forensic',
  standalone: false,
  templateUrl: './forensic.component.html',
  styleUrls: ['./forensic.component.scss'],
})
export class ForensicComponent implements OnInit, OnDestroy {
  selectedCauseCardId: string;
  selectedLocationCard: TgForensicCard;
  selectedCauseCardOptionId: string;
  selectedLocationCardOptionId: string;
  selectedOtherCardOptionId: string;
  replaceCardId: string;
  loading = true;
  private subscription = new Subscription();

  constructor(
    private route: ActivatedRoute,
    public cardApi: CardApiService,
    public forensicApi: ForensicApiService,
    public gameApi: GameApiService,
    public auth: AuthService,
    public chatApi: ChatApiService,
    private snack: SnackBarService,
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
      }),
    );
    this.subscription.add(
      this.route.queryParams.subscribe((params) => {
        const roomAuth = params.roomAuth || null;
        if (this.gameApi.gameId$.value) {
          this.gameApi.setGameContext(this.gameApi.gameId$.value, roomAuth);
        }
      }),
    );
    this.subscription.add(
      this.gameApi.snapshot$.subscribe((snapshot) => {
        if (!snapshot || !snapshot.game) {
          return;
        }
        this.syncRoute(snapshot).catch((error) => console.warn('Failed to sync forensic route', error));
      }),
    );
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
  }

  getChosenMeansCard(privateObj: TgForensicPrivateData) {
    return privateObj.murderer.meansCards.find((card: TgCard) => card.id === privateObj.murdererMeansCardId);
  }

  getChosenClueCard(privateObj: TgForensicPrivateData) {
    return privateObj.murderer.clueCards.find((card: TgCard) => card.id === privateObj.murdererClueCardId);
  }

  startGame() {
    this.forensicApi.startGame();
  }

  causeCardClick(card: TgForensicCard) {
    if (this.selectedCauseCardId !== card.cardId) {
      this.selectedCauseCardId = card.cardId;
      this.selectedCauseCardOptionId = card.choiceIds[0];
    }
  }

  locationCardClick(card: TgForensicCard) {
    if (this.selectedLocationCard !== card) {
      this.selectedLocationCard = card;
      this.selectedLocationCardOptionId = card.choiceIds[0];
    }
  }

  nextCard(game: TgGame) {
    return game.otherCards.find((value) => !value.selectedChoiceId);
  }

  async selectCauseCard() {
    await this.gameApi.selectForensicCauseCard(await this.cardApi.getCauseCard(this.selectedCauseCardId, this.selectedCauseCardOptionId));
    this.selectedCauseCardId = null;
    this.selectedCauseCardOptionId = null;
  }

  async selectLocationCard() {
    await this.gameApi.selectForensicLocationCard({
      ...this.selectedLocationCard,
      selectedChoiceId: this.selectedLocationCardOptionId,
      selectedChoice:
        this.selectedLocationCard?.choices?.[this.selectedLocationCard.choiceIds.indexOf(this.selectedLocationCardOptionId)] || '',
    });
    this.selectedLocationCard = null;
    this.selectedLocationCardOptionId = null;
  }

  async selectNextOtherCard() {
    this.gameApi.game$.pipe(take(1)).subscribe((game) => {
      if (this.toReplace(game) && !this.replaceCardId) {
        this.snack.error('Please select a card to replace first!');
        return;
      }
      const nextCard = this.nextCard(game);
      this.gameApi.selectNextForensicOtherCard(
        {
          ...nextCard,
          selectedChoiceId: this.selectedOtherCardOptionId,
          selectedChoice: nextCard?.choices?.[nextCard.choiceIds.indexOf(this.selectedOtherCardOptionId)] || '',
        },
        this.replaceCardId,
      );
      this.selectedOtherCardOptionId = null;
      this.replaceCardId = null;
    });
  }

  selectReplaceCard = (card: TgForensicCard) => {
    this.replaceCardId = card.cardId;
  };

  toReplace(game: TgGame) {
    return (
      game.otherCards.filter((card) => card.selectedChoiceId).length >= 4 && game.otherCards.filter((card) => card.replaced).length < 2
    );
  }

  canSelectNextOtherCard(game: TgGame) {
    return !!this.selectedOtherCardOptionId && (!this.toReplace(game) || !!this.replaceCardId);
  }

  showNextOtherCard(game: TgGame) {
    return game.causeCard && game.locationCard && game.otherCards.filter((card) => card.selectedChoiceId).length < 6;
  }

  waitingToEnd(game: TgGame) {
    return game.causeCard && game.locationCard && game.otherCards.filter((card) => card.selectedChoiceId).length >= 6;
  }

  getResultHeadline(game: TgGame) {
    switch (game.winner) {
      case 'investigatorTeam':
        return 'Case solved';
      case 'murdererTeam':
        return 'The killer got away';
      default:
        return 'Round concluded';
    }
  }

  getWinnerLabel(winner: string) {
    switch (winner) {
      case 'investigatorTeam':
        return 'Investigator team wins';
      case 'murdererTeam':
        return 'Murderer team wins';
      default:
        return 'Game ended';
    }
  }

  getFinishedReasonLabel(reason: string) {
    switch (reason) {
      case 'correct-guess':
        return 'Correct accusation';
      case 'all-guesses-used':
        return 'Good-team guesses exhausted';
      case 'witnesses-found':
        return 'Witnesses identified';
      case 'witnesses-missed':
        return 'Witnesses escaped';
      case 'moderator-ended':
        return 'Ended by moderator';
      default:
        return 'Round complete';
    }
  }

  getWitnessModeLabel(game: TgGame) {
    if (!game.witnessCount) {
      return 'Off';
    }
    return `${game.witnessesToFind}/${game.witnessCount} to find`;
  }

  sortedGuesses(guesses: TgGuess[]) {
    return [...(guesses || [])].sort((a, b) => {
      const aTime = a.createdTimestamp ? Date.parse(a.createdTimestamp) : 0;
      const bTime = b.createdTimestamp ? Date.parse(b.createdTimestamp) : 0;
      return bTime - aTime;
    });
  }

  participantName(uid: string, participantsDict: Map<string, TgParticipant>) {
    return participantsDict?.get(uid)?.name || 'Someone';
  }

  playerName(uid: string, playersDict: Map<string, TgPlayer>) {
    return playersDict?.get(uid)?.name || 'Unknown suspect';
  }

  trackByGuess(index: number, guess: TgGuess) {
    return `${guess.guessedByUid}-${guess.murdererUid}-${guess.meansCardId}-${guess.clueCardId}-${guess.createdTimestamp || index}`;
  }

  endGame() {
    this.forensicApi.endGame();
  }

  restartGame() {
    this.forensicApi.restartGame();
  }

  async copyMigrateLink() {
    const link = await this.gameApi.createMigrateLink();
    if (!link) {
      this.snack.error('Could not create a migrate-device link.');
      return;
    }
    const copied = await copyTextToClipboard(link);
    if (copied) {
      this.snack.success('Migrate-device link copied.');
      return;
    }
    this.snack.error('Could not copy the migrate-device link.');
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
