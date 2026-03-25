import { ComponentFixture, TestBed } from '@angular/core/testing';
import { stocks } from "./stock_page"

describe('Stock Page', () => {
    let component: stocks;
    let fixture: ComponentFixture<stocks>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [stocks],
        }).compileComponents();

        fixture = TestBed.createComponent(stocks);
        component = fixture.componentInstance;
        await fixture.whenStable();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should increase count', () => {
        const count_before = component.stock_amount.value;
        component.increase_count();
        const count_after = component.stock_amount.value;
        expect(count_after).above(count_before);
    });

    it('should decrease count', () => {
        const count_before = component.stock_amount.value;
        component.decrease_count();
        const count_after = component.stock_amount.value;
        expect(count_before).above(count_after);
    });

    it('should get stock info', () => {
        component.stock_name.setValue("AAPL")
        component.on_submit();
        expect(component.posts).toBeTruthy();
    });

    it('should validate stock name', () => {
        component.stock_name.setValue("H");
        expect(component.stock_name.valid).toBe(false);
        component.stock_name.setValue("");
        expect(component.stock_name.valid).toBe(false);
        component.stock_name.setValue("APPL");
        expect(component.stock_name.valid).toBe(true);
    })
})

