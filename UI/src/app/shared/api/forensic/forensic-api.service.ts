import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { firstValueFrom, Observable } from 'rxjs';
import { map, shareReplay } from 'rxjs/operators';
import { AuthService } from './../auth/auth.service';
import { randomReadableId } from './../util';
import { GameApiService } from './../game/game-api.service';
import { TgForensicPrivateData, TgGame } from '../models/models';
import { SnackBarService } from '../snack-bar/snack-bar.service';

@Injectable({
  providedIn: 'root'
})
export class ForensicApiService {
  forensicPrivateData$: Observable<TgForensicPrivateData>;

  constructor(
    private http: HttpClient,
    private authService: AuthService,
    public gameApi: GameApiService,
    private snack: SnackBarService
  ) {
    this.forensicPrivateData$ = this.gameApi.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.forensicPrivateData : null)),
      shareReplay(1)
    );
  }

  updateGameId(gameId: string, roomAuth?: string) {
    this.gameApi.setGameContext(gameId, roomAuth);
  }

  getGame(): Observable<TgGame> {
    return this.gameApi.game$;
  }

  getPrivateData(): Observable<TgForensicPrivateData> {
    return this.forensicPrivateData$;
  }

  async createGame() {
    if (!(await this.authService.ensureDisplayName())) {
      return;
    }
    const requestedGameId = randomReadableId();
    const response = await firstValueFrom(this.http.post<{ success: boolean; gameId: string }>(`/api/games`, { gameId: requestedGameId }));

    if (response.success) {
      this.gameApi.setGameContext(response.gameId || requestedGameId, null);
      await this.gameApi.navigateTo('join', response.gameId || requestedGameId);
    }
  }

  async startGame() {
    const participants = this.gameApi.getCurrentSnapshot() ? this.gameApi.getCurrentSnapshot().participants : [];
    const players = participants ? participants.filter(participant => participant.role === 'player') : [];
    const gameId = this.gameApi.gameId$.value;
    if (!players || players.length < 4) {
      this.snack.error('Need at least 4 players to start!');
      return;
    }
    const response = await firstValueFrom(this.http.post<{ success: boolean }>(`/api/games/${gameId}/start`, {}, this.getRoomRequestOptions()));
    if (response.success) {
      await this.gameApi.refreshSnapshot();
    }
  }

  async endGame() {
    const gameId = this.gameApi.gameId$.value;
    const response = await firstValueFrom(this.http.post<{ success: boolean }>(`/api/games/${gameId}/end`, {}, this.getRoomRequestOptions()));
    if (response.success) {
      await this.gameApi.refreshSnapshot();
    }
  }

  async restartGame() {
    const gameId = this.gameApi.gameId$.value;
    const response = await firstValueFrom(this.http.post<{ success: boolean }>(`/api/games/${gameId}/restart`, {}, this.getRoomRequestOptions()));
    if (response.success) {
      await this.gameApi.refreshSnapshot();
    }
  }

  private getRoomRequestOptions() {
    let headers = new HttpHeaders();
    if (this.gameApi.roomAuth$.value) {
      headers = headers.set('X-Treachery-Room-Auth', this.gameApi.roomAuth$.value);
    }
    return { headers };
  }
}
