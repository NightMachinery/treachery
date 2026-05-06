import { NO_ERRORS_SCHEMA } from '@angular/core';
import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { BehaviorSubject } from 'rxjs';

import { GameApiService } from './../../../shared/api/game/game-api.service';
import { JoinedPlayersListComponent } from './joined-players-list.component';

describe('JoinedPlayersListComponent', () => {
  let component: JoinedPlayersListComponent;
  let fixture: ComponentFixture<JoinedPlayersListComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [JoinedPlayersListComponent],
      providers: [
        {
          provide: GameApiService,
          useValue: {
            game$: new BehaviorSubject(null),
            participants$: new BehaviorSubject([]),
            setParticipantRole: jasmine.createSpy('setParticipantRole'),
            toggleScientist: jasmine.createSpy('toggleScientist'),
          },
        },
      ],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(JoinedPlayersListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
