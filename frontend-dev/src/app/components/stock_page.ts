import { Api } from "../api";
import { Component, inject, signal} from "@angular/core";
import { MatCardModule } from "@angular/material/card";
import { FormControl, ReactiveFormsModule, Validators } from "@angular/forms";
import { Stock } from "./stock_interface";
import { stock_chart } from "./stock-chart";
import { AgCharts } from "ag-charts-angular";
import { MatButtonModule } from "@angular/material/button";
import { MatIconModule } from "@angular/material/icon";

@Component 
({
    selector: 'stocks',
    templateUrl: './stock_page.html',
    styleUrl: './stock_page.css',
    imports: [ReactiveFormsModule, MatCardModule, AgCharts, MatButtonModule, MatIconModule],
})

export class stocks {
    // track post as signal (ensures angular change tracking will notice)
    posts = signal<Stock | null>(null);
    error = "";
    show_chart = false;
    // form input for stock search
    // add validation
    stock_name = new FormControl('' , {nonNullable: true, validators:[
        Validators.required,
        Validators.minLength(2),
    ]});
    // create api object for calls from api class
    private api = inject(Api);
    public stock_chart = new stock_chart();
    
    // load posts from api call
    load_quote(symbol: string): void 
    {
        this.api.getQuote(symbol).subscribe((data) => {
            this.posts.set(data);
        });
    }

    // submission function for input form
    onSubmit()
    {
        // load quote data from searched stock
        this.load_quote(this.stock_name.value);
        this.load_intra(this.stock_name.value);
    }

    // load intraday data (for chart)
    load_intra(symbol: string): void
    {
        this.api.getIntraday(symbol).subscribe((data) => {
            // set data of stock chart and set show chart to true
            const intra_data = data.data;
            this.stock_chart.setData(intra_data);
            this.show_chart = true;
        });
    }
}