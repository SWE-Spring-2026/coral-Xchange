import { Api } from "../api";
import { Component, inject, signal} from "@angular/core";
import { MatCardModule } from "@angular/material/card";
import { FormControl, ReactiveFormsModule, Validators} from "@angular/forms";
import { Stock } from "./stock_interface";
import { stock_chart } from "./stock-chart";
import { AgCharts } from "ag-charts-angular";
import { MatButtonModule } from "@angular/material/button";
import { MatIconModule } from "@angular/material/icon";
import { order_select } from "./drop_down";
import { Auth } from "../auth/auth";
import { snack_bar } from "../snack_bar";
import { HttpHeaders } from "@angular/common/http";

@Component 
({
    selector: 'stocks',
    templateUrl: './stock_page.html',
    styleUrl: './stock_page.css',
    imports: [ReactiveFormsModule, MatCardModule, AgCharts, MatButtonModule, MatIconModule, order_select],
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
    // form input for buying stock
    stock_amount = new FormControl(1, {nonNullable: true, validators: [
        Validators.min(0)
    ]});
    // create api object for calls from api class
    private api = inject(Api);
    private auth = inject(Auth);
    private snack = inject(snack_bar);
    public stock_chart = signal(new stock_chart());
    public order_type = signal("");
    // load posts from api call
    load_quote(symbol: string): void 
    {
        const header = new HttpHeaders().set('Authorization', `Bearer ${this.auth.getToken()}`);
        this.api.getQuote(symbol, header).subscribe((data) => {
            this.posts.set(data);
        });
    }

    // submission function for input form
    on_submit()
    {
        // load quote data from searched stock
        if(!this.auth.isLoggedIn())
        {
            this.snack.openSnackBar("Must be logged in to use", "Close");
        }
        else
        {
            this.load_quote(this.stock_name.value);
            this.load_intra(this.stock_name.value);
            this.show_chart = true;
        }
    }

    // load intraday data (for chart)
    load_intra(symbol: string): void
    {
        this.api.getIntraday(symbol).subscribe((data) => {
            // set data of stock chart and set show chart to true
            const intra_data = data.data;
            const updated_chart = new stock_chart();
            updated_chart.setData(intra_data);
            this.stock_chart.set(updated_chart);
        });
    }

    increase_count()
    {
        const current_val = this.stock_amount.value;
        this.stock_amount.setValue(current_val + 1);
    }

    decrease_count()
    {
        const current_val = this.stock_amount.value;
        if(current_val >= 1)
        {
            this.stock_amount.setValue(current_val - 1)
        }
    }

    makeTrade(type: string)
    {
        if(this.auth.isLoggedIn())
        {
            if(type == 'buy-0')
            {
                this.api.placeOrder(
                    {
                        headers: 
                        {
                            'Authorization': `Bearer ${localStorage.getItem("token")}`
                        }
                    }, 
                    {
                        symbol: `${this.stock_name.value}`,
                        side: 'BUY',
                        quantity: this.stock_amount.value
                    }
                ).subscribe({
                    next: (res) => {
                        this.snack.openSnackBar(`Succesful buy order, Quant:${this.stock_amount.value}`, "Close");
                        // after buy order update balance
                        this.auth.updateBalance();
                    },
                    error: (err) => {
                        console.log(err);
                    }
                });
            }
            else if(type == 'sell-1')
            {
                this.api.placeOrder(
                    {
                        headers: 
                        {
                            'Authorization': `Bearer ${localStorage.getItem("token")}`
                        }
                    }, 
                    {
                        symbol: `${this.stock_name.value}`,
                        side: 'SELL',
                        quantity: this.stock_amount.value
                    }
                ).subscribe({
                    next: (res) => {
                        this.snack.openSnackBar(`Succesful sell order, Quant:${this.stock_amount.value}`, "Close");
                        this.auth.updateBalance();
                    },
                    error: (err) => {
                        console.log(err);
                    }
                });
            }
            else if(type == 'stop-2')
            {
                // TODO when backend has stop order 
            }
            else
            {
                this.snack.openSnackBar("No order type selected", "Close");
            }
        }
        else
        {
            this.snack.openSnackBar("Must be logged in to make trades", "Close");
        }
    }

    orderChange(value: string)
    {
        this.order_type.set(value);
    }
}