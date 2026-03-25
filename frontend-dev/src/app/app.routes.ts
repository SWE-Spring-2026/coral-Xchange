import { Routes } from '@angular/router';
import { App } from './app';
import { stocks } from './components/stock_page';
import { home } from './components/home_page';
import { UserProfile } from './components/user-profile/user-profile';
import { LoginPage } from './auth/login-page/login-page';
import { SignUpPage } from './auth/sign-up-page/sign-up-page';
import { DiscoverPage } from './components/discover-page/discover-page';

export const routes: Routes = [
    {path: '', component: home},
    {path: 'stocks', component:stocks},
    {path: 'user-profile', component: UserProfile},
    {path: 'login', component: LoginPage},
    {path: 'signup', component: SignUpPage},
    {path: 'discovery', component: DiscoverPage},
];
