import { Component, inject } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { DatePipe, DecimalPipe } from '@angular/common';
import { MatIconModule } from "@angular/material/icon";
import { MatButtonModule } from '@angular/material/button';
import { Auth } from '../../auth/auth';
import { Api } from '../../api';

@Component({
  selector: 'app-user-profile',
  imports: [MatIconModule, RouterLink, MatButtonModule, DatePipe, DecimalPipe],
  templateUrl: './user-profile.html',
  styleUrl: './user-profile.css',
})
export class UserProfile {
  auth = inject(Auth);
  private router = inject(Router);
  private api = inject(Api);

  logout(): void {
    this.auth.logout();
    this.router.navigate(['/']);
  }
}