import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { BehaviorSubject } from 'rxjs';
import { ChatApiService } from 'src/app/shared/api/chat/chat-api.service';
import { GameApiService } from '../../../shared/api/game/game-api.service';
import { TgGuess, TgPlayer, TgViewer } from '../../../shared/api/models/models';
import { SnackBarService } from 'src/app/shared/api/snack-bar/snack-bar.service';

import { GuessComponent } from './guess.component';

describe('GuessComponent', () => {
  let component: GuessComponent;
  let fixture: ComponentFixture<GuessComponent>;
  let gameApi: jasmine.SpyObj<GameApiService> & {
    guesses$: BehaviorSubject<TgGuess[]>;
    players$: BehaviorSubject<TgPlayer[]>;
    playersDict$: BehaviorSubject<Map<string, TgPlayer>>;
    viewer$: BehaviorSubject<TgViewer>;
  };
  let snack: jasmine.SpyObj<SnackBarService>;
  let chatApi: jasmine.SpyObj<ChatApiService>;

  const murderer: TgPlayer = {
    uid: 'murderer',
    name: 'Murderer',
    guessed: false,
    meansCards: [
      { id: 'rope', name: 'Rope', imgUrl: 'rope.jpg', altImgUrl: 'rope-fallback.jpg', hasImage: true, guessedBy: [] },
      { id: 'knife', name: 'Knife', imgUrl: 'knife.jpg', altImgUrl: 'knife-fallback.jpg', hasImage: true, guessedBy: [] }
    ],
    clueCards: [
      { id: 'shoes', name: 'Shoes', imgUrl: 'shoes.jpg', altImgUrl: 'shoes-fallback.jpg', hasImage: true, guessedBy: [] },
      { id: 'watch', name: 'Watch', imgUrl: 'watch.jpg', altImgUrl: 'watch-fallback.jpg', hasImage: true, guessedBy: [] }
    ]
  };

  beforeEach(
    waitForAsync(() => {
      gameApi = Object.assign(jasmine.createSpyObj<GameApiService>('GameApiService', ['makeGuess']), {
        guesses$: new BehaviorSubject<TgGuess[]>([]),
        players$: new BehaviorSubject<TgPlayer[]>([murderer]),
        playersDict$: new BehaviorSubject<Map<string, TgPlayer>>(new Map([[murderer.uid, murderer]])),
        viewer$: new BehaviorSubject<TgViewer>({ uid: 'investigator' } as TgViewer)
      });
      gameApi.makeGuess.and.returnValue(Promise.resolve());
      snack = jasmine.createSpyObj<SnackBarService>('SnackBarService', ['error']);
      chatApi = jasmine.createSpyObj<ChatApiService>('ChatApiService', ['sendMessage']);

      TestBed.configureTestingModule({
        declarations: [GuessComponent],
        providers: [
          { provide: GameApiService, useValue: gameApi },
          { provide: SnackBarService, useValue: snack },
          { provide: ChatApiService, useValue: chatApi }
        ],
        schemas: [NO_ERRORS_SCHEMA]
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(GuessComponent);
    component = fixture.componentInstance;
    component.guess = {
      murdererUid: 'murderer',
      meansCardId: 'knife',
      meansCardName: 'Knife',
      clueCardId: 'shoes',
      clueCardName: 'Shoes'
    };
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('blocks an exact duplicate full guess before calling the API', () => {
    gameApi.guesses$.next([
      {
        guessedByUid: 'someone-else',
        murdererUid: 'murderer',
        meansCardId: 'knife',
        meansCardName: 'Knife',
        clueCardId: 'shoes',
        clueCardName: 'Shoes',
        correct: false
      }
    ]);

    component.makeGuess();

    expect(snack.error).toHaveBeenCalledWith('That exact guess was already submitted.');
    expect(gameApi.makeGuess).not.toHaveBeenCalled();
    expect(chatApi.sendMessage).not.toHaveBeenCalled();
  });
});
