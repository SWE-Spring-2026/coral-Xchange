import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PortfolioPage } from './portfolio-page';

describe('PortfolioPage', () => {
  let component: PortfolioPage;
  let fixture: ComponentFixture<PortfolioPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PortfolioPage]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PortfolioPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should have portfolio data', () => {
    // Expect holdings to exist from initilization
    expect(component.holdings().totalValue).toBe(-1);
    expect(component.holdings().holdings).toBeTruthy();
  });
});
