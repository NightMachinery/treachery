import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { BehaviorSubject, Observable, of, timer } from 'rxjs';
import { catchError, map, shareReplay, switchMap } from 'rxjs/operators';
import { AuthService } from './../auth/auth.service';
import { TgGame, TgGameSnapshot, TgGuess, TgPartialGuess, TgPlayer, TgPlayerPrivateData, TgForensicCard } from '../models/models';
import { SnackBarService } from '../snack-bar/snack-bar.service';

@Injectable({
  providedIn: 'root'
})
export class GameApiService {
  public snapshot$: Observable<TgGameSnapshot>;
  public game$: Observable<TgGame>;
  public players$: Observable<TgPlayer[]>;
  public me$: Observable<TgPlayer>;
  public gameId$: BehaviorSubject<string>;
  public guesses$: Observable<TgGuess[]>;
  public playerPrivateData$: Observable<TgPlayerPrivateData>;
  public playersDict$: Observable<Map<string, TgPlayer>>;
  public joinLink$: Observable<string>;
  public activeGames$: Observable<TgGame[]>;

  private readonly snapshotSubject = new BehaviorSubject<TgGameSnapshot>(null);
  private eventSource: EventSource;

  constructor(
    private http: HttpClient,
    private auth: AuthService,
    private router: Router,
    private snack: SnackBarService
  ) {
    this.gameId$ = new BehaviorSubject<string>(null);
    this.snapshot$ = this.snapshotSubject.asObservable().pipe(shareReplay(1));

    this.game$ = this.snapshot$.pipe(map(snapshot => snapshot ? snapshot.game : null), shareReplay(1));
    this.players$ = this.snapshot$.pipe(map(snapshot => snapshot ? snapshot.players : null), shareReplay(1));
    this.me$ = this.players$.pipe(map(players => {
      if (players && this.auth.user) {
        return players.find(player => player.uid === this.auth.user.uid);
      }
      return null;
    }), shareReplay(1));
    this.guesses$ = this.snapshot$.pipe(map(snapshot => snapshot ? snapshot.guesses : []), shareReplay(1));
    this.playerPrivateData$ = this.snapshot$.pipe(map(snapshot => snapshot ? snapshot.playerPrivateData || {} as TgPlayerPrivateData : {} as TgPlayerPrivateData), shareReplay(1));
    this.playersDict$ = this.players$.pipe(map(players => {
      if (!players) {
        return null;
      }
      const result = new Map<string, TgPlayer>();
      players.forEach(player => result.set(player.uid, player));
      return result;
    }), shareReplay(1));
    this.joinLink$ = this.gameId$.pipe(map(value => value ? `${window.location.origin}/join/${value}` : `${window.location.origin}`), shareReplay(1));
    this.activeGames$ = timer(0, 5000).pipe(
      switchMap(() => this.http.get<TgGame[]>('/api/games').pipe(catchError(() => of([])))),
      shareReplay(1)
    );

    this.auth.user$.subscribe(user => {
      if (user && this.gameId$.value) {
        this.connectEvents();
        this.refreshSnapshot();
      }
    });
  }

  setGameId(gameId: string) {
    const normalizedGameId = gameId ? gameId.toUpperCase() : null;
    if (normalizedGameId === this.gameId$.value) {
      return;
    }
    this.closeEvents();
    this.snapshotSubject.next(null);
    this.gameId$.next(normalizedGameId);
    if (normalizedGameId && this.auth.user) {
      this.connectEvents();
      this.refreshSnapshot();
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

  getCurrentGame(): Observable<TgGame> {
    return this.game$;
  }

  getCurrentSnapshot(): TgGameSnapshot {
    return this.snapshotSubject.value;
  }

  getCurrentGamePlayers(): Observable<TgPlayer[]> {
    return this.players$;
  }

  async joinGame(gameId: string, playerName: string) {
    this.setGameId(gameId);
    const response = await this.http.post<{ success: boolean }>(`/api/games/${gameId.toUpperCase()}/join`, { playerName }).toPromise();
    if (response.success) {
      await this.refreshSnapshot();
      this.router.navigateByUrl(`/play/${gameId.toUpperCase()}`);
    }
  }

  async selectMurdererCards(clueCardName: string, meansCardName: string) {
    const gameId = this.gameId$.value;
    const response = await this.http.post<{ success: boolean }>(`/api/games/${gameId}/murderer-selection`, { clueCardName, meansCardName }).toPromise();
    if (response.success) {
      await this.refreshSnapshot();
    }
  }

  async makeGuess(guess: TgPartialGuess) {
    const gameId = this.gameId$.value;
    const response = await this.http.post<{ success: boolean }>(`/api/games/${gameId}/guess`, guess).toPromise();
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
    await this.http.post(`/api/games/${gameId}/forensic/cause`, { card }).toPromise();
    await this.refreshSnapshot();
  }

  async selectForensicLocationCard(card: TgForensicCard) {
    if (!card || !card.selectedChoice) {
      this.snack.error('Please select an option from the cards!');
      return;
    }
    const gameId = this.gameId$.value;
    await this.http.post(`/api/games/${gameId}/forensic/location`, { card }).toPromise();
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
    await this.http.post(`/api/games/${gameId}/forensic/other`, { card, replaceCardName }).toPromise();
    await this.refreshSnapshot();
  }

  findPlayer(players: TgPlayer[], uid: string): TgPlayer {
    return players.find(player => player.uid === uid);
  }

  async gameExists(gameId: string): Promise<boolean> {
    try {
      await this.http.get<TgGameSnapshot>(`/api/games/${gameId.toUpperCase()}/snapshot`).toPromise();
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
    const snapshot = await this.http.get<TgGameSnapshot>(`/api/games/${gameId}/snapshot`).toPromise();
    this.snapshotSubject.next(snapshot);
    return snapshot;
  }

  private connectEvents() {
    if (!this.gameId$.value || !this.auth.user || this.eventSource) {
      return;
    }
    this.eventSource = new EventSource(`/api/games/${this.gameId$.value}/events`);
    this.eventSource.addEventListener('update', () => {
      this.refreshSnapshot();
    });
    this.eventSource.onmessage = () => {
      this.refreshSnapshot();
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
