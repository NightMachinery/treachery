import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, firstValueFrom, Observable } from 'rxjs';
import { distinctUntilChanged } from 'rxjs/operators';
import { TgAuthUser } from '../models/models';

const SESSION_STORAGE_KEY = 'treachery.session.token';
const DISPLAY_NAME_STORAGE_KEY = 'treachery.session.displayName';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  public user$: Observable<TgAuthUser>;
  public displayName$: Observable<string>;
  public user: TgAuthUser;
  private readonly userSubject = new BehaviorSubject<TgAuthUser>(null);
  private readonly displayNameSubject = new BehaviorSubject<string>(localStorage.getItem(DISPLAY_NAME_STORAGE_KEY) || '');

  constructor(private http: HttpClient) {
    this.user$ = this.userSubject.asObservable();
    this.displayName$ = this.displayNameSubject.asObservable().pipe(distinctUntilChanged());
    this.user$.subscribe(value => {
      this.user = value;
      if (!value) {
        return;
      }
      this.persistSession(value);
      this.displayNameSubject.next(value.displayName || '');
    });
  }

  getStoredToken() {
    return localStorage.getItem(SESSION_STORAGE_KEY) || '';
  }

  getStoredDisplayName() {
    return this.displayNameSubject.value || '';
  }

  anonymousLogin() {
    this.http.post<TgAuthUser>('/api/session', {}).subscribe({
      next: user => {
        this.userSubject.next(user);
      },
      error: error => {
        console.error('Failed to establish anonymous session', error);
      }
    });
  }

  async updateProfile(displayName: string, gameId?: string, roomAuth?: string): Promise<TgAuthUser> {
    let headers = new HttpHeaders();
    if (roomAuth) {
      headers = headers.set('X-Treachery-Room-Auth', roomAuth);
    }
    const user = await firstValueFrom(this.http.put<TgAuthUser>('/api/me', { displayName, gameId }, { headers }));
    this.userSubject.next({ ...(this.user || ({} as TgAuthUser)), ...user, displayName: user.displayName || displayName });
    return this.user;
  }

  async ensureDisplayName(gameId?: string, roomAuth?: string): Promise<boolean> {
    if (this.getStoredDisplayName()) {
      return true;
    }
    return this.promptForDisplayName(gameId, roomAuth);
  }

  async promptForDisplayName(gameId?: string, roomAuth?: string): Promise<boolean> {
    const current = this.getStoredDisplayName();
    const result = window.prompt('Choose a display name for Treachery', current || '');
    const trimmed = (result || '').trim();
    if (!trimmed) {
      return false;
    }
    await this.updateProfile(trimmed, gameId, roomAuth);
    return true;
  }

  private persistSession(user: TgAuthUser) {
    if (user.token) {
      localStorage.setItem(SESSION_STORAGE_KEY, user.token);
    }
    if (user.displayName) {
      localStorage.setItem(DISPLAY_NAME_STORAGE_KEY, user.displayName);
    } else {
      localStorage.removeItem(DISPLAY_NAME_STORAGE_KEY);
    }
  }
}
