import { Component, OnInit } from '@angular/core';
import { combineLatest, Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { AuthService } from '../../api/auth/auth.service';
import { GameApiService } from '../../api/game/game-api.service';

@Component({
  selector: 'tg-navbar',
  templateUrl: './navbar.component.html',
  styleUrls: ['./navbar.component.scss']
})
export class NavbarComponent implements OnInit {
  identity$: Observable<{ uid: string; name: string }>;

  constructor(public authService: AuthService, public gameApi: GameApiService) {
    this.identity$ = combineLatest([this.authService.user$, this.authService.displayName$, this.gameApi.viewer$, this.gameApi.game$]).pipe(
      map(([user, displayName, viewer, game]) => {
        if (game && viewer && viewer.uid) {
          return { uid: viewer.uid, name: viewer.name || displayName || 'Anonymous' };
        }
        return { uid: user ? user.uid : '', name: displayName || 'Set name' };
      })
    );
  }

  ngOnInit(): void {}

  async rename() {
    const changed = await this.authService.promptForDisplayName(this.gameApi.gameId$.value, this.gameApi.roomAuth$.value);
    if (changed && this.gameApi.gameId$.value) {
      await this.gameApi.refreshSnapshot();
    }
  }
}
