import { of } from 'rxjs';
import * as util from './../util';

import { ForensicApiService } from './forensic-api.service';

describe('ForensicApiService', () => {
  it('creates a game before navigating and does not prefetch the snapshot', async () => {
    spyOn(util, 'randomReadableId').and.returnValue('abcd');

    const http = {
      post: jasmine.createSpy('post').and.returnValue(of({ success: true, gameId: 'WXYZ' }))
    };
    const gameApi = {
      snapshot$: of(null),
      game$: of(null),
      setGameContext: jasmine.createSpy('setGameContext'),
      navigateTo: jasmine.createSpy('navigateTo').and.resolveTo(true)
    };
    const authService = {
      ensureDisplayName: jasmine.createSpy('ensureDisplayName').and.resolveTo(true)
    };

    const service = new ForensicApiService(http as any, authService as any, gameApi as any, {} as any);

    await service.createGame();

    expect(http.post).toHaveBeenCalledWith('/api/games', { gameId: 'abcd' });
    expect(authService.ensureDisplayName).toHaveBeenCalled();
    expect(gameApi.setGameContext).toHaveBeenCalledWith('WXYZ', null);
    expect(gameApi.navigateTo).toHaveBeenCalledWith('join', 'WXYZ');
  });
});
