import { Component } from '@angular/core';
import { GameApiService } from '../../../shared/api/game/game-api.service';

@Component({
  selector: 'app-role-reveal',
  standalone: false,
  templateUrl: './role-reveal.component.html',
  styleUrls: ['./role-reveal.component.scss'],
})
export class RoleRevealComponent {
  constructor(public gameApi: GameApiService) {}
}
