import { Injectable, inject } from "@angular/core";
import { MatSnackBar } from "@angular/material/snack-bar";

@Injectable({providedIn: 'root'})

export class snack_bar{
    private snack_bar = inject(MatSnackBar);

    openSnackBar(message: string, action: string)
    {
        this.snack_bar.open(message, action, {
            duration: 2000,
        });
    }
}