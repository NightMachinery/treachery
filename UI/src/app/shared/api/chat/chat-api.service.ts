import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { map, shareReplay } from 'rxjs/operators';
import { GameApiService } from '../game/game-api.service';
import { TgMessage, TgMessageType } from '../models/models';
import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class ChatApiService {
  messages$: Observable<TgMessage[]>;
  collapsed = false;

  constructor(private http: HttpClient, private gameApi: GameApiService) {
    this.messages$ = this.gameApi.snapshot$.pipe(
      map(snapshot => snapshot ? snapshot.messages : []),
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

  async sendMessage(message: string, type: TgMessageType = TgMessageType.CHAT) {
    const gameId = this.gameApi.gameId$.value;
    await this.http.post(`/api/games/${gameId}/messages`, { message }).toPromise();
    await this.gameApi.refreshSnapshot();
  }
}
