import { Component } from '@angular/core';
import { GameApiService } from '../../../shared/api/game/game-api.service';

@Component({
  selector: 'app-moderator-witness-controls',
  standalone: false,
  templateUrl: './moderator-witness-controls.component.html',
  styleUrls: ['./moderator-witness-controls.component.scss']
})
export class ModeratorWitnessControlsComponent {
  roomModsBusy = false;

  constructor(public gameApi: GameApiService) {}

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
}
