import { CommonModule } from '@angular/common';
import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { BehaviorSubject } from 'rxjs';
import { GameApiService } from '../../../shared/api/game/game-api.service';
import { TgGame, TgModeratorPrivateData, TgViewer } from '../../../shared/api/models/models';

import { ModeratorWitnessControlsComponent } from './moderator-witness-controls.component';

describe('ModeratorWitnessControlsComponent', () => {
  let component: ModeratorWitnessControlsComponent;
  let fixture: ComponentFixture<ModeratorWitnessControlsComponent>;
  let gameApi: jasmine.SpyObj<GameApiService> & {
    game$: BehaviorSubject<TgGame>;
    viewer$: BehaviorSubject<TgViewer>;
    moderatorPrivateData$: BehaviorSubject<TgModeratorPrivateData>;
  };

  beforeEach(
    waitForAsync(() => {
      gameApi = Object.assign(jasmine.createSpyObj<GameApiService>('GameApiService', ['showWitnessSelectionPrompt', 'updateRoomMods']), {
        game$: new BehaviorSubject<TgGame>({ pendingWitnessSelection: true, startedOn: '2026-04-20T00:00:00Z', meansCluesTextOnly: false } as TgGame),
        viewer$: new BehaviorSubject<TgViewer>({ isCreator: true } as TgViewer),
        moderatorPrivateData$: new BehaviorSubject<TgModeratorPrivateData>({
          witnessPromptCandidates: [
            { uid: 'creator', name: 'Creator' },
            { uid: 'p1', name: 'Player 1' }
          ]
        })
      });
      gameApi.showWitnessSelectionPrompt.and.returnValue(Promise.resolve());
      gameApi.updateRoomMods.and.returnValue(Promise.resolve());

      TestBed.configureTestingModule({
        imports: [CommonModule],
        declarations: [ModeratorWitnessControlsComponent],
        providers: [{ provide: GameApiService, useValue: gameApi }]
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(ModeratorWitnessControlsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should render the blind suspect list without role labels', () => {
    const panelText = fixture.nativeElement.textContent;

    expect(panelText).toContain('Room mods');
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
});
