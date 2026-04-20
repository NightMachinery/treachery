import { NO_ERRORS_SCHEMA } from '@angular/core';
import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { BehaviorSubject } from 'rxjs';

import { ForensicApiService } from './../../shared/api/forensic/forensic-api.service';
import { GameApiService } from './../../shared/api/game/game-api.service';
import { AllGamesComponent } from './all-games.component';

describe('AllGamesComponent', () => {
  let component: AllGamesComponent;
  let fixture: ComponentFixture<AllGamesComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [AllGamesComponent],
      providers: [
        {
          provide: ForensicApiService,
          useValue: {}
        },
        {
          provide: GameApiService,
          useValue: {
            activeGames$: new BehaviorSubject([])
          }
        },
        {
          provide: Router,
          useValue: {
            navigateByUrl: jasmine.createSpy('navigateByUrl')
          }
        }
      ],
      schemas: [NO_ERRORS_SCHEMA]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AllGamesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
