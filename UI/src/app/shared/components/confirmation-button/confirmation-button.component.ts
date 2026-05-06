import { Component, EventEmitter, Input, Output } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { firstValueFrom } from 'rxjs';
import { ConfirmActionDialogComponent } from '../confirm-action-dialog/confirm-action-dialog.component';

@Component({
  selector: 'tg-confirmation-button',
  standalone: false,
  templateUrl: './confirmation-button.component.html',
  styleUrls: ['./confirmation-button.component.scss'],
})
export class ConfirmationButtonComponent {
  @Input() mode: 'inline' | 'dialog' = 'inline';
  @Input() color: 'primary' | 'accent' | 'warn' = 'accent';
  @Input() variant: 'default' | 'danger' = 'default';
  @Input() confirmTitle = 'Confirm action';
  @Input() confirmMessage = 'Are you sure you want to continue?';
  @Input() confirmButtonLabel = 'Confirm';
  @Input() cancelButtonLabel = 'Cancel';
  @Output() whenClicked = new EventEmitter();
  clicked = false;

  constructor(private dialog: MatDialog) {}

  get buttonColor() {
    return this.variant === 'danger' ? 'warn' : this.color;
  }

  async handleClick($event: Event) {
    if (this.mode === 'dialog') {
      const dialogRef = this.dialog.open(ConfirmActionDialogComponent, {
        autoFocus: false,
        restoreFocus: true,
        disableClose: false,
        hasBackdrop: true,
        width: 'min(28rem, calc(100vw - 1.5rem))',
        panelClass: 'tg-confirm-dialog',
        data: {
          title: this.confirmTitle,
          message: this.confirmMessage,
          confirmLabel: this.confirmButtonLabel,
          cancelLabel: this.cancelButtonLabel,
          variant: this.variant,
        },
      });
      const confirmed = await firstValueFrom(dialogRef.afterClosed());
      if (confirmed) {
        this.whenClicked.emit($event);
      }
      return;
    }

    if (this.clicked) {
      this.clicked = false;
      this.whenClicked.emit($event);
    } else {
      this.clicked = true;
    }
  }
}
