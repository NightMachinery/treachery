import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';

import { PlayerDeckComponent } from './player-deck.component';

describe('PlayerDeckComponent', () => {
  let component: PlayerDeckComponent;
  let fixture: ComponentFixture<PlayerDeckComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [PlayerDeckComponent]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(PlayerDeckComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
