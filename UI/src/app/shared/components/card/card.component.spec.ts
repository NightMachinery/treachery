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

  it('shrinks long labels with the fit-text hook instead of overflowing', () => {
    const label = fixture.nativeElement.querySelector('.card-name') as HTMLElement;
    component.card = { ...component.card, name: 'Prescription' };
    component.cardNameElement = { nativeElement: label } as any;
    Object.defineProperty(label, 'clientWidth', { configurable: true, value: 80 });
    Object.defineProperty(label, 'clientHeight', { configurable: true, value: 36 });
    spyOn<any>(component, 'labelFits').and.callFake(() => {
      const scale = Number(label.style.getPropertyValue('--tg-card-name-scale') || 1);
      return scale <= 0.76;
    });

    component.fitCardLabel();

    expect(Number(label.style.getPropertyValue('--tg-card-name-scale'))).toBeLessThan(1);
    expect(label.classList).toContain('card-name--compressed');
  });
});
