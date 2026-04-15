import { of } from 'rxjs';
import * as util from './../util';

import { ForensicApiService } from './forensic-api.service';

describe('ForensicApiService', () => {
  it('creates a game before navigating and does not prefetch the snapshot', async () => {
    spyOn(util, 'randomReadableId').and.returnValue('abcd');

    const http = {
      post: jasmine.createSpy('post').and.returnValue(of({ success: true, gameId: 'WXYZ' }))
    };
    const router = {
      navigateByUrl: jasmine.createSpy('navigateByUrl')
    };
    const gameApi = {
      snapshot$: of(null),
      game$: of(null),
      setGameId: jasmine.createSpy('setGameId'),
      refreshSnapshot: jasmine.createSpy('refreshSnapshot')
    };

    const service = new ForensicApiService(http as any, {} as any, router as any, gameApi as any, {} as any);

    await service.createGame();

    expect(http.post).toHaveBeenCalledWith('/api/games', { gameId: 'abcd' });
    expect(gameApi.setGameId).not.toHaveBeenCalled();
    expect(gameApi.refreshSnapshot).not.toHaveBeenCalled();
    expect(router.navigateByUrl).toHaveBeenCalledWith('/forensic/WXYZ');
  });
});
