import { BehaviorSubject } from 'rxjs';

import { RoomTimerComponent } from './room-timer.component';

describe('RoomTimerComponent', () => {
  function createComponent() {
    const snapshot$ = new BehaviorSubject<any>(null);
    const gameId$ = new BehaviorSubject<string>('ABCD');
    const gameApi = {
      snapshot$,
      gameId$,
      startRoomTimer: jasmine.createSpy('startRoomTimer').and.returnValue(Promise.resolve()),
      pauseRoomTimer: jasmine.createSpy('pauseRoomTimer').and.returnValue(Promise.resolve()),
      resumeRoomTimer: jasmine.createSpy('resumeRoomTimer').and.returnValue(Promise.resolve()),
      resetRoomTimer: jasmine.createSpy('resetRoomTimer').and.returnValue(Promise.resolve()),
      clearRoomTimer: jasmine.createSpy('clearRoomTimer').and.returnValue(Promise.resolve())
    };
    const snack = {
      error: jasmine.createSpy('error')
    };
    const component = new RoomTimerComponent(gameApi as any, snack as any);
    spyOn<any>(component, 'startTicker').and.callFake(() => undefined);
    spyOn<any>(component, 'playExpiryBeep').and.returnValue(Promise.resolve());
    return { component, snapshot$, gameApi, snack };
  }

  it('shows running countdown using the snapshot server timestamp offset', () => {
    const { component, snapshot$ } = createComponent();
    const now = new Date('2026-04-20T12:00:00Z').getTime();
    spyOn(Date, 'now').and.returnValue(now + 5000);

    component.ngOnInit();
    snapshot$.next({
      serverTimestamp: new Date(now).toISOString(),
      viewer: { isCreator: false },
      game: {
        roomTimer: {
          durationSeconds: 40,
          expiresAt: new Date(now + 10000).toISOString(),
          pausedRemainingSeconds: 0,
          runId: 1
        }
      }
    });

    expect(component.status).toBe('running');
    expect(component.remainingSeconds).toBe(5);
    expect(component.timeLabel).toBe('0:05');
  });

  it('normalizes the duration input before starting a new timer', async () => {
    const { component, gameApi } = createComponent();
    component.durationSeconds = 0;

    await component.startTimer();

    expect(component.durationSeconds).toBe(40);
    expect(gameApi.startRoomTimer).toHaveBeenCalledWith(40);
  });

  it('plays the expiry cue only once per room and run id', () => {
    const { component, snapshot$ } = createComponent();
    const now = new Date('2026-04-20T12:00:00Z').getTime();
    spyOn(Date, 'now').and.returnValue(now);
    sessionStorage.removeItem('treachery-room-timer-cue:ABCD:7');

    component.ngOnInit();
    const expiredSnapshot = {
      serverTimestamp: new Date(now).toISOString(),
      viewer: { isCreator: true },
      game: {
        roomTimer: {
          durationSeconds: 40,
          expiresAt: new Date(now - 1000).toISOString(),
          pausedRemainingSeconds: 0,
          runId: 7
        }
      }
    };

    snapshot$.next(expiredSnapshot);
    snapshot$.next(expiredSnapshot);

    expect(component.status).toBe('expired');
    expect((component as any).playExpiryBeep).toHaveBeenCalledTimes(1);
  });
});
