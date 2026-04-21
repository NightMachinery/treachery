import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { firstValueFrom, Observable } from 'rxjs';
import { map, shareReplay } from 'rxjs/operators';
import { GameApiService } from '../game/game-api.service';
import { TgMessage } from '../models/models';

@Injectable({
  providedIn: 'root'
})
export class ChatApiService {
  messages$: Observable<TgMessage[]>;
  collapsed = true;

  constructor(private http: HttpClient, private gameApi: GameApiService) {
    this.messages$ = this.gameApi.snapshot$.pipe(
      map(snapshot => (snapshot ? snapshot.messages : [])),
      shareReplay(1)
    );
  }

  toggleCollapse(value?: boolean) {
    if (value === undefined) {
      this.collapsed = !this.collapsed;
    } else {
      this.collapsed = !!value;
    }
  }

  async sendMessage(message: string) {
    const gameId = this.gameApi.gameId$.value;
    await firstValueFrom(this.http.post(`/api/games/${gameId}/messages`, { message }, this.getRoomRequestOptions()));
    await this.gameApi.refreshSnapshot();
  }

  private getRoomRequestOptions() {
    let headers = new HttpHeaders();
    if (this.gameApi.roomAuth$.value) {
      headers = headers.set('X-Treachery-Room-Auth', this.gameApi.roomAuth$.value);
    }
    return { headers };
  }
}
