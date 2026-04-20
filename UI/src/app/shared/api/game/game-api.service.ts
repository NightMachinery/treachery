import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { BehaviorSubject, Observable, of, timer } from 'rxjs';
import { catchError, map, shareReplay, switchMap } from 'rxjs/operators';
import { AuthService } from './../auth/auth.service';
import {
  TgGame,
  TgGameSnapshot,
  TgGameSettingsInput,
  TgGuess,
  TgPartialGuess,
  TgPlayer,
  TgPlayerPrivateData,
  TgForensicCard,
  TgModeratorPrivateData,
  TgParticipant,
  TgRoomTimer,
  TgRoleRevealEntry,
  TgViewer
} from '../models/models';
import { SnackBarService } from '../snack-bar/snack-bar.service';

@Injectable({
  providedIn: 'root'
})
export class GameApiService {
  public snapshot$: Observable<TgGameSnapshot>;
  public game$: Observable<TgGame>;
  public participants$: Observable<TgParticipant[]>;
  public players$: Observable<TgPlayer[]>;
  public me$: Observable<TgPlayer>;
  public viewer$: Observable<TgViewer>;
  public gameId$: BehaviorSubject<string>;
  public roomAuth$: BehaviorSubject<string>;
  public guesses$: Observable<TgGuess[]>;
  public playerPrivateData$: Observable<TgPlayerPrivateData>;
  public moderatorPrivateData$: Observable<TgModeratorPrivateData>;
  public roomTimer$: Observable<TgRoomTimer>;
  public roleReveal$: Observable<TgRoleRevealEntry[]>;
  public participantsDict$: Observable<Map<string, TgParticipant>>;
  public playersDict$: Observable<Map<string, TgPlayer>>;
  public joinLink$: Observable<string>;
  public migrateLink$: Observable<string>;
  public activeGames$: Observable<TgGame[]>;

  private readonly snapshotSubject = new BehaviorSubject<TgGameSnapshot>(null);
  private eventSource: EventSource;

  constructor(private http: HttpClient, private auth: AuthService, private router: Router, private snack: SnackBarService) {
    this.gameId$ = new BehaviorSubject<string>(null);
    this.roomAuth$ = new BehaviorSubject<string>(null);
    this.snapshot$ = this.snapshotSubject.asObservable().pipe(shareReplay(1));

    this.game$ = this.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.game : null)),
      shareReplay(1)
    );
    this.participants$ = this.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.participants : [])),
      shareReplay(1)
    );
    this.players$ = this.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.players : [])),
      shareReplay(1)
    );
    this.viewer$ = this.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.viewer : null)),
      shareReplay(1)
    );
    this.me$ = this.snapshot$.pipe(
      map(snapshot => {
        if (!snapshot || !snapshot.viewer) {
          return null;
        }
        return snapshot.players.find(player => player.uid === snapshot.viewer.uid) || null;
      }),
      shareReplay(1)
    );
    this.guesses$ = this.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.guesses : [])),
      shareReplay(1)
    );
    this.playerPrivateData$ = this.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.playerPrivateData || ({} as TgPlayerPrivateData) : ({} as TgPlayerPrivateData))),
      shareReplay(1)
    );
    this.moderatorPrivateData$ = this.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.moderatorPrivateData || ({ witnessPromptTargets: [] } as TgModeratorPrivateData) : ({ witnessPromptTargets: [] } as TgModeratorPrivateData))),
      shareReplay(1)
    );
    this.roleReveal$ = this.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.roleReveal || [] : [])),
      shareReplay(1)
    );
    this.roomTimer$ = this.game$.pipe(
      map(game => (game ? game.roomTimer || null : null)),
      shareReplay(1)
    );
    this.participantsDict$ = this.participants$.pipe(
      map(participants => {
        const result = new Map<string, TgParticipant>();
        participants.forEach(participant => result.set(participant.uid, participant));
        return result;
      }),
      shareReplay(1)
    );
    this.playersDict$ = this.players$.pipe(
      map(players => {
        const result = new Map<string, TgPlayer>();
        players.forEach(player => result.set(player.uid, player));
        return result;
      }),
      shareReplay(1)
    );
    this.joinLink$ = this.gameId$.pipe(
      map(value => (value ? `${window.location.origin}/join/${value}` : `${window.location.origin}`)),
      shareReplay(1)
    );
    this.migrateLink$ = this.gameId$.pipe(
      map(value => {
        const roomAuth = this.roomAuth$.value;
        if (!value || !roomAuth) {
          return '';
        }
        return `${window.location.origin}/join/${value}?roomAuth=${encodeURIComponent(roomAuth)}`;
      }),
      shareReplay(1)
    );
    this.activeGames$ = timer(0, 5000).pipe(
      switchMap(() => this.http.get<TgGame[]>('/api/games').pipe(catchError(() => of([])))),
      shareReplay(1)
    );

    this.auth.user$.subscribe(user => {
      if (user && this.gameId$.value) {
        this.connectEvents();
        this.refreshSnapshotInBackground();
      }
    });
  }

  setGameContext(gameId: string, roomAuth?: string) {
    const normalizedGameId = gameId ? gameId.toUpperCase() : null;
    const normalizedRoomAuth = roomAuth ? roomAuth.trim() : null;
    if (normalizedGameId === this.gameId$.value && normalizedRoomAuth === this.roomAuth$.value) {
      return;
    }
    this.closeEvents();
    this.snapshotSubject.next(null);
    this.gameId$.next(normalizedGameId);
    this.roomAuth$.next(normalizedRoomAuth);
    if (normalizedGameId && this.auth.user) {
      this.connectEvents();
      this.refreshSnapshotInBackground();
    }
  }

  getGameGuesses(): Observable<TgGuess[]> {
    return this.guesses$;
  }

  getPrivateData(): Observable<TgPlayerPrivateData> {
    return this.playerPrivateData$;
  }

  getCurrentGamePlayer(): Observable<TgPlayer> {
    return this.me$;
  }

  getCurrentViewer(): Observable<TgViewer> {
    return this.viewer$;
  }

  getCurrentGame(): Observable<TgGame> {
    return this.game$;
  }

  getCurrentSnapshot(): TgGameSnapshot {
    return this.snapshotSubject.value;
  }

  getCurrentGamePlayers(): Observable<TgPlayer[]> {
    return this.players$;
  }

  getCurrentParticipants(): Observable<TgParticipant[]> {
    return this.participants$;
  }

  async joinGame(gameId: string, role: 'player' | 'observer' = 'player') {
    this.setGameContext(gameId, this.roomAuth$.value);
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId.toUpperCase()}/join`, { role }, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async setParticipantRole(uid: string, role: 'player' | 'observer') {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId}/participants/${uid}/role`, { role }, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async toggleScientist(uid: string) {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId}/scientist/toggle`, { uid }, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async updateGameSettings(settings: TgGameSettingsInput) {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId}/settings`, settings, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async createMigrateLink() {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean; token: string }>(`/api/games/${gameId}/migrate-device`, {}, this.getRoomRequestOptions())
      .toPromise();
    if (!response.success || !response.token) {
      return '';
    }
    this.roomAuth$.next(response.token);
    return `${window.location.origin}/join/${gameId}?roomAuth=${encodeURIComponent(response.token)}`;
  }

  async selectMurdererCards(clueCardName: string, meansCardName: string) {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId}/murderer-selection`, { clueCardName, meansCardName }, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async startRoomTimer(seconds?: number) {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId}/room-timer/start`, { seconds }, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async pauseRoomTimer() {
    const gameId = this.gameId$.value;
    const response = await this.http.post<{ success: boolean }>(`/api/games/${gameId}/room-timer/pause`, {}, this.getRoomRequestOptions()).toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async resumeRoomTimer() {
    const gameId = this.gameId$.value;
    const response = await this.http.post<{ success: boolean }>(`/api/games/${gameId}/room-timer/resume`, {}, this.getRoomRequestOptions()).toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async resetRoomTimer() {
    const gameId = this.gameId$.value;
    const response = await this.http.post<{ success: boolean }>(`/api/games/${gameId}/room-timer/reset`, {}, this.getRoomRequestOptions()).toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async clearRoomTimer() {
    const gameId = this.gameId$.value;
    const response = await this.http.post<{ success: boolean }>(`/api/games/${gameId}/room-timer/clear`, {}, this.getRoomRequestOptions()).toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async showWitnessSelectionPrompt(targetUid: string) {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId}/witness-selection/show`, { targetUid }, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async submitWitnessSelection(selectedUids: string[]) {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId}/witness-selection`, { selectedUids }, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async dismissWitnessSelectionPrompt() {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId}/witness-selection/dismiss`, {}, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async makeGuess(guess: TgPartialGuess) {
    const gameId = this.gameId$.value;
    const response = await this.http
      .post<{ success: boolean }>(`/api/games/${gameId}/guess`, guess, this.getRoomRequestOptions())
      .toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async selectForensicCauseCard(card: TgForensicCard) {
    if (!card || !card.selectedChoice) {
      this.snack.error('Please select an option from the cards!');
      return;
    }
    const gameId = this.gameId$.value;
    await this.http.post(`/api/games/${gameId}/forensic/cause`, { card }, this.getRoomRequestOptions()).toPromise();
    await this.refreshSnapshot();
  }

  async selectForensicLocationCard(card: TgForensicCard) {
    if (!card || !card.selectedChoice) {
      this.snack.error('Please select an option from the cards!');
      return;
    }
    const gameId = this.gameId$.value;
    await this.http.post(`/api/games/${gameId}/forensic/location`, { card }, this.getRoomRequestOptions()).toPromise();
    await this.refreshSnapshot();
  }

  countSelectedOtherCards(game: TgGame): number {
    return game.otherCards.filter(card => card.selectedChoice).length;
  }

  async selectNextForensicOtherCard(card: TgForensicCard, replaceCardName?: string) {
    if (!card || !card.selectedChoice) {
      this.snack.error('Please select an option from the card!');
      return;
    }
    const gameId = this.gameId$.value;
    await this.http.post(`/api/games/${gameId}/forensic/other`, { card, replaceCardName }, this.getRoomRequestOptions()).toPromise();
    await this.refreshSnapshot();
  }

  findPlayer(players: TgPlayer[], uid: string): TgPlayer {
    return players.find(player => player.uid === uid);
  }

  findParticipant(participants: TgParticipant[], uid: string): TgParticipant {
    return participants.find(participant => participant.uid === uid);
  }

  async gameExists(gameId: string): Promise<boolean> {
    try {
      await this.http.get<TgGameSnapshot>(`/api/games/${gameId.toUpperCase()}/snapshot`, this.getRoomRequestOptions()).toPromise();
      return true;
    } catch (error) {
      return false;
    }
  }

  async refreshSnapshot(): Promise<TgGameSnapshot> {
    const gameId = this.gameId$.value;
    if (!gameId || !this.auth.user) {
      return null;
    }
    const snapshot = await this.http.get<TgGameSnapshot>(`/api/games/${gameId}/snapshot`, this.getRoomRequestOptions()).toPromise();
    this.snapshotSubject.next(snapshot);
    return snapshot;
  }

  async navigateTo(route: 'join' | 'play' | 'forensic' | 'observe', gameId?: string) {
    const targetGameId = (gameId || this.gameId$.value || '').toUpperCase();
    await this.router.navigate([`/${route}`, targetGameId], { queryParams: this.getRoomQueryParams() });
  }

  getRoomQueryParams() {
    return this.roomAuth$.value ? { roomAuth: this.roomAuth$.value } : {};
  }

  private refreshSnapshotInBackground() {
    this.refreshSnapshot().catch(error => {
      console.warn('Failed to refresh game snapshot in background.', error);
    });
  }

  private getRoomRequestOptions() {
    let headers = new HttpHeaders();
    if (this.roomAuth$.value) {
      headers = headers.set('X-Treachery-Room-Auth', this.roomAuth$.value);
    }
    return { headers };
  }

  private connectEvents() {
    if (!this.gameId$.value || !this.auth.user || this.eventSource) {
      return;
    }
    const params = [`sessionToken=${encodeURIComponent(this.auth.getStoredToken())}`];
    if (this.roomAuth$.value) {
      params.push(`roomAuth=${encodeURIComponent(this.roomAuth$.value)}`);
    }
    this.eventSource = new EventSource(`/api/games/${this.gameId$.value}/events?${params.join('&')}`);
    this.eventSource.addEventListener('update', () => {
      this.refreshSnapshotInBackground();
    });
    this.eventSource.onmessage = () => {
      this.refreshSnapshotInBackground();
    };
    this.eventSource.onerror = () => {
      console.warn('Game event stream disconnected; waiting for automatic retry.');
    };
  }

  private closeEvents() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
  }
}
