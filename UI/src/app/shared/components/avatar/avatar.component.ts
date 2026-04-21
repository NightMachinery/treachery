import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
import { AvatarService } from './../../api/avatar/avatar.service';
import { getAvatarFallbackColor } from '../../api/avatar/avatar-generator';

@Component({
  selector: 'tg-avatar',
  standalone: false,
  templateUrl: './avatar.component.html',
  styleUrls: ['./avatar.component.scss']
})
export class AvatarComponent implements OnChanges {
  @Input() uid: string;
  @Input() name?: string;
  @Input() diameter = 50;
  src = this.defaultSvgUri();
  private latestRequestId = 0;

  constructor(private avatar: AvatarService) {}

  ngOnChanges(_: SimpleChanges): void {
    void this.loadAvatar();
  }

  defaultSvgUri() {
    const fill = getAvatarFallbackColor(this.uid, this.name).replace('#', '');
    return `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23${fill}' width='18px' height='18px'%3E%3Cpath d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z'/%3E%3Cpath d='M0 0h24v24H0z' fill='none'/%3E%3C/svg%3E`;
  }

  async loadAvatar() {
    const requestId = ++this.latestRequestId;
    const fallback = this.defaultSvgUri();
    this.src = fallback;

    try {
      const src = await this.avatar.getAvatar(this.uid, this.name);

      if (requestId !== this.latestRequestId) {
        return;
      }

      this.src = src || fallback;
    } catch {
      if (requestId === this.latestRequestId) {
        this.src = fallback;
      }
    }
  }

  onImageError(event: Event) {
    const img = event.target as HTMLImageElement;
    const fallback = this.defaultSvgUri();
    if (img && img.src !== fallback) {
      img.src = fallback;
    }
  }
}
