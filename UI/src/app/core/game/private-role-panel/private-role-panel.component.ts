import { Component } from '@angular/core';
import { combineLatest } from 'rxjs';
import { map } from 'rxjs/operators';
import { GameApiService } from '../../../shared/api/game/game-api.service';
import { findClueCard, findMeansCard } from '../../../shared/utils/find';

@Component({
  selector: 'app-private-role-panel',
  standalone: false,
  templateUrl: './private-role-panel.component.html',
  styleUrls: ['./private-role-panel.component.scss']
})
export class PrivateRolePanelComponent {
  vm$ = combineLatest([this.gameApi.playerPrivateData$, this.gameApi.players$]).pipe(
    map(([privateData, players]) => {
      const murderer = players.find(player => player.uid === (privateData.knownMurdererTeam || []).find(member => member.role === 'murderer')?.uid);
      return {
        privateData,
        murderer,
        murdererClueCard: murderer ? findClueCard(murderer, privateData.knownMurdererClueCardName) : null,
        murdererMeansCard: murderer ? findMeansCard(murderer, privateData.knownMurdererMeansCardName) : null
      };
    })
  );

  constructor(public gameApi: GameApiService) {}

  getRoleTitle(role: string) {
    switch (role) {
      case 'murderer':
        return 'You are the murderer';
      case 'accomplice':
        return 'You are an accomplice';
      case 'witness':
        return 'You are a witness';
      default:
        return 'You are an investigator';
    }
  }
}
