import { Component, Input } from '@angular/core';
import { By } from '@angular/platform-browser';
import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';

import { AvatarService } from './../../../shared/api/avatar/avatar.service';
import { TgParticipant } from '../../../shared/api/models/models';
import { JoinedPlayerListItemComponent } from './joined-player-list-item.component';

@Component({
  selector: 'tg-avatar',
  template: '',
})
class AvatarStubComponent {
  @Input() uid: string;
  @Input() name?: string;
}

describe('JoinedPlayerListItemComponent', () => {
  let component: JoinedPlayerListItemComponent;
  let fixture: ComponentFixture<JoinedPlayerListItemComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [JoinedPlayerListItemComponent, AvatarStubComponent],
      providers: [
        {
          provide: AvatarService,
          useValue: {
            getAvatar: () => 'avatar.svg',
          },
        },
      ],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(JoinedPlayerListItemComponent);
    component = fixture.componentInstance;
    component.participant = {
      uid: 'observer-1',
      name: 'Quiet Watcher',
      role: 'observer',
      isCreator: false,
      isScientist: false,
      isMarkedScientist: false,
    } as TgParticipant;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('renders an avatar for observer participants', () => {
    const avatar = fixture.debugElement.query(By.directive(AvatarStubComponent)).componentInstance as AvatarStubComponent;

    expect(avatar).toBeTruthy();
    expect(avatar.uid).toBe('observer-1');
    expect(avatar.name).toBe('Quiet Watcher');
    expect(fixture.nativeElement.textContent).toContain('Observer');
  });
});
