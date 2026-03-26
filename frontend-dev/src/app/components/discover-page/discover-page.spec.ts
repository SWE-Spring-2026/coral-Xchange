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

  it('should load articles', () => {
    component.load_selected_news('general');
    expect(component.news().length).toBeGreaterThan(0);
  });

  it('should change news type', () => {
    component.changeNews('forex');
    expect(component.news_type).toBe('forex');
  });
});
