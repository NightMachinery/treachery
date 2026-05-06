import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';
import { ForensicApiService } from './../../shared/api/forensic/forensic-api.service';
import { TgGame } from '../../shared/api/models/models';
import { GameApiService } from '../../shared/api/game/game-api.service';

@Component({
  selector: 'app-all-games',
  standalone: false,
  templateUrl: './all-games.component.html',
  styleUrls: ['./all-games.component.scss'],
})
export class AllGamesComponent implements OnInit {
  games$: Observable<TgGame[]>;

  constructor(
    public forensicApi: ForensicApiService,
    public gameApi: GameApiService,
    private router: Router,
  ) {
    this.games$ = this.gameApi.activeGames$;
  }

  ngOnInit() {}

  joinGame(gameId) {
    this.router.navigateByUrl(`/join/${gameId.toUpperCase()}`);
  }
}
