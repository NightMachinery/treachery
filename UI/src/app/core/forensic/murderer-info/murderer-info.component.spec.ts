import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';

import { MurdererInfoComponent } from './murderer-info.component';

describe('MurdererInfoComponent', () => {
  let component: MurdererInfoComponent;
  let fixture: ComponentFixture<MurdererInfoComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [MurdererInfoComponent],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(MurdererInfoComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
