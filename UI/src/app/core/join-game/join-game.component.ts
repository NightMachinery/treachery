import { ActivatedRoute } from '@angular/router';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { Subscription } from 'rxjs';
import { AuthService } from './../../shared/api/auth/auth.service';
import { CardApiService } from './../../shared/api/card/card-api.service';
import { GameApiService } from '../../shared/api/game/game-api.service';
import { ForensicApiService } from '../../shared/api/forensic/forensic-api.service';
import {
  TgCrimePackCatalogEntry,
  TgGame,
  TgGameSettingsInput,
  TgHintPackCatalogEntry,
  TgParticipant,
  TgWordpackCatalog,
} from '../../shared/api/models/models';
import { SnackBarService } from '../../shared/api/snack-bar/snack-bar.service';
import { copyTextToClipboard } from '../../shared/utils/clipboard';

@Component({
  selector: 'app-join-game',
  standalone: false,
  templateUrl: './join-game.component.html',
  styleUrls: ['./join-game.component.scss'],
})
export class JoinGameComponent implements OnInit, OnDestroy {
  gameId: string;
  roomAuth: string;
  loading = true;
  missingGame = false;
  settings: TgGameSettingsInput = this.defaultSettings();
  settingsDirty = false;
  settingsSaving = false;
  settingsSaveError = '';
  settingsSaved = false;
  devModeEnabled = false;
  addingBots = false;
  private readonly settingsAutoSaveDelayMs = 500;
  private readonly settingsRetryDelayMs = 1500;
  private settingsSaveTimer: ReturnType<typeof setTimeout> = null;
  private settingsChangeVersion = 0;
  private subscription = new Subscription();
  private autoJoiningObserver = false;
  private catalog: TgWordpackCatalog = null;

  constructor(
    private route: ActivatedRoute,
    public gameApi: GameApiService,
    private auth: AuthService,
    public forensicApi: ForensicApiService,
    private snack: SnackBarService,
    public cardApi: CardApiService,
  ) {}

  ngOnInit() {
    this.subscription.add(this.cardApi.catalog$.subscribe((catalog) => (this.catalog = catalog)));

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
      }),
    );

    this.subscription.add(
      this.route.queryParams.subscribe((params) => {
        this.roomAuth = params.roomAuth || null;
        if (this.gameId) {
          this.gameApi.setGameContext(this.gameId, this.roomAuth);
        }
      }),
    );

    this.subscription.add(
      this.gameApi.snapshot$.subscribe((snapshot) => {
        if (!snapshot || !snapshot.game) {
          return;
        }
        if (!snapshot.game.startedOn && !this.hasSettingsAutoSaveWork) {
          this.syncSettingsFromGame(snapshot.game);
        }
        this.handleSnapshot(snapshot).catch((error) => console.warn('Failed to handle join snapshot', error));
      }),
    );
  }

  ngOnDestroy(): void {
    this.clearSettingsSaveTimer();
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
    if (this.hasSettingsAutoSaveWork) {
      this.snack.error('Lobby settings are still saving.');
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
    const copied = await copyTextToClipboard(link);
    if (copied) {
      this.snack.success('Migrate-device link copied.');
      return;
    }
    this.snack.error('Could not copy the migrate-device link.');
  }

  async addBots() {
    const snapshot = this.gameApi.getCurrentSnapshot();
    if (!snapshot || !snapshot.game) {
      return;
    }
    const playerCount = (snapshot.participants || []).filter((p) => p.role === 'player').length;
    const count = Math.max(0, 4 - playerCount);

    if (count <= 0) {
      this.snack.success('Already have enough players to start.');
      return;
    }

    this.addingBots = true;
    try {
      await this.gameApi.addBots(count);
      this.snack.success(`Added ${count} bot${count > 1 ? 's' : ''} to the lobby.`);
    } catch (error) {
      this.snack.error('Failed to add bots.');
      console.error('Error adding bots:', error);
    } finally {
      this.addingBots = false;
    }
  }

  canAddBots(): boolean {
    const snapshot = this.gameApi.getCurrentSnapshot();
    if (!snapshot || !snapshot.game || !snapshot.viewer) {
      return false;
    }
    const playerCount = (snapshot.participants || []).filter((p) => p.role === 'player').length;
    return this.devModeEnabled && snapshot.viewer.isCreator && !snapshot.game.startedOn && playerCount < 4;
  }

  handleMeansCardCountChange() {
    this.settings.meansCardsPerPlayer = this.normalizePositiveNumber(this.settings.meansCardsPerPlayer, 4);
    if (this.settings.linkClueCountToMeans) {
      this.settings.clueCardsPerPlayer = this.settings.meansCardsPerPlayer;
    }
    this.markSettingsChanged();
  }

  handleClueCardCountChange() {
    this.settings.clueCardsPerPlayer = this.normalizePositiveNumber(
      this.settings.clueCardsPerPlayer,
      this.settings.meansCardsPerPlayer || 4,
    );
    this.markSettingsChanged();
  }

  handleLinkClueCountChange() {
    if (this.settings.linkClueCountToMeans) {
      this.settings.clueCardsPerPlayer = this.settings.meansCardsPerPlayer;
    }
    this.markSettingsChanged();
  }

  handleMeansCluesTextOnlyChange() {
    this.markSettingsChanged();
  }

  handleRandomMurdererCardSelectionChange() {
    this.markSettingsChanged();
  }

  handleShowAllRolesToScientistChange() {
    this.markSettingsChanged();
  }

  handleCrimePackChange() {
    const pack = this.selectedCrimePack;
    if (!pack) {
      return;
    }
    this.settings.crimePackLanguage = pack.defaultLanguage;
    this.settings.crimePackAssetSetId = pack.defaultAssetSetId || pack.assetSets[0]?.id || '';
    if (!pack.hasAnyImages) {
      this.settings.meansCluesTextOnly = true;
    }
    this.markSettingsChanged();
  }

  handleCrimePackLanguageChange() {
    this.markSettingsChanged();
  }

  handleCrimePackAssetSetChange() {
    this.markSettingsChanged();
  }

  handleHintPackChange() {
    const pack = this.selectedHintPack;
    if (pack) {
      this.settings.hintPackLanguage = pack.defaultLanguage;
    }
    this.markSettingsChanged();
  }

  handleHintPackLanguageChange() {
    this.markSettingsChanged();
  }

  handleRoleCountChange() {
    this.normalizeRoleSettings();
    this.markSettingsChanged();
  }

  adjustMeansCards(delta: number) {
    this.settings.meansCardsPerPlayer = (Number(this.settings.meansCardsPerPlayer) || 4) + delta;
    this.handleMeansCardCountChange();
  }

  adjustClueCards(delta: number) {
    if (this.settings.linkClueCountToMeans) {
      return;
    }
    this.settings.clueCardsPerPlayer = (Number(this.settings.clueCardsPerPlayer) || this.settings.meansCardsPerPlayer || 4) + delta;
    this.handleClueCardCountChange();
  }

  adjustAccomplices(delta: number) {
    this.settings.accompliceCount = (Number(this.settings.accompliceCount) || 0) + delta;
    this.handleRoleCountChange();
  }

  adjustWitnesses(delta: number) {
    this.settings.witnessCount = (Number(this.settings.witnessCount) || 0) + delta;
    this.handleRoleCountChange();
  }

  adjustWitnessesToFind(delta: number) {
    const fallback = this.settings.witnessCount > 0 ? 1 : 0;
    this.settings.witnessesToFind = (Number(this.settings.witnessesToFind) || fallback) + delta;
    this.handleRoleCountChange();
  }

  async flushSettingsAutoSave() {
    this.clearSettingsSaveTimer();
    if (!this.settingsDirty || this.settingsSaving) {
      return;
    }
    const saveVersion = this.settingsChangeVersion;
    this.settingsSaving = true;
    this.settingsSaveError = '';
    try {
      await this.gameApi.updateGameSettings(this.getNormalizedSettingsPayload());
      if (this.settingsChangeVersion === saveVersion) {
        this.settingsDirty = false;
        this.settingsSaved = true;
      }
    } catch (error) {
      console.warn('Failed to auto-save lobby settings', error);
      if (this.settingsChangeVersion === saveVersion) {
        this.settingsDirty = true;
        this.settingsSaved = false;
        this.settingsSaveError = 'Settings could not be saved. Retrying…';
      }
    } finally {
      this.settingsSaving = false;
      if (this.settingsDirty) {
        this.scheduleSettingsAutoSave(this.settingsSaveError ? this.settingsRetryDelayMs : this.settingsAutoSaveDelayMs);
      }
    }
  }

  async resetSettings(game: TgGame) {
    this.clearSettingsSaveTimer();
    this.syncSettingsFromGame(game);
    this.settingsDirty = false;
    this.settingsSaving = false;
    this.settingsSaveError = '';
    this.settingsSaved = false;
    this.settingsChangeVersion++;
  }

  getStartValidationMessage(game: TgGame, participants: TgParticipant[]): string {
    const playerCount = (participants || []).filter((participant) => participant.role === 'player').length;
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
    const crimePack = this.findCrimePack(game.crimePackId);
    if (crimePack && suspects * game.clueCardsPerPlayer > crimePack.clueCount) {
      return `Not enough clue cards for ${game.clueCardsPerPlayer} evidence cards per suspect.`;
    }
    if (crimePack && suspects * game.meansCardsPerPlayer > crimePack.meansCount) {
      return `Not enough means cards for ${game.meansCardsPerPlayer} means cards per suspect.`;
    }
    return '';
  }

  get selectedCrimePack(): TgCrimePackCatalogEntry {
    return this.findCrimePack(this.settings.crimePackId);
  }

  get selectedHintPack(): TgHintPackCatalogEntry {
    return this.findHintPack(this.settings.hintPackId);
  }

  get effectiveTextOnly(): boolean {
    return this.settings.meansCluesTextOnly || !this.selectedCrimePack?.hasAnyImages;
  }

  get showTextOnlyToggle(): boolean {
    return !!this.selectedCrimePack?.hasAnyImages;
  }

  get showCrimeAssetSelector(): boolean {
    return (this.selectedCrimePack?.assetSets?.length || 0) > 0 && this.selectedCrimePack?.hasAnyImages;
  }

  get hasSettingsAutoSaveWork(): boolean {
    return this.settingsDirty || this.settingsSaving || !!this.settingsSaveTimer;
  }

  private markSettingsChanged() {
    this.settingsDirty = true;
    this.settingsSaved = false;
    this.settingsSaveError = '';
    this.settingsChangeVersion++;
    this.scheduleSettingsAutoSave();
  }

  private scheduleSettingsAutoSave(delayMs = this.settingsAutoSaveDelayMs) {
    this.clearSettingsSaveTimer();
    this.settingsSaveTimer = setTimeout(() => {
      this.flushSettingsAutoSave().catch((error) => console.warn('Failed to flush lobby settings', error));
    }, delayMs);
  }

  private clearSettingsSaveTimer() {
    if (this.settingsSaveTimer) {
      clearTimeout(this.settingsSaveTimer);
      this.settingsSaveTimer = null;
    }
  }

  private getNormalizedSettingsPayload(): TgGameSettingsInput {
    this.settings.meansCardsPerPlayer = this.normalizePositiveNumber(this.settings.meansCardsPerPlayer, 4);
    if (this.settings.linkClueCountToMeans) {
      this.settings.clueCardsPerPlayer = this.settings.meansCardsPerPlayer;
    } else {
      this.settings.clueCardsPerPlayer = this.normalizePositiveNumber(
        this.settings.clueCardsPerPlayer,
        this.settings.meansCardsPerPlayer || 4,
      );
    }
    this.normalizeRoleSettings();
    return {
      ...this.settings,
      meansCluesTextOnly: this.effectiveTextOnly,
      crimePackAssetSetId: this.settings.crimePackAssetSetId || this.selectedCrimePack?.defaultAssetSetId || '',
    };
  }

  private normalizeRoleSettings() {
    this.settings.accompliceCount = this.normalizeBoundedNumber(this.settings.accompliceCount, 0, 10, 0);
    this.settings.witnessCount = this.normalizeBoundedNumber(this.settings.witnessCount, 0, 10, 0);
    if (this.settings.witnessCount === 0) {
      this.settings.witnessesToFind = 0;
    } else {
      this.settings.witnessesToFind = this.normalizeBoundedNumber(this.settings.witnessesToFind, 1, this.settings.witnessCount, 1);
    }
  }

  private defaultSettings(): TgGameSettingsInput {
    return {
      meansCardsPerPlayer: 4,
      clueCardsPerPlayer: 4,
      linkClueCountToMeans: true,
      meansCluesTextOnly: false,
      randomMurdererCardSelection: false,
      showAllRolesToScientist: true,
      crimePackId: 'treachery',
      crimePackLanguage: 'fa',
      crimePackAssetSetId: 'gouache-treachery',
      hintPackId: 'treachery-hints',
      hintPackLanguage: 'fa',
      accompliceCount: 0,
      witnessCount: 0,
      witnessesToFind: 0,
    };
  }

  private syncSettingsFromGame(game: TgGame) {
    this.settings = {
      meansCardsPerPlayer: game.meansCardsPerPlayer,
      clueCardsPerPlayer: game.clueCardsPerPlayer,
      linkClueCountToMeans: game.linkClueCountToMeans,
      meansCluesTextOnly: game.meansCluesTextOnly,
      randomMurdererCardSelection: game.randomMurdererCardSelection,
      showAllRolesToScientist: game.showAllRolesToScientist,
      crimePackId: game.crimePackId,
      crimePackLanguage: game.crimePackLanguage,
      crimePackAssetSetId: game.crimePackAssetSetId,
      hintPackId: game.hintPackId,
      hintPackLanguage: game.hintPackLanguage,
      accompliceCount: game.accompliceCount,
      witnessCount: game.witnessCount,
      witnessesToFind: game.witnessesToFind,
    };
  }

  private findCrimePack(id: string) {
    return this.catalog?.crimePacks?.find((pack) => pack.id === id) || null;
  }

  private findHintPack(id: string) {
    return this.catalog?.hintPacks?.find((pack) => pack.id === id) || null;
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
