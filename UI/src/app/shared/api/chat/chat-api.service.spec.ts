import { provideHttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { BehaviorSubject } from 'rxjs';

import { GameApiService } from '../game/game-api.service';
import { ChatApiService } from './chat-api.service';

describe('ChatApiService', () => {
  let service: ChatApiService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        {
          provide: GameApiService,
          useValue: {
            gameId$: new BehaviorSubject(null),
            roomAuth$: new BehaviorSubject(null),
            snapshot$: new BehaviorSubject(null),
            refreshSnapshot: jasmine.createSpy('refreshSnapshot'),
          },
        },
      ],
    });
    service = TestBed.inject(ChatApiService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
