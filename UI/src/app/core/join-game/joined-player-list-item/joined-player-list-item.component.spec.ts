import { NO_ERRORS_SCHEMA } from '@angular/core';
import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';

import { AvatarService } from './../../../shared/api/avatar/avatar.service';
import { JoinedPlayerListItemComponent } from './joined-player-list-item.component';

describe('JoinedPlayerListItemComponent', () => {
  let component: JoinedPlayerListItemComponent;
  let fixture: ComponentFixture<JoinedPlayerListItemComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [JoinedPlayerListItemComponent],
      providers: [
        {
          provide: AvatarService,
          useValue: {
            getAvatar: () => 'avatar.svg'
          }
        }
      ],
      schemas: [NO_ERRORS_SCHEMA]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(JoinedPlayerListItemComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
