import { stock_chart } from "./stock-chart";
import { TestBed, ComponentFixture } from '@angular/core/testing';

describe("Stock Chart", () => {
    let service: stock_chart;

    beforeEach( () => {
        TestBed.configureTestingModule({});
        service = new stock_chart();
    });

    it('should create', () => {
        expect(service).toBeTruthy();
    });

    it('should set data', () => {
        const data_arr = [
            {date: "10/13/14", open: 213.4},
            {date: "10/11/14", open: 213.4},
            {date: "10/13/12", open: 213.4},
        ];
        service.setData(data_arr);
        expect(service.options.data).toEqual(data_arr);
    });
    
    it('should set title', () => {
        const title = "New title";
        service.setTitle(title);
        expect(service.options.title?.text).toEqual(title);
    });

    it('should format dates', () => {
        const data_arr = [
            {date: "2023-08-12T16:00:00.000Z", open: 213.4},
            {date: "2023-09-12T16:00:00.000Z", open: 123.4},
            {date: "2023-10-12T16:00:00.000Z", open: 223.4},
        ];
        const new_data = service.formatIntra(data_arr);
        expect(new_data).toBeTruthy();
    })
})