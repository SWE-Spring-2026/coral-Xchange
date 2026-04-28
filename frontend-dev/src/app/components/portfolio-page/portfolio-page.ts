import { Component, inject, OnInit, signal } from '@angular/core';
import { Holdings } from './holdings_interface';
import { Api } from '../../api';
import { Auth } from '../../auth/auth';
import { AgGridAngular } from "ag-grid-angular";
import type { ColDef } from "ag-grid-community";
import { AllCommunityModule, ModuleRegistry } from "ag-grid-community";

ModuleRegistry.registerModules([ AllCommunityModule ]);

// I-Row interface for grid
interface IRow 
{
  ticker: string,
  quantity: number,
  price: number
}

@Component({
  selector: 'app-portfolio-page',
  imports: [AgGridAngular],
  templateUrl: './portfolio-page.html',
  styleUrl: './portfolio-page.css',
})
export class PortfolioPage implements OnInit{
  public holdings = signal<Holdings>(
    {
      totalValue: -1,
      holdings: [],
    }
  );
  private api = inject(Api);
  private auth = inject(Auth);
  public show_grid = false;

  // Column definitions for angular grid
  col_defs: ColDef<IRow>[] = [
    {field: "ticker"},
    {field: "quantity"},
    {field: "price"}
  ];

  // row data, to be added to on page load
  row_data: IRow[] = [];

  ngOnInit(): void {
    if(this.auth.isLoggedIn())
    {
      const token = this.auth.getToken();
      if (!token) {
        console.error('User appears logged in, but no auth token was found.');
        return;
      }
      this.api.userPortfolio({
        headers:
        {
          'Authorization': `Bearer ${token}`
        }
      }).subscribe({
        next: (res) => {
          this.holdings.set(res);
          this.loadGridData();
        },
        error: (err) => {
          console.log(err);
        }
      });
    }
  }

  loadGridData(): void 
  {
    for(const holding of this.holdings().holdings)
    {
      this.row_data.push(
        {
          ticker: holding.ticker,
          quantity: holding.quantity,
          price: holding.price,
        }
      );
    }
    this.show_grid = true;
  }

}
