import { Component, inject } from '@angular/core';
import { Router, RouterLink} from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { Auth } from '../auth';

@Component({
  selector: 'app-login-page',
  imports: [RouterLink, MatButtonModule, MatCardModule],
  templateUrl: './login-page.html',
  styleUrl: './login-page.css',
})
export class LoginPage {
  private auth = inject(Auth);
  private router = inject(Router);

  loginAsDefaultUser(): void {
    this.auth.loginAsDefaultUser();
    this.router.navigate(['/user-profile']);
  }

}
