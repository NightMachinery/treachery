import { ActivatedRoute } from '@angular/router';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { Subscription } from 'rxjs';
import { AuthService } from './../../shared/api/auth/auth.service';
import { GameApiService } from '../../shared/api/game/game-api.service';
import { ForensicApiService } from '../../shared/api/forensic/forensic-api.service';
import { SnackBarService } from '../../shared/api/snack-bar/snack-bar.service';

@Component({
  selector: 'app-join-game',
  templateUrl: './join-game.component.html',
  styleUrls: ['./join-game.component.scss']
})
export class JoinGameComponent implements OnInit, OnDestroy {
  gameId: string;
  roomAuth: string;
  loading = true;
  missingGame = false;
  private subscription = new Subscription();
  private autoJoiningObserver = false;

  constructor(
    private route: ActivatedRoute,
    public gameApi: GameApiService,
    private auth: AuthService,
    public forensicApi: ForensicApiService,
    private snack: SnackBarService
  ) {}

  ngOnInit() {
    this.subscription.add(
      this.route.params.subscribe(async ({ gameId }) => {
        this.gameId = gameId;
        this.roomAuth = this.route.snapshot.queryParamMap.get('roomAuth') || null;
        this.gameApi.setGameContext(this.gameId, this.roomAuth);
        this.loading = true;
        this.missingGame = !(await this.gameApi.gameExists(this.gameId));
        this.loading = false;
        if (!this.missingGame) {
          await this.gameApi.refreshSnapshot();
        }
      })
    );

    this.subscription.add(
      this.route.queryParams.subscribe(params => {
        this.roomAuth = params.roomAuth || null;
        if (this.gameId) {
          this.gameApi.setGameContext(this.gameId, this.roomAuth);
        }
      })
    );

    this.subscription.add(
      this.gameApi.snapshot$.subscribe(snapshot => {
        if (!snapshot || !snapshot.game) {
          return;
        }
        this.handleSnapshot(snapshot).catch(error => console.warn('Failed to handle join snapshot', error));
      })
    );
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
  }

  async joinAsPlayer() {
    if (!(await this.auth.ensureDisplayName(this.gameId, this.roomAuth))) {
      return;
    }
    await this.gameApi.joinGame(this.gameId, 'player');
  }

  async observeStartedGame() {
    if (!(await this.auth.ensureDisplayName(this.gameId, this.roomAuth))) {
      return;
    }
    await this.gameApi.joinGame(this.gameId, 'observer');
    await this.gameApi.navigateTo('observe', this.gameId);
  }

  async startGame() {
    await this.forensicApi.startGame();
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

  private async handleSnapshot(snapshot) {
    const { game, viewer } = snapshot;
    if (!game.startedOn) {
      return;
    }

    if (viewer.isScientist) {
      await this.gameApi.navigateTo('forensic', game.gameId);
      return;
    }
    if (viewer.isParticipant && viewer.role === 'player') {
      await this.gameApi.navigateTo('play', game.gameId);
      return;
    }
    if (viewer.isParticipant && viewer.role === 'observer') {
      await this.gameApi.navigateTo('observe', game.gameId);
      return;
    }
    if (!this.autoJoiningObserver && this.auth.getStoredDisplayName()) {
      this.autoJoiningObserver = true;
      try {
        await this.gameApi.joinGame(game.gameId, 'observer');
        await this.gameApi.navigateTo('observe', game.gameId);
      } finally {
        this.autoJoiningObserver = false;
      }
    }
  }
}
