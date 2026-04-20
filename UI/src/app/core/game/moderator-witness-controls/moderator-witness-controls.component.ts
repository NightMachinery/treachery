import { Component } from '@angular/core';
import { GameApiService } from '../../../shared/api/game/game-api.service';

@Component({
  selector: 'app-moderator-witness-controls',
  templateUrl: './moderator-witness-controls.component.html',
  styleUrls: ['./moderator-witness-controls.component.scss']
})
export class ModeratorWitnessControlsComponent {
  constructor(public gameApi: GameApiService) {}

  async showPrompt(targetUid: string) {
    await this.gameApi.showWitnessSelectionPrompt(targetUid);
  }
}
