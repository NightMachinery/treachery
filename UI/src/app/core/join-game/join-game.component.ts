import { ActivatedRoute } from '@angular/router';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { Subscription } from 'rxjs';
import { AuthService } from './../../shared/api/auth/auth.service';
import { CardApiService } from './../../shared/api/card/card-api.service';
import { GameApiService } from '../../shared/api/game/game-api.service';
import { ForensicApiService } from '../../shared/api/forensic/forensic-api.service';
import { TgGame, TgGameSettingsInput, TgParticipant } from '../../shared/api/models/models';
import { SnackBarService } from '../../shared/api/snack-bar/snack-bar.service';

@Component({
  selector: 'app-join-game',
  standalone: false,
  templateUrl: './join-game.component.html',
  styleUrls: ['./join-game.component.scss']
})
export class JoinGameComponent implements OnInit, OnDestroy {
  gameId: string;
  roomAuth: string;
  loading = true;
  missingGame = false;
  settings: TgGameSettingsInput = this.defaultSettings();
  settingsDirty = false;
  private subscription = new Subscription();
  private autoJoiningObserver = false;
  private clueDeckSize = 0;
  private meansDeckSize = 0;

  constructor(
    private route: ActivatedRoute,
    public gameApi: GameApiService,
    private auth: AuthService,
    public forensicApi: ForensicApiService,
    private snack: SnackBarService,
    private cardApi: CardApiService
  ) {}

  ngOnInit() {
    this.subscription.add(
      this.cardApi.cards$.subscribe(cards => {
        this.clueDeckSize = cards.clueCards.length;
        this.meansDeckSize = cards.meansCards.length;
      })
    );

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
        if (!snapshot.game.startedOn && !this.settingsDirty) {
          this.syncSettingsFromGame(snapshot.game);
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
    const snapshot = this.gameApi.getCurrentSnapshot();
    if (!snapshot || !snapshot.game) {
      return;
    }
    const validation = this.getStartValidationMessage(snapshot.game, snapshot.participants);
    if (validation) {
      this.snack.error(validation);
      return;
    }
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

  handleMeansCardCountChange() {
    this.settingsDirty = true;
    this.settings.meansCardsPerPlayer = this.normalizePositiveNumber(this.settings.meansCardsPerPlayer, 4);
    if (this.settings.linkClueCountToMeans) {
      this.settings.clueCardsPerPlayer = this.settings.meansCardsPerPlayer;
    }
  }

  handleClueCardCountChange() {
    this.settingsDirty = true;
    this.settings.clueCardsPerPlayer = this.normalizePositiveNumber(this.settings.clueCardsPerPlayer, this.settings.meansCardsPerPlayer || 4);
  }

  handleLinkClueCountChange() {
    this.settingsDirty = true;
    if (this.settings.linkClueCountToMeans) {
      this.settings.clueCardsPerPlayer = this.settings.meansCardsPerPlayer;
    }
  }

  handleMeansCluesTextOnlyChange() {
    this.settingsDirty = true;
  }

  handleRoleCountChange() {
    this.settingsDirty = true;
    this.settings.accompliceCount = this.normalizeBoundedNumber(this.settings.accompliceCount, 0, 10, 0);
    this.settings.witnessCount = this.normalizeBoundedNumber(this.settings.witnessCount, 0, 10, 0);
    if (this.settings.witnessCount === 0) {
      this.settings.witnessesToFind = 0;
    } else {
      this.settings.witnessesToFind = this.normalizeBoundedNumber(this.settings.witnessesToFind, 1, this.settings.witnessCount, 1);
    }
  }

  async saveSettings() {
    this.handleMeansCardCountChange();
    if (!this.settings.linkClueCountToMeans) {
      this.handleClueCardCountChange();
    }
    this.handleRoleCountChange();
    await this.gameApi.updateGameSettings({ ...this.settings });
    this.settingsDirty = false;
  }

  async resetSettings(game: TgGame) {
    this.syncSettingsFromGame(game);
    this.settingsDirty = false;
  }

  getStartValidationMessage(game: TgGame, participants: TgParticipant[]): string {
    const playerCount = (participants || []).filter(participant => participant.role === 'player').length;
    if (playerCount < 4) {
      return 'Need at least 4 lobby players to start.';
    }
    const suspects = playerCount - 1;
    if (suspects < 3) {
      return 'Need at least 3 suspects after selecting the forensic scientist.';
    }
    if (1 + game.accompliceCount + game.witnessCount > suspects) {
      return 'Too many special roles for the available suspects.';
    }
    if (game.witnessCount === 0 && game.witnessesToFind !== 0) {
      return 'Witnesses to find must be 0 when witness count is 0.';
    }
    if (game.witnessCount > 0 && (game.witnessesToFind < 1 || game.witnessesToFind > game.witnessCount)) {
      return 'Witnesses to find must be between 1 and the witness count.';
    }
    if (this.clueDeckSize && suspects * game.clueCardsPerPlayer > this.clueDeckSize) {
      return `Not enough clue cards for ${game.clueCardsPerPlayer} evidence cards per suspect.`;
    }
    if (this.meansDeckSize && suspects * game.meansCardsPerPlayer > this.meansDeckSize) {
      return `Not enough means cards for ${game.meansCardsPerPlayer} means cards per suspect.`;
    }
    return '';
  }

  private defaultSettings(): TgGameSettingsInput {
    return {
      meansCardsPerPlayer: 4,
      clueCardsPerPlayer: 4,
      linkClueCountToMeans: true,
      meansCluesTextOnly: false,
      accompliceCount: 0,
      witnessCount: 0,
      witnessesToFind: 0
    };
  }

  private syncSettingsFromGame(game: TgGame) {
    this.settings = {
      meansCardsPerPlayer: game.meansCardsPerPlayer,
      clueCardsPerPlayer: game.clueCardsPerPlayer,
      linkClueCountToMeans: game.linkClueCountToMeans,
      meansCluesTextOnly: game.meansCluesTextOnly,
      accompliceCount: game.accompliceCount,
      witnessCount: game.witnessCount,
      witnessesToFind: game.witnessesToFind
    };
  }

  private normalizePositiveNumber(value: number, fallback: number) {
    const parsed = Number(value);
    return parsed >= 1 ? Math.floor(parsed) : fallback;
  }

  private normalizeBoundedNumber(value: number, min: number, max: number, fallback: number) {
    const parsed = Math.floor(Number(value));
    if (Number.isNaN(parsed)) {
      return fallback;
    }
    return Math.min(max, Math.max(min, parsed));
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
