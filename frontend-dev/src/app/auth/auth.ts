import { Injectable, computed, signal } from '@angular/core';

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

  loginAsDefaultUser(): void {
    this.currentUser.set(DEFAULT_USER);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(DEFAULT_USER));
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

    this.currentUser.set(user);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(user));
  }

  logout(): void {
    this.currentUser.set(null);
    localStorage.removeItem(STORAGE_KEY);
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
