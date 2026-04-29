import { Component, inject } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators, ValidationErrors, ValidatorFn, AbstractControl } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { Auth } from '../auth';

@Component({
  selector: 'app-sign-up-page',
  imports: [RouterLink, ReactiveFormsModule, MatButtonModule, MatCardModule, MatFormFieldModule, MatInputModule, MatSelectModule, MatIconModule],
  templateUrl: './sign-up-page.html',
  styleUrl: './sign-up-page.css',
})
export class SignUpPage {
  private fb = inject(FormBuilder);
  private auth = inject(Auth);
  private router = inject(Router);

  // To add functionality to show/hide the password
  hidePassword = true;
  hideConfirmPassword = true;

  // Add an optional signup field to gauge the user's trading experience (could be used to give tips later on)
  readonly tradingExperienceOptions = ['Beginner', 'Intermediate', 'Advanced'];

  // Add an optinal signup field for the user to select what kind of stocks they're interested in (could be used in the 'Discover' page to show stocks that might interest them)
  readonly sectorOptions = ['Technology', 'Finance', 'Energy', 'Healthcare', 'Consumer Goods', 'Utilities', 'Real Estate', 'Industrials'];

  // We want to make sure the confirm password matches. This function compares two fields and returns an error if they don't match
  private readonly passwordMatchValidator: ValidatorFn = (control: AbstractControl): ValidationErrors | null => {
    const password = control.get('password')?.value;
    const confirmPassword = control.get('confirmPassword')?.value;

    if (!password || !confirmPassword) {
      return null; // No error if either field is empty
    }
    return password == confirmPassword ? null : {passwordMismatch: true}; // if they dont match, return an error object
  };

  readonly signUpForm = this.fb.nonNullable.group({
    name: ['', [Validators.required, Validators.minLength(2)]],
    username: ['', [Validators.required, Validators.minLength(3)]],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]],
    confirmPassword: ['', [Validators.required]],
    // optional fields not sent to backend yet
    tradingExperience: [''],
    interestedSectors: this.fb.nonNullable.control<string[]>([]),
  }, {
    validators: this.passwordMatchValidator,
  });

  get name() {
  return this.signUpForm.controls.name;
}

  get username() {
    return this.signUpForm.controls.username;
  }

  get email() {
    return this.signUpForm.controls.email;
  }

  get password() {
    return this.signUpForm.controls.password;
  }

  get confirmPassword() {
    return this.signUpForm.controls.confirmPassword;
  }

  get passwordsDoNotMatch() : boolean {
    return (this.signUpForm.hasError('passwordMismatch') && this.confirmPassword.touched);
  }

  submit(): void {
    if (this.signUpForm.invalid) {
      this.signUpForm.markAllAsTouched();
      return;
    }

    const { name, username, email, password } = this.signUpForm.getRawValue();

    this.auth.register({ name, username, email, password }).subscribe({
      next: () => {
        this.router.navigate(['/login']);
      },
      error: () => {
      },
    });
  }
}