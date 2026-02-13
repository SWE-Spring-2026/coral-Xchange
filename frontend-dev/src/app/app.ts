import { Component, signal } from '@angular/core';
import { bootstrapApplication } from '@angular/platform-browser';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { MatSidenavModule} from '@angular/material/sidenav';

@Component({
  selector: 'app-root',
  standalone: true,
  templateUrl: "./app.html",
  styleUrl: './app.css',
  imports: [RouterLink, RouterOutlet, RouterLinkActive, MatSidenavModule],
})
export class App {
  // protected readonly title = signal('frontend');
  title = "Title placeholder";
}