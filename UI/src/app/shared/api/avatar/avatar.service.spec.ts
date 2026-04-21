import { TestBed } from '@angular/core/testing';

import { AvatarService } from './avatar.service';
import { buildAvatarSeed } from './avatar-generator';

describe('AvatarService', () => {
  let service: AvatarService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(AvatarService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('builds the avatar seed with the uid prepended to the display name', () => {
    expect(buildAvatarSeed('user-123', 'Detective Nova')).toBe('user-123:Detective Nova');
  });

  it('returns the same avatar for the same uid and display name', async () => {
    const first = await service.getAvatar('user-123', 'Detective Nova');
    const second = await service.getAvatar('user-123', 'Detective Nova');

    expect(first).toBe(second);
  });

  it('returns different avatars for different user ids with the same display name', async () => {
    const first = await service.getAvatar('user-123', 'Detective Nova');
    const second = await service.getAvatar('user-456', 'Detective Nova');

    expect(first).not.toBe(second);
  });

  it('falls back to uid-only and anonymous seeds when needed', async () => {
    const uidOnly = await service.getAvatar('user-123', '');
    const anonymous = await service.getAvatar('', '');

    expect(uidOnly).toContain('data:image/svg+xml');
    expect(anonymous).toContain('data:image/svg+xml');
    expect(uidOnly).not.toBe(anonymous);
  });
});
