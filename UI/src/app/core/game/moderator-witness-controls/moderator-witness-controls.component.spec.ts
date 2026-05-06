import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { BehaviorSubject } from 'rxjs';
import { CardApiService } from '../../../shared/api/card/card-api.service';
import { GameApiService } from '../../../shared/api/game/game-api.service';
import {
  TgCrimePackCatalogEntry,
  TgGame,
  TgHintPackCatalogEntry,
  TgModeratorPrivateData,
  TgViewer,
  TgWordpackCatalog,
} from '../../../shared/api/models/models';

import { ModeratorWitnessControlsComponent } from './moderator-witness-controls.component';

describe('ModeratorWitnessControlsComponent', () => {
  let component: ModeratorWitnessControlsComponent;
  let fixture: ComponentFixture<ModeratorWitnessControlsComponent>;
  let gameApi: jasmine.SpyObj<GameApiService> & {
    game$: BehaviorSubject<TgGame>;
    viewer$: BehaviorSubject<TgViewer>;
    moderatorPrivateData$: BehaviorSubject<TgModeratorPrivateData>;
  };
  let cardApi: { catalog$: BehaviorSubject<TgWordpackCatalog> };

  const crimePack = {
    id: 'treachery',
    name: 'Treachery',
    defaultLanguage: 'en',
    defaultAssetSetId: 'treachery',
    meansCount: 90,
    clueCount: 200,
    hasAnyImages: true,
    languages: [
      { id: 'en', name: 'English' },
      { id: 'fa', name: 'فارسی' },
    ],
    assetSets: [
      { id: 'treachery', name: 'Treachery', hasAnyImages: true },
      { id: 'sketch', name: 'Sketch', hasAnyImages: true },
    ],
  } as TgCrimePackCatalogEntry;

  const hintPack = {
    id: 'treachery-hints',
    name: 'Treachery Hints',
    defaultLanguage: 'en',
    causeCount: 1,
    locationCount: 4,
    otherCount: 24,
    languages: [
      { id: 'en', name: 'English' },
      { id: 'fa', name: 'فارسی' },
    ],
  } as TgHintPackCatalogEntry;

  beforeEach(waitForAsync(() => {
    gameApi = Object.assign(jasmine.createSpyObj<GameApiService>('GameApiService', ['showWitnessSelectionPrompt', 'updateRoomMods']), {
      game$: new BehaviorSubject<TgGame>({
        pendingWitnessSelection: true,
        startedOn: '2026-04-20T00:00:00Z',
        meansCluesTextOnly: false,
        crimePackId: crimePack.id,
        crimePackLanguage: 'en',
        crimePackAssetSetId: 'treachery',
        hintPackId: hintPack.id,
        hintPackLanguage: 'en',
      } as TgGame),
      viewer$: new BehaviorSubject<TgViewer>({ isCreator: true } as TgViewer),
      moderatorPrivateData$: new BehaviorSubject<TgModeratorPrivateData>({
        witnessPromptCandidates: [
          { uid: 'creator', name: 'Creator' },
          { uid: 'p1', name: 'Player 1' },
        ],
      }),
    });
    gameApi.showWitnessSelectionPrompt.and.returnValue(Promise.resolve());
    gameApi.updateRoomMods.and.returnValue(Promise.resolve());
    cardApi = {
      catalog$: new BehaviorSubject<TgWordpackCatalog>({
        crimePacks: [crimePack],
        hintPacks: [hintPack],
      }),
    };

    TestBed.configureTestingModule({
      imports: [CommonModule, FormsModule],
      declarations: [ModeratorWitnessControlsComponent],
      providers: [
        { provide: GameApiService, useValue: gameApi },
        { provide: CardApiService, useValue: cardApi },
      ],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ModeratorWitnessControlsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should render the blind suspect list without role labels', () => {
    const panelText = fixture.nativeElement.textContent;

    expect(panelText).toContain('Room mods');
    expect(panelText).toContain('CrimePack language');
    expect(panelText).toContain('HintPack language');
    expect(panelText).toContain('Means/clues text only');
    expect(panelText).toContain('Witness prompt controls');
    expect(panelText).toContain('Creator');
    expect(panelText).toContain('Player 1');
    expect(panelText).not.toContain('murderer');
    expect(panelText).not.toContain('accomplice');
    expect(panelText).not.toContain('prompt active');
  });

  it('should send the selected suspect uid when a prompt button is clicked', async () => {
    const buttons = fixture.debugElement.queryAll(By.css('.target-row button'));

    buttons[1].nativeElement.click();
    await fixture.whenStable();

    expect(gameApi.showWitnessSelectionPrompt).toHaveBeenCalledWith('p1');
  });

  it('should send the room-mod value when the text-only toggle changes', async () => {
    const toggle = fixture.debugElement.query(By.css('.toggle-row input'));

    toggle.nativeElement.checked = true;
    toggle.nativeElement.dispatchEvent(new Event('change'));
    await fixture.whenStable();

    expect(gameApi.updateRoomMods).toHaveBeenCalledWith({ meansCluesTextOnly: true });
  });

  it('should send live language and asset changes', async () => {
    const selects = fixture.debugElement.queryAll(By.css('.settings-grid select'));

    selects[0].nativeElement.value = selects[0].nativeElement.options[1].value;
    selects[0].nativeElement.dispatchEvent(new Event('change'));
    selects[1].nativeElement.value = selects[1].nativeElement.options[1].value;
    selects[1].nativeElement.dispatchEvent(new Event('change'));
    selects[2].nativeElement.value = selects[2].nativeElement.options[1].value;
    selects[2].nativeElement.dispatchEvent(new Event('change'));
    await fixture.whenStable();

    expect(gameApi.updateRoomMods).toHaveBeenCalledWith({ crimePackLanguage: 'fa' });
    expect(gameApi.updateRoomMods).toHaveBeenCalledWith({ crimePackAssetSetId: 'sketch' });
    expect(gameApi.updateRoomMods).toHaveBeenCalledWith({ hintPackLanguage: 'fa' });
  });
});
