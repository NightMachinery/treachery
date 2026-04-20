import { ActivatedRoute, Router } from '@angular/router';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { Subscription } from 'rxjs';
import { ChatApiService } from 'src/app/shared/api/chat/chat-api.service';
import { TgGameSnapshot, TgPartialGuess } from './../../shared/api/models/models';
import { GameApiService } from '../../shared/api/game/game-api.service';
import { AuthService } from './../../shared/api/auth/auth.service';
import { ForensicApiService } from './../../shared/api/forensic/forensic-api.service';
import { SnackBarService } from './../../shared/api/snack-bar/snack-bar.service';

@Component({
  selector: 'app-game',
  templateUrl: './game.component.html',
  styleUrls: ['./game.component.scss']
})
export class GameComponent implements OnInit, OnDestroy {
  guess: TgPartialGuess = {} as TgPartialGuess;
  loading = true;
  private subscription = new Subscription();

  constructor(
    private route: ActivatedRoute,
    public gameApi: GameApiService,
    private auth: AuthService,
    private router: Router,
    public chatApi: ChatApiService,
    public forensicApi: ForensicApiService,
    private snack: SnackBarService
  ) {}

  ngOnInit() {
    this.subscription.add(
      this.route.params.subscribe(async ({ gameId }) => {
        const roomAuth = this.route.snapshot.queryParamMap.get('roomAuth') || null;
        this.gameApi.setGameContext(gameId, roomAuth);
        this.loading = true;
        if (!(await this.gameApi.gameExists(gameId))) {
          this.router.navigateByUrl('/');
          return;
        }
        await this.gameApi.refreshSnapshot();
        this.loading = false;
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
        this.syncRoute(snapshot).catch(error => console.warn('Failed to sync game route', error));
      })
    );
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
  }

  isObserverRoute() {
    return this.router.url.startsWith('/observe/');
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

  async endGame() {
    await this.forensicApi.endGame();
  }

  async restartGame() {
    await this.forensicApi.restartGame();
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

  private async syncRoute(snapshot: TgGameSnapshot) {
    const { game, viewer } = snapshot;
    if (!game.startedOn) {
      await this.gameApi.navigateTo('join', game.gameId);
      return;
    }
    if (viewer.isScientist) {
      await this.gameApi.navigateTo('forensic', game.gameId);
      return;
    }
    if (viewer.isParticipant && viewer.role === 'player') {
      if (this.isObserverRoute()) {
        await this.gameApi.navigateTo('play', game.gameId);
      }
      return;
    }
    if (!viewer.isParticipant) {
      if (!this.auth.getStoredDisplayName()) {
        await this.gameApi.navigateTo('join', game.gameId);
        return;
      }
      await this.gameApi.joinGame(game.gameId, 'observer');
      return;
    }
    if (viewer.role === 'observer' && !this.isObserverRoute()) {
      await this.gameApi.navigateTo('observe', game.gameId);
    }
  }
}
