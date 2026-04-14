import { Component, inject } from '@angular/core';
import { Router, RouterLink} from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { FormControl, Validators, ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { MatFormField, MatLabel, MatInput } from '@angular/material/input';
import { Auth } from '../auth';

@Component({
  selector: 'app-login-page',
  imports: [RouterLink, MatButtonModule, MatCardModule, ReactiveFormsModule, MatFormField, MatLabel, MatInput],
  templateUrl: './login-page.html',
  styleUrl: './login-page.css',
})
export class LoginPage {
  private auth = inject(Auth);
  private router = inject(Router);
  private form_builder = inject(FormBuilder);

  readonly login_form = this.form_builder.nonNullable.group({
    username: [''],
    password: [''],
  });

  login(): void {
    const {username, password} = this.login_form.getRawValue();
    const if_success = this.auth.login(username, password);
    if(if_success)
    {
      this.router.navigate(['/user-profile']);
    }
    this.router.navigate(['/user-profile']);
  }
  
}
