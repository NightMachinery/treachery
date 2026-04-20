import { Component, ElementRef, Input, OnInit, ViewChild } from '@angular/core';

@Component({
  selector: 'tg-copy-link',
  standalone: false,
  templateUrl: './copy-link.component.html',
  styleUrls: ['./copy-link.component.scss']
})
export class CopyLinkComponent implements OnInit {
  @Input() text: string;
  @ViewChild('copyInput') copyInput: ElementRef<HTMLInputElement>;

  constructor() {}

  ngOnInit(): void {}

  async copyText() {
    const text = this.text || '';
    if (!text) {
      return;
    }

    if (navigator.clipboard && window.isSecureContext) {
      try {
        await navigator.clipboard.writeText(text);
        return;
      } catch (error) {
        console.warn('navigator.clipboard failed, falling back to execCommand', error);
      }
    }

    const input = this.copyInput.nativeElement;
    input.focus();
    input.select();
    input.setSelectionRange(0, text.length);
    document.execCommand('copy');
    input.blur();
  }
}
