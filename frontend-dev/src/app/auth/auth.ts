import { Injectable, computed, inject, signal } from '@angular/core';
import { Observable, catchError, map, switchMap, tap, throwError } from 'rxjs';
import { Api } from '../api';
import { snack_bar } from '../snack_bar';
import { Observable } from 'rxjs';

export interface AppUser {
  name: string;
  username: string;
  email: string;
  memberSince: string;
  balance: number;
}

export interface SignUpPayload {
  name: string;
  username: string;
  email: string;
  password: string;
}

const STORAGE_KEY = 'coral-xchange-user';
const TOKEN_KEY = 'coral-xchange-token';

@Injectable({
  providedIn: 'root',
})
export class Auth {
  currentUser = signal<AppUser | null>(this.loadStoredUser());
  isLoggedIn = computed(() => this.currentUser() !== null);

  private api = inject(Api);
  private snack = inject(snack_bar);

  login(username: string, password: string): Observable<AppUser> {
    return this.api.loginUser({ username, password }).pipe(
      tap((loginRes) => {
        localStorage.setItem(TOKEN_KEY, loginRes.token);
      }),
      switchMap((loginRes) =>
        this.api.userInfo(this.authOptions(loginRes.token)).pipe(
          switchMap((meRes) =>
            this.api.userBalance(this.authOptions(loginRes.token)).pipe(
              map((accountRes) => {
                const user: AppUser = {
                  name: meRes.name,
                  username: meRes.username,
                  email: meRes.email,
                  memberSince: meRes.createdAt,
                  balance: accountRes.cashBalance,
                };
                return user;
              })
            )
          )
        )
      ),
      tap((user) => {
        this.currentUser.set(user);
        localStorage.setItem(STORAGE_KEY, JSON.stringify(user));
        this.snack.openSnackBar('Login successful', 'Close');
      }),
      catchError((err) => {
        this.logout();
        this.snack.openSnackBar(
          err?.error?.error ?? 'Incorrect username or password', 'Close');
        return throwError(() => err);
      })
    );
  }

  register(payload: SignUpPayload): Observable<any> {
    return this.api.registerUser(payload).pipe(
      tap(() => {
        this.snack.openSnackBar("Registered Successfully!", "Close");
      }),
      catchError((err) => {
        this.snack.openSnackBar(
          err?.error?.error ?? "Registration Failed", "Close");
        return throwError(() => err);
      })
    );
  }

  logout(): void {
    this.currentUser.set(null);
    localStorage.removeItem(STORAGE_KEY);
    localStorage.removeItem(TOKEN_KEY);
  }

  getToken(): string | null {
    // Return token of current user
    return localStorage.getItem(TOKEN_KEY);
  }

  private authOptions(token: string) {
    return {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    };
  }

  // Not sure if we need this function anymore. I think we can just ask for an username at signup
  // and not generate one from the email. 
  private createUsername(email: string, name: string): string {
    const emailPrefix = email.split('@')[0]?.trim().toLowerCase().replace(/[^a-z0-9._-]/g, '');
    if (emailPrefix) {
      return emailPrefix;
    }
    return name.trim().toLowerCase().replace(/\s+/g, '');
  }

  private loadStoredUser(): AppUser | null {

    const stored = localStorage.getItem(STORAGE_KEY);

    if (!stored) {

      return null;
    }

    try {
      return JSON.parse(stored) as AppUser;
    } catch {
      return null;
    }
  }

  updateBalance() {
    this.api.userInfo({
          headers: 
          {
            'Authorization': `Bearer ${localStorage.getItem("token")}`
          } 
        }).subscribe({
          next: (res) => {
            // Create user from result, set as user and set in local storage
            const base_user: AppUser = {
              name: res.username,
              username: res.username,
              email: res.email,
              memberSince: res.createdAt,
              balance: 0,
            };
            // Get the cash of current user
            this.api.userBalance({
              headers:
              {
                'Authorization': `Bearer ${localStorage.getItem("token")}`
              }
            }).subscribe({
              next: (res) => {
                const full_user: AppUser = 
                {
                  ...base_user,
                  balance: res.cashBalance
                };
                this.currentUser.set(full_user);
                localStorage.setItem(STORAGE_KEY, JSON.stringify(full_user));
              },
              error: (err) => {
                console.log(err);
              }
            });
          },
          error: (err) => {
            console.log(err);
          }
        });
  }
}
