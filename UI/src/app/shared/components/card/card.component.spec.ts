import { NO_ERRORS_SCHEMA } from '@angular/core';
import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { BehaviorSubject } from 'rxjs';

import { GameApiService } from '../../api/game/game-api.service';
import { TgGame } from '../../api/models/models';
import { CardComponent } from './card.component';

describe('CardComponent', () => {
  let component: CardComponent;
  let fixture: ComponentFixture<CardComponent>;
  let game$: BehaviorSubject<TgGame>;

  beforeEach(waitForAsync(() => {
    game$ = new BehaviorSubject<TgGame>({ meansCluesTextOnly: false } as TgGame);
    TestBed.configureTestingModule({
      declarations: [CardComponent],
      providers: [{ provide: GameApiService, useValue: { game$ } }],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CardComponent);
    component = fixture.componentInstance;
    component.card = {
      id: 'card-1',
      name: 'שלום world',
      imgUrl: 'card.jpg',
      altImgUrl: 'fallback.jpg',
      guessedBy: [],
      hasImage: true,
    };
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should render the image in normal mode', () => {
    expect(fixture.nativeElement.querySelector('img')).toBeTruthy();
    expect(fixture.nativeElement.querySelector('.card').classList).toContain('image-card');
    expect(fixture.nativeElement.querySelector('.card-name').getAttribute('dir')).toBe('auto');
  });

  it('should render text-only mode without an image', () => {
    game$.next({ meansCluesTextOnly: true } as TgGame);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('img')).toBeFalsy();
    expect(fixture.nativeElement.querySelector('.text-only-body')).toBeTruthy();
    expect(fixture.nativeElement.querySelector('.text-only-body').getAttribute('dir')).toBe('auto');
    expect(fixture.nativeElement.querySelector('.card-name').getAttribute('dir')).toBe('auto');
  });
});
