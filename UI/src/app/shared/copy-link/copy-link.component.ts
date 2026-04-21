import { Component, Input, OnInit } from '@angular/core';
import { copyTextToClipboard } from '../utils/clipboard';

@Component({
  selector: 'tg-copy-link',
  standalone: false,
  templateUrl: './copy-link.component.html',
  styleUrls: ['./copy-link.component.scss']
})
export class CopyLinkComponent implements OnInit {
  @Input() text: string;
  copied = false;
  private copyTimeout: any;

  constructor() {}

  ngOnInit(): void {}

  async copyText() {
    const text = this.text || '';
    if (!text) {
      return;
    }

    const copied = await copyTextToClipboard(text);
    if (copied) {
      this.showCopiedFeedback();
    }
  }

  private showCopiedFeedback() {
    this.copied = true;
    if (this.copyTimeout) {
      clearTimeout(this.copyTimeout);
    }
    this.copyTimeout = setTimeout(() => {
      this.copied = false;
    }, 2000);
  }
}
