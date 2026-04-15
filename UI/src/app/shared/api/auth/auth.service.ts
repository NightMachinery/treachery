import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { TgAuthUser } from '../models/models';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  public user$: Observable<TgAuthUser>;
  public user: TgAuthUser;
  private readonly userSubject = new BehaviorSubject<TgAuthUser>(null);

  constructor(private http: HttpClient) {
    this.user$ = this.userSubject.asObservable();
    this.user$.subscribe(value => {
      this.user = value;
    });
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
}
