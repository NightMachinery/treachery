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
      this.selectedClue = this.guess?.clueCardId || null;
      this.selectedMeans = this.guess?.meansCardId || null;
    }
  }

  handleClueChange(player: TgPlayer) {
    if (!player.meansCards.some(card => card.id === this.selectedMeans)) {
      this.selectedMeans = null;
    }
    this.syncGuess(player);
  }

  handleMeansChange(player: TgPlayer) {
    if (!player.clueCards.some(card => card.id === this.selectedClue)) {
      this.selectedClue = null;
    }
    this.syncGuess(player);
  }

  private syncGuess(player: TgPlayer) {
    if (!this.selectedMeans && !this.selectedClue) {
      this.guess.murdererUid = null;
      this.guess.meansCardId = null;
      this.guess.clueCardId = null;
      return;
    }
    this.guess.murdererUid = player.uid;
    this.guess.meansCardId = this.selectedMeans;
    this.guess.clueCardId = this.selectedClue;
    this.guess.meansCardName = player.meansCards.find(card => card.id === this.selectedMeans)?.name || '';
    this.guess.clueCardName = player.clueCards.find(card => card.id === this.selectedClue)?.name || '';
  }

  otherPlayers(players: TgPlayer[], viewer: TgViewer) {
    return players ? players.filter(player => !viewer || player.uid !== viewer.uid) : [];
  }

  trackByPlayerUid(index: number, player: TgPlayer) {
    return player.uid;
  }
}
