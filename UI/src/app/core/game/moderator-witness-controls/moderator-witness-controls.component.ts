import { Component } from '@angular/core';
import { CardApiService } from '../../../shared/api/card/card-api.service';
import { GameApiService } from '../../../shared/api/game/game-api.service';
import { TgCrimePackCatalogEntry, TgGame, TgHintPackCatalogEntry, TgWordpackCatalog } from '../../../shared/api/models/models';

@Component({
  selector: 'app-moderator-witness-controls',
  standalone: false,
  templateUrl: './moderator-witness-controls.component.html',
  styleUrls: ['./moderator-witness-controls.component.scss']
})
export class ModeratorWitnessControlsComponent {
  roomModsBusy = false;

  constructor(
    public gameApi: GameApiService,
    public cardApi: CardApiService
  ) {}

  async showPrompt(targetUid: string) {
    await this.gameApi.showWitnessSelectionPrompt(targetUid);
  }

  async setMeansCluesTextOnly(enabled: boolean, currentValue: boolean) {
    if (this.roomModsBusy || enabled === currentValue) {
      return;
    }
    this.roomModsBusy = true;
    try {
      await this.gameApi.updateRoomMods({ meansCluesTextOnly: enabled });
    } finally {
      this.roomModsBusy = false;
    }
  }

  async setCrimePackLanguage(language: string, currentValue: string) {
    if (this.roomModsBusy || !language || language === currentValue) {
      return;
    }
    this.roomModsBusy = true;
    try {
      await this.gameApi.updateRoomMods({ crimePackLanguage: language });
    } finally {
      this.roomModsBusy = false;
    }
  }

  async setCrimePackAssetSetId(assetSetId: string, currentValue: string) {
    if (this.roomModsBusy || !assetSetId || assetSetId === currentValue) {
      return;
    }
    this.roomModsBusy = true;
    try {
      await this.gameApi.updateRoomMods({ crimePackAssetSetId: assetSetId });
    } finally {
      this.roomModsBusy = false;
    }
  }

  async setHintPackLanguage(language: string, currentValue: string) {
    if (this.roomModsBusy || !language || language === currentValue) {
      return;
    }
    this.roomModsBusy = true;
    try {
      await this.gameApi.updateRoomMods({ hintPackLanguage: language });
    } finally {
      this.roomModsBusy = false;
    }
  }

  getCrimePack(catalog: TgWordpackCatalog, game: TgGame): TgCrimePackCatalogEntry | null {
    return catalog?.crimePacks?.find(pack => pack.id === game?.crimePackId) || null;
  }

  getHintPack(catalog: TgWordpackCatalog, game: TgGame): TgHintPackCatalogEntry | null {
    return catalog?.hintPacks?.find(pack => pack.id === game?.hintPackId) || null;
  }

  showCrimeAssetSelector(crimePack: TgCrimePackCatalogEntry | null): boolean {
    return !!crimePack?.hasAnyImages && (crimePack.assetSets?.length || 0) > 1;
  }
}
