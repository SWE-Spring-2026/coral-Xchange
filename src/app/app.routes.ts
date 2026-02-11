import { Routes } from '@angular/router';
import { App } from './app';
import { stocks } from './components/stock_page';
import { home } from './components/home_page';

export const routes: Routes = [
    {path: '', component: home},
    {path: 'stocks', component:stocks},
];
