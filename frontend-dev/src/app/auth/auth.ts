import { Injectable, computed, signal } from '@angular/core';

export interface AppUser {
  name: string;
  username: string;
  email: string;
  memberSince: string; // maybe use Date
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

  logout(): void {
    this.currentUser.set(null);
    localStorage.removeItem(STORAGE_KEY);
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
