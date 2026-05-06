import { Component, Input, OnInit } from '@angular/core';
import { GameApiService } from '../../../shared/api/game/game-api.service';
import { TgParticipant } from '../../../shared/api/models/models';

@Component({
  selector: 'app-joined-players-list',
  standalone: false,
  templateUrl: './joined-players-list.component.html',
  styleUrls: ['./joined-players-list.component.scss'],
})
export class JoinedPlayersListComponent implements OnInit {
  @Input() showModControls = false;

  constructor(public gameApi: GameApiService) {}

  ngOnInit() {}

  async toggleRole(participant: TgParticipant) {
    const nextRole = participant.role === 'player' ? 'observer' : 'player';
    await this.gameApi.setParticipantRole(participant.uid, nextRole);
  }

  async toggleScientist(participant: TgParticipant) {
    await this.gameApi.toggleScientist(participant.uid);
  }
}
