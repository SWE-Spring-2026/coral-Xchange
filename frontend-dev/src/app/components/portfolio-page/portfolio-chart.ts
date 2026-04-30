import { Component, inject, OnInit, signal, ChangeDetectorRef  } from "@angular/core";
import { AgCharts } from "ag-charts-angular";
import {
  AgChartOptions,
  LegendModule,
  ModuleRegistry,
  PieSeriesModule,
} from "ag-charts-community";
import { Auth } from "../../auth/auth";
import { Api } from "../../api";
import { Holdings } from "./holdings_interface";

ModuleRegistry.registerModules([LegendModule, PieSeriesModule]);

@Component({
  selector: "portfolio-chart",
  standalone: true,
  imports: [AgCharts],
  template: `<ag-charts
        [options]="options"
        ></ag-charts>`,
})

export class portfolio_chart {
  public options: AgChartOptions;
  private cdr = inject(ChangeDetectorRef);
  public holdings = signal<Holdings>(
    {
      totalValue: -1,
      holdings: [],
    }
  );
  private api = inject(Api);
  private auth = inject(Auth);

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
          this.loadChartData();
        },
        error: (err) => {
          console.log(err);
        }
      });
    }
  }

  constructor() {
    this.options = {
      data: [],
      title: {
        text: "Portfolio Composition",
      },
      series: [
        {
          type: "pie",
          angleKey: "value",
          legendItemKey: "ticker",
        },
      ],
    };
  }

  loadChartData(): void
  {
    const data = this.holdings().holdings.map(holding => ({
        value: holding.positionValue,
        ticker: holding.ticker,
    }));

    console.log(data);

    this.options = {
        data: data,
        title: {
            text: "Portfolio Composition",
        },
        series: [
            {
            type: "pie",
            angleKey: "value",
            legendItemKey: "ticker",
            calloutLabelKey: "ticker",
            },
        ],
    };
    this.cdr.detectChanges();
  }
}