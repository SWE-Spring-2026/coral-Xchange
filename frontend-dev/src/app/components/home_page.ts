import { Component } from "@angular/core";
import { MatButtonModule } from "@angular/material/button";
import { MatCardModule } from "@angular/material/card";

@Component 
({
    selector: 'home',
    templateUrl: './home_page.html',
    styleUrl: './home_page.css',
    // using mat cards for info cards (from angular material), also importing button for
    // future sign up/ login button
    imports: [MatCardModule, MatButtonModule]
})

export class home {
    
};
