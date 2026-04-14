import { Component, inject, OnInit, signal } from '@angular/core';
import { Holdings } from './holdings_interface';
import { Api } from '../../api';
import { Auth } from '../../auth/auth';

@Component({
  selector: 'app-portfolio-page',
  imports: [],
  templateUrl: './portfolio-page.html',
  styleUrl: './portfolio-page.css',
})
export class PortfolioPage implements OnInit{
  public holdings = signal<Holdings | null>(null);
  private api = inject(Api);
  private auth = inject(Auth);

  ngOnInit(): void {
    if(this.auth.isLoggedIn())
    {
      this.api.userPortfolio({
        headers:
        {
          'Authorization': `Bearer ${localStorage.getItem("token")}`
        }
      }).subscribe({
        next: (res) => {
          this.holdings.set(res);
        },
        error: (err) => {
          console.log(err);
        }
      });
    }
  }
}
