import { Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';

export interface TgDisplayNameDialogData {
  initialValue?: string;
  title?: string;
  helperText?: string;
  submitLabel?: string;
}

@Component({
  selector: 'tg-display-name-dialog',
  standalone: false,
  templateUrl: './display-name-dialog.component.html',
  styleUrls: ['./display-name-dialog.component.scss']
})
export class DisplayNameDialogComponent {
  displayName = '';

  constructor(
    private dialogRef: MatDialogRef<DisplayNameDialogComponent, string | null>,
    @Inject(MAT_DIALOG_DATA) public data: TgDisplayNameDialogData
  ) {
    this.displayName = data?.initialValue || '';
  }

  cancel() {
    this.dialogRef.close(null);
  }

  save() {
    const trimmed = (this.displayName || '').trim();
    if (!trimmed) {
      return;
    }
    this.dialogRef.close(trimmed);
  }
}
