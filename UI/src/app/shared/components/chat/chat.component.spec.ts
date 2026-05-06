import { NO_ERRORS_SCHEMA } from '@angular/core';
import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { BehaviorSubject } from 'rxjs';

import { GameApiService } from './../../api/game/game-api.service';
import { ChatApiService } from './../../api/chat/chat-api.service';
import { ChatComponent } from './chat.component';

describe('ChatComponent', () => {
  let component: ChatComponent;
  let fixture: ComponentFixture<ChatComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ChatComponent],
      providers: [
        {
          provide: ChatApiService,
          useValue: {
            collapsed: true,
            messages$: new BehaviorSubject([]),
            toggleCollapse: jasmine.createSpy('toggleCollapse'),
          },
        },
        {
          provide: GameApiService,
          useValue: {
            participantsDict$: new BehaviorSubject(new Map()),
            game$: new BehaviorSubject(null),
            viewer$: new BehaviorSubject(null),
          },
        },
      ],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ChatComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
