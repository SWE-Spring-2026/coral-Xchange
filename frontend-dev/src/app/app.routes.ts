import { Routes } from '@angular/router';
import { App } from './app';
import { stocks } from './components/stock_page';
import { home } from './components/home_page';
import { UserProfile } from './components/user-profile/user-profile';

export const routes: Routes = [
    {path: '', component: home},
    {path: 'stocks', component:stocks},
    {path: 'user-profile', component: UserProfile},
];
