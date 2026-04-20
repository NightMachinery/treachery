import { provideHttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { BehaviorSubject } from 'rxjs';

import { AuthService } from './../auth/auth.service';
import { SnackBarService } from '../snack-bar/snack-bar.service';
import { GameApiService } from './game-api.service';

describe('GameApiService', () => {
  beforeEach(() =>
    TestBed.configureTestingModule({
      imports: [RouterTestingModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        GameApiService,
        {
          provide: AuthService,
          useValue: {
            user$: new BehaviorSubject(null),
            user: null,
            getStoredToken: () => ''
          }
        },
        {
          provide: SnackBarService,
          useValue: {
            error: jasmine.createSpy('error')
          }
        }
      ]
    })
  );

  it('should be created', () => {
    const service: GameApiService = TestBed.inject(GameApiService);
    expect(service).toBeTruthy();
  });
});
