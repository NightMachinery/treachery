import { Component, Input } from '@angular/core';
import { GameApiService } from '../../../shared/api/game/game-api.service';
import { TgPlayer } from '../../../shared/api/models/models';

@Component({
  selector: 'app-murderer-select-dialog',
  templateUrl: './murderer-select-dialog.component.html',
  styleUrls: ['./murderer-select-dialog.component.scss']
})
export class MurdererSelectDialogComponent {
  @Input() player: TgPlayer;
  selectedClue: string;
  selectedMeans: string;
  minimized = false;

  constructor(public gameApi: GameApiService) {}

  selectCards() {
    this.gameApi.selectMurdererCards(this.selectedClue, this.selectedMeans);
  }
}
