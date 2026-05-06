import { NO_ERRORS_SCHEMA } from '@angular/core';
import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { BehaviorSubject } from 'rxjs';

import { AuthService } from './../../api/auth/auth.service';
import { GameApiService } from './../../api/game/game-api.service';
import { NavbarComponent } from './navbar.component';

describe('NavbarComponent', () => {
  let component: NavbarComponent;
  let fixture: ComponentFixture<NavbarComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [NavbarComponent],
      providers: [
        {
          provide: AuthService,
          useValue: {
            user$: new BehaviorSubject(null),
            displayName$: new BehaviorSubject('Player'),
            promptForDisplayName: jasmine.createSpy('promptForDisplayName').and.resolveTo(false),
          },
        },
        {
          provide: GameApiService,
          useValue: {
            viewer$: new BehaviorSubject(null),
            game$: new BehaviorSubject(null),
            gameId$: new BehaviorSubject(null),
            roomAuth$: new BehaviorSubject(null),
            refreshSnapshot: jasmine.createSpy('refreshSnapshot'),
          },
        },
      ],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(NavbarComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
