import { TgForensicCard } from '../../../shared/api/models/models';
import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';

@Component({
  selector: 'app-forensic-card',
  standalone: false,
  templateUrl: './forensic-card.component.html',
  styleUrls: ['./forensic-card.component.scss']
})
export class ForensicCardComponent implements OnInit {
  @Input() forensicCard: TgForensicCard;
  @Input() disabled: boolean;
  @Input() selected: boolean;
  @Input() selectedOptionId: string;
  @Output() selectedOptionIdChange = new EventEmitter<string>();

  constructor() {}

  ngOnInit() {}

  isSelected(choiceId: string) {
    return choiceId === this.forensicCard.selectedChoiceId || (this.selected && choiceId === this.selectedOptionId);
  }

  choiceClick(choiceId: string) {
    this.selectedOptionIdChange.emit(choiceId);
  }
}
