import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';

import { AvatarService } from './../../api/avatar/avatar.service';
import { AvatarComponent } from './avatar.component';

describe('AvatarComponent', () => {
  let component: AvatarComponent;
  let fixture: ComponentFixture<AvatarComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [AvatarComponent],
      providers: [
        {
          provide: AvatarService,
          useValue: {
            getAvatar: () => 'avatar.svg'
          }
        }
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AvatarComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
