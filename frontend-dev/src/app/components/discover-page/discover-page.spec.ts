import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DiscoverPage } from './discover-page';

describe('DiscoverPage', () => {
  let component: DiscoverPage;
  let fixture: ComponentFixture<DiscoverPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DiscoverPage]
    })
    .compileComponents();

    fixture = TestBed.createComponent(DiscoverPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
