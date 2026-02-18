import { Api } from "../api";
import { Component, inject, OnInit, signal} from "@angular/core";
import { FormControl, ReactiveFormsModule, Validators } from "@angular/forms";

@Component 
({
    selector: 'stocks',
    templateUrl: './stock_page.html',
    imports: [ReactiveFormsModule],
})

export class stocks {
    posts: any[] = [];
    error = "";
    // form input for stock search
    // add validation
    stock_name = new FormControl('' , {nonNullable: true, validators:[
        Validators.required,
        Validators.minLength(2),
    ]});
    // create api object for calls from api class
    private api = inject(Api);
    
    // load posts from api call
    load_quote(symbol: string): void 
    {
        this.api.getQuote(symbol).subscribe((data) => {
            this.posts = data.data;
            console.log(this.posts);
        });
    }

    // submission function for input form
    onSubmit()
    {
        console.log(this.stock_name.value);
        // load quote data from searched stock
        this.load_quote(this.stock_name.value);
        // this.load_intraday(this.stock_name.value);
        
    }

    // load intraday data
    load_intraday(symbol: string): void
    {
        this.api.getIntraday(symbol).subscribe((data) => {
           console.log(data); 
        });
    }
}