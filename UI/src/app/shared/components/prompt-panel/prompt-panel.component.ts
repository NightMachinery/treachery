import { Component, EventEmitter, Input, Output } from '@angular/core';

@Component({
  selector: 'tg-prompt-panel',
  standalone: false,
  templateUrl: './prompt-panel.component.html',
  styleUrls: ['./prompt-panel.component.scss']
})
export class PromptPanelComponent {
  @Input() title = '';
  @Input() dismissible = false;
  @Input() minimized = false;
  @Output() minimizedChange = new EventEmitter<boolean>();
  @Output() dismissed = new EventEmitter<void>();

  toggleMinimized() {
    this.minimized = !this.minimized;
    this.minimizedChange.emit(this.minimized);
  }

  dismiss() {
    this.dismissed.emit();
  }
}
