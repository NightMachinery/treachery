import { Component, Input, OnChanges, OnInit, SimpleChanges } from '@angular/core';
import { Observable } from 'rxjs';
import { GameApiService } from './../../../shared/api/game/game-api.service';
import { TgPartialGuess, TgPlayer, TgViewer } from '../../../shared/api/models/models';

@Component({
  selector: 'app-player-deck-pager',
  standalone: false,
  templateUrl: './player-deck-pager.component.html',
  styleUrls: ['./player-deck-pager.component.scss']
})
export class PlayerDeckPagerComponent implements OnInit, OnChanges {
  players$: Observable<TgPlayer[]>;
  viewer$: Observable<TgViewer>;
  @Input() guess: TgPartialGuess = {} as TgPartialGuess;
  @Input() disableSelection: boolean;
  selectedClue: string;
  selectedMeans: string;
  player$: Observable<TgPlayer>;

  constructor(gameApi: GameApiService) {
    this.players$ = gameApi.getCurrentGamePlayers();
    this.player$ = gameApi.getCurrentGamePlayer();
    this.viewer$ = gameApi.getCurrentViewer();
  }

  ngOnInit() {}

  ngOnChanges(changes: SimpleChanges) {
    if (changes.guess) {
      this.selectedClue = this.guess?.clueCardName || null;
      this.selectedMeans = this.guess?.meansCardName || null;
    }
  }

  handleClueChange(player: TgPlayer) {
    if (!player.meansCards.some(card => card.name === this.selectedMeans)) {
      this.selectedMeans = null;
    }
    this.syncGuess(player);
  }

  handleMeansChange(player: TgPlayer) {
    if (!player.clueCards.some(card => card.name === this.selectedClue)) {
      this.selectedClue = null;
    }
    this.syncGuess(player);
  }

  private syncGuess(player: TgPlayer) {
    if (!this.selectedMeans && !this.selectedClue) {
      this.guess.murdererUid = null;
      this.guess.meansCardName = null;
      this.guess.clueCardName = null;
      return;
    }
    this.guess.murdererUid = player.uid;
    this.guess.meansCardName = this.selectedMeans;
    this.guess.clueCardName = this.selectedClue;
  }

  otherPlayers(players: TgPlayer[], viewer: TgViewer) {
    return players ? players.filter(player => !viewer || player.uid !== viewer.uid) : [];
  }

  trackByPlayerUid(index: number, player: TgPlayer) {
    return player.uid;
  }
}
