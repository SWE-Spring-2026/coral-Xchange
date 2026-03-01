import { Component, signal } from '@angular/core';
import { bootstrapApplication } from '@angular/platform-browser';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { MatSidenavModule} from '@angular/material/sidenav';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatToolbarModule } from '@angular/material/toolbar';
import { provideHttpClient } from '@angular/common/http';

@Component({
  selector: 'app-root',
  standalone: true,
  templateUrl: "./app.html",
  styleUrl: './app.css',
  imports: [RouterLink, RouterOutlet, RouterLinkActive, MatSidenavModule, MatButtonModule, MatIconModule, MatToolbarModule],
})
export class App {
  // protected readonly title = signal('frontend');
  title = "coral-Xchange";
}