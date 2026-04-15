import { Injectable } from '@angular/core';
import ColorHash from 'color-hash';

@Injectable({
  providedIn: 'root'
})
export class AvatarService {
  private readonly colorHash = new ColorHash();

  getAvatar(uid: string) {
    const seed = uid || 'anonymous';
    const fill = this.colorHash.hex(seed).replace('#', '');
    const label = seed.slice(0, 2).toUpperCase();
    return `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' rx='32' fill='%23${fill}'/%3E%3Ctext x='32' y='38' font-size='22' text-anchor='middle' fill='white' font-family='Roboto,Arial,sans-serif'%3E${label}%3C/text%3E%3C/svg%3E`;
  }
}
