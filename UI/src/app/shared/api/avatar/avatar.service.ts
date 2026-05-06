import { Injectable } from '@angular/core';
import { buildAvatarSeed, renderAvatarDataUri } from './avatar-generator';

@Injectable({
  providedIn: 'root',
})
export class AvatarService {
  private readonly avatarCache = new Map<string, Promise<string>>();

  getAvatar(uid?: string, displayName?: string) {
    const seed = buildAvatarSeed(uid, displayName);

    if (!this.avatarCache.has(seed)) {
      this.avatarCache.set(seed, Promise.resolve(renderAvatarDataUri(uid, displayName)));
    }

    return this.avatarCache.get(seed)!;
  }
}
