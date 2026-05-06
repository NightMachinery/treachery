import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';

import { AvatarService } from './../../api/avatar/avatar.service';
import { AvatarComponent } from './avatar.component';

describe('AvatarComponent', () => {
  let component: AvatarComponent;
  let fixture: ComponentFixture<AvatarComponent>;
  let getAvatarSpy: jasmine.Spy;

  beforeEach(waitForAsync(() => {
    getAvatarSpy = jasmine.createSpy('getAvatar').and.resolveTo('avatar.svg');
    TestBed.configureTestingModule({
      declarations: [AvatarComponent],
      providers: [
        {
          provide: AvatarService,
          useValue: {
            getAvatar: getAvatarSpy,
          },
        },
      ],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AvatarComponent);
    component = fixture.componentInstance;
    component.uid = 'player-1';
    component.name = 'Detective Nova';
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('loads the generated avatar with the uid and display name', async () => {
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(getAvatarSpy).toHaveBeenCalledWith('player-1', 'Detective Nova');
    expect(component.src).toBe('avatar.svg');
  });

  it('shows the fallback icon while the generated avatar is still loading', () => {
    let resolveAvatar!: (value: string) => void;
    getAvatarSpy.and.returnValue(
      new Promise<string>((resolve) => {
        resolveAvatar = resolve;
      }),
    );

    fixture.detectChanges();

    expect(component.src).toContain('data:image/svg+xml');
    expect(component.src).not.toBe('avatar.svg');

    resolveAvatar('avatar.svg');
  });
});
