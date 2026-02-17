import { Api } from "../api";
import { Component, inject, Injectable, OnInit } from "@angular/core";

//@Injectable({providedIn: 'root'}) //Components are not services.

@Component 
({
    selector: 'stocks',
    templateUrl: './stock_page.html',
})

export class stocks implements OnInit {
    posts: any[] = [];
    error = "";
    private api = inject(Api);
    
    // load posts from api call
    load_posts(): void 
    {
        this.api.getPosts().subscribe((data) => {
            this.posts = data;
        });
    }

    ngOnInit(): void 
    {
        this.load_posts();
    }
}