import { Injectable, computed, inject, signal } from '@angular/core';
import { Api } from '../api';
import { snack_bar } from '../snack_bar';

export interface AppUser {
  name: string;
  username: string;
  email: string;
  memberSince: string; // maybe use Date
}

export interface SignUpPayload {
  name: string;
  email: string;
  password: string;
}

const STORAGE_KEY = 'coral-xchange-user';

const DEFAULT_USER: AppUser = {
  name: 'Carlos Diaz',
  username: 'cdiaz',
  email: 'carlos.diaz@example.com',
  memberSince: 'January 2026',
};

@Injectable({
  providedIn: 'root',
})

export class Auth {
  currentUser = signal<AppUser | null>(this.loadStoredUser());
  isLoggedIn = computed(() => this.currentUser() !== null);
  private api = inject(Api);
  private snack = inject(snack_bar);

  login(username: string, password: string): void {
    // Attempt to login with user infromation
    const login_data = 
    {
      username: username,
      password: password,
    };

    let req_res = null;

    this.api.loginUser(login_data).subscribe({
      next: (res) => {
        // Store token for user, to be used in all other api calls
        localStorage.setItem("token", res.token);
        // Request user information
        this.api.userInfo({
          headers: 
          {
            'Authorization': `Bearer ${res.token}`
          } 
        }).subscribe({
          next: (res) => {
            // Create user from result, set as user and set in local storage
            const user: AppUser = {
              name: res.username,
              username: res.username,
              email: res.email,
              memberSince: res.createdAt
            };
            this.currentUser.set(user);
            localStorage.setItem(STORAGE_KEY, JSON.stringify(user));
            this.snack.openSnackBar("Login Succesful", "Close");
          },
          error: (err) => {
            console.log(err);
          }
        });
      },
      error: (err) => {
        console.log(err);
        this.snack.openSnackBar("Incorrect Username/Password", "Close");
      }
    });
  }

  registerLocalUser(payload: SignUpPayload): void {
    const user: AppUser = {
      name: payload.name.trim(),
      username: this.createUsername(payload.email, payload.name),
      email: payload.email.trim().toLowerCase(),
      memberSince: new Date().toLocaleDateString('en-US', {
        month: 'long',
        year: 'numeric',
      }),
    };
    // Save user to data base
    const register_data = {
      username: payload.name,
      email: payload.email,
      password: payload.password
    };
    this.api.registerUser(register_data).subscribe({
      next: (res) => {
        console.log(res);
        this.snack.openSnackBar("Register Succesful", "Close");
        // this.currentUser.set(user);
        // localStorage.setItem(STORAGE_KEY, JSON.stringify(user));
      },
      error: (err) => {
        console.log(err);
        this.snack.openSnackBar("Register Failed", "Close");
      }
    });
  }

  logout(): void {
    this.currentUser.set(null);
    localStorage.removeItem(STORAGE_KEY);
  }

  getToken(): any
  {
    // Return token of current user
    const stored = localStorage.getItem("token");

    if(!stored)
    {
      return null;
    }
    return stored;
  }

  private createUsername(email: string, name: string): string {
    const emailPrefix = email.split('@')[0]?.trim().toLowerCase().replace(/[^a-z0-9._-]/g, '');
    if (emailPrefix) {
      return emailPrefix;
    }
    return name.trim().toLowerCase().replace(/\s+/g, '');
  }

  private loadStoredUser(): AppUser | null {
    const stored = localStorage.getItem(STORAGE_KEY);

    if(!stored) {
      return null;
    }
    try {
      return JSON.parse(stored) as AppUser;
    } catch {}
    return null;
  }

  
}
