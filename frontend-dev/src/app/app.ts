import { Component, signal, inject } from '@angular/core';
import { bootstrapApplication } from '@angular/platform-browser';
import { RouterOutlet, RouterLink, RouterLinkActive, Router } from '@angular/router';
import { MatSidenavModule} from '@angular/material/sidenav';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatMenuModule } from '@angular/material/menu';
import { Auth } from './auth/auth';
import { AppUser } from './auth/auth';

@Component({
  selector: 'app-root',
  standalone: true,
  templateUrl: "./app.html",
  styleUrl: './app.css',
  imports: [RouterLink, RouterOutlet, RouterLinkActive, MatSidenavModule, MatButtonModule, MatIconModule, MatToolbarModule, MatMenuModule],
})
export class App {
  // protected readonly title = signal('frontend');
  title = "coral-Xchange";

  auth = inject(Auth);
  private router = inject(Router);
  public user = this.auth.currentUser;


  logout(): void {
    this.auth.logout();
    this.router.navigate(['/']);
  }

  set_user(): void
  {
    this.user.set(this.auth.currentUser());
  }

  current_user()
  {
    return this.user();
  }
}