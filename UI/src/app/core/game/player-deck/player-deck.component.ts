import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { TgCard, TgPlayer } from '../../../shared/api/models/models';

@Component({
  selector: 'app-player-deck',
  standalone: false,
  templateUrl: './player-deck.component.html',
  styleUrls: ['./player-deck.component.scss']
})
export class PlayerDeckComponent implements OnInit {
  @Input() player: TgPlayer;
  @Input() disableSelection = false;
  @Input() selectedClue: string;
  @Output() selectedClueChange = new EventEmitter<string>();
  @Input() selectedMeans: string;
  @Output() selectedMeansChange = new EventEmitter<string>();
  @Input() disableSelectionDisplay = false;
  @Input() selectedSuspectUid: string;

  constructor() {}

  clueClick(cardName) {
    this.selectedClue = this.selectedClue === cardName ? null : cardName;
    this.selectedClueChange.emit(this.selectedClue);
  }

  meansClick(cardName) {
    this.selectedMeans = this.selectedMeans === cardName ? null : cardName;
    this.selectedMeansChange.emit(this.selectedMeans);
  }

  ngOnInit() {}

  get deckSelected() {
    return !!this.selectedSuspectUid && this.selectedSuspectUid === this.player?.uid;
  }

  get deckDimmed() {
    return !!this.selectedSuspectUid && this.selectedSuspectUid !== this.player?.uid;
  }

  shouldSubdueMeans(cardName: string) {
    return this.deckSelected && !!this.selectedMeans && this.selectedMeans !== cardName;
  }

  shouldSubdueClue(cardName: string) {
    return this.deckSelected && !!this.selectedClue && this.selectedClue !== cardName;
  }

  trackByCardName(index: number, card: TgCard) {
    return card.name;
  }
}
