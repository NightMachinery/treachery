import { NO_ERRORS_SCHEMA } from '@angular/core';
import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';

import { ActiveGamesListItemComponent } from './active-games-list-item.component';

describe('ActiveGamesListItemComponent', () => {
  let component: ActiveGamesListItemComponent;
  let fixture: ComponentFixture<ActiveGamesListItemComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ActiveGamesListItemComponent],
      schemas: [NO_ERRORS_SCHEMA]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ActiveGamesListItemComponent);
    component = fixture.componentInstance;
    component.game = { gameId: 'ABCD' } as any;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
