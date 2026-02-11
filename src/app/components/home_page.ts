import { Component } from "@angular/core";
import { MatGridListModule } from "@angular/material/grid-list";

export interface tile{
    color: string;
    cols: number;
    rows: number;
    text: string;
}

@Component 
({
    selector: 'home',
    templateUrl: './home_page.html',
    styleUrl: './home_page.css',
    imports: [MatGridListModule]
})

export class home {
    tiles: tile[] = [
        {text: 'One', cols: 2, rows: 6, color: '#0F1563'},
        {text: 'Two', cols: 1, rows: 2, color: '#0F3F63'},
        {text: 'Three', cols: 1, rows: 1, color: '#0F1563'},
        {text: 'Four', cols: 2, rows: 13, color: '#0F3F63'},
    ];
};
