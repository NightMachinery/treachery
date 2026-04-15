import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';
import { map, shareReplay } from 'rxjs/operators';
import { AuthService } from './../auth/auth.service';
import { randomReadableId } from './../util';
import { GameApiService } from './../game/game-api.service';
import { TgForensicPrivateData, TgGame } from '../models/models';
import { SnackBarService } from '../snack-bar/snack-bar.service';
import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class ForensicApiService {
  forensicPrivateData$: Observable<TgForensicPrivateData>;

  constructor(
    private http: HttpClient,
    private authService: AuthService,
    private router: Router,
    public gameApi: GameApiService,
    private snack: SnackBarService
  ) {
    this.forensicPrivateData$ = this.gameApi.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.forensicPrivateData : null)),
      shareReplay(1)
    );
  }

  updateGameId(gameId: string) {
    this.gameApi.setGameId(gameId);
  }

  getGame(): Observable<TgGame> {
    return this.gameApi.game$;
  }

  getPrivateData(): Observable<TgForensicPrivateData> {
    return this.forensicPrivateData$;
  }

  async createGame() {
    const requestedGameId = randomReadableId();
    const response = await this.http
      .post<{ success: boolean; gameId: string }>(`/api/games`, { gameId: requestedGameId })
      .toPromise();

    if (response.success) {
      this.router.navigateByUrl(`/forensic/${response.gameId || requestedGameId}`);
    }
  }

  async startGame() {
    const players = this.gameApi.getCurrentSnapshot() ? this.gameApi.getCurrentSnapshot().players : [];
    const gameId = this.gameApi.gameId$.value;
    if (!players || players.length < 3) {
      this.snack.error('Need at least 3 players to start!');
      return;
    }
    const response = await this.http.post<{ success: boolean }>(`/api/games/${gameId}/start`, {}).toPromise();
    if (response.success) {
      await this.gameApi.refreshSnapshot();
    }
  }

  async endGame() {
    const gameId = this.gameApi.gameId$.value;
    const response = await this.http.post<{ success: boolean }>(`/api/games/${gameId}/end`, {}).toPromise();
    if (response.success) {
      await this.gameApi.refreshSnapshot();
    }
  }
}
