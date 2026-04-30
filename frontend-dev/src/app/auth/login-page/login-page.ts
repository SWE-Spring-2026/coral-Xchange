import { Component, inject } from '@angular/core';
import { Router, RouterLink} from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { FormControl, Validators, ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { MatFormField, MatLabel, MatInput } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatError } from '@angular/material/form-field';
import { Auth } from '../auth';

@Component({
  selector: 'app-login-page',
  imports: [RouterLink, MatButtonModule, MatCardModule, ReactiveFormsModule, MatFormField, MatLabel, MatInput, MatFormFieldModule],
  templateUrl: './login-page.html',
  styleUrl: './login-page.css',
})
export class LoginPage {
  private auth = inject(Auth);
  private router = inject(Router);
  private formBuilder = inject(FormBuilder);

  readonly loginForm = this.formBuilder.nonNullable.group({
    username: ['', [Validators.required]],
    password: ['', [Validators.required]],
  });

  get username() {
    return this.loginForm.controls.username;
  }

  get password() {
    return this.loginForm.controls.password;
  }

  login(): void {
    if (this.loginForm.invalid) {
      this.loginForm.markAllAsTouched();
      return;
    }

    const { username, password } = this.loginForm.getRawValue();

    this.auth.login(username, password).subscribe({
      next: () => {
        this.router.navigate(['/user-profile']);
      },
      error: () => {
      },
    });
  }
}
