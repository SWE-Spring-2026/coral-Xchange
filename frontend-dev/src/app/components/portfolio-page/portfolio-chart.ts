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

const PALETTE = ['#00c9b1','#0093a7','#4ae8d0','#2d5f5f','#94d8d0','#4a8a8a','#021624'];

const BASE_OPTIONS: AgChartOptions = {
  background: { fill: 'transparent' },
  theme: {
    baseTheme: 'ag-default-dark',
    overrides: {
      common: {
        background: { fill: 'transparent' },
        title: {
          color: '#94d8d0',
          fontFamily: "'Cormorant Garamond', serif",
          fontSize: 16,
          fontWeight: 'normal',
        },
        legend: {
          item: {
            label: {
              color: '#94d8d0',
              fontFamily: "'Outfit', sans-serif",
              fontSize: 12,
            },
            marker: { padding: 6 },
          },
          spacing: 8,
        },
      },
      pie: {
        series: {
          calloutLabel: {
            color: '#94d8d0',
            fontFamily: "'Outfit', sans-serif",
            fontSize: 11,
          },
          tooltip: {
            renderer: ({ datum, angleKey }: any) => ({
              title: datum.ticker,
              content: `$${Number(datum[angleKey]).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
            }),
          },
          strokeWidth: 1,
          fills: PALETTE,
          strokes: PALETTE.map(() => 'rgba(2,13,26,0.6)'),
        },
      },
    },
  },
  data: [],
  title: { text: 'Portfolio Composition' },
  series: [
    {
      type: 'pie',
      angleKey: 'value',
      legendItemKey: 'ticker',
      calloutLabelKey: 'ticker',
    },
  ],
};

@Component({
  selector: 'portfolio-chart',
  standalone: true,
  imports: [AgCharts],
  template: `
    <div class="chart-header">
      <span class="chart-label">ALLOCATION</span>
      <span class="chart-tag">{{ holdingCount() }} positions</span>
    </div>
    <div class="chart-area">
      <ag-charts [options]="options"></ag-charts>
    </div>
  `,
  styles: [`
    :host {
      display: flex;
      flex-direction: column;
      width: 100%;
      height: 100%;
      min-height: 0;
    }
    .chart-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 12px;
      flex-shrink: 0;
    }
    .chart-label {
      font-size: 10px;
      font-weight: 600;
      letter-spacing: 0.16em;
      text-transform: uppercase;
      color: #4a8a8a;
      font-family: 'Outfit', sans-serif;
    }
    .chart-tag {
      font-size: 11px;
      font-weight: 500;
      color: #00c9b1;
      background: rgba(0,201,177,0.1);
      border: 1px solid rgba(0,201,177,0.2);
      padding: 3px 10px;
      border-radius: 20px;
      letter-spacing: 0.04em;
      font-family: 'Outfit', sans-serif;
    }
    .chart-area {
      flex: 1;
      min-height: 0;
      border-radius: 10px;
      overflow: hidden;
      background: rgba(2,13,26,0.45);
    }
    ag-charts {
      display: block;
      width: 100%;
      height: 100%;
    }
  `],
})
export class portfolio_chart implements OnInit {
  public options: AgChartOptions = BASE_OPTIONS;
  public holdingCount = signal(0);
  private cdr = inject(ChangeDetectorRef);
  public holdings = signal<Holdings>({ totalValue: -1, holdings: [] });
  private api = inject(Api);
  private auth = inject(Auth);

  ngOnInit(): void {
    if (this.auth.isLoggedIn()) {
      const token = this.auth.getToken();
      if (!token) {
        console.error('User appears logged in, but no auth token was found.');
        return;
      }
      this.api.userPortfolio({
        headers: { 'Authorization': `Bearer ${token}` }
      }).subscribe({
        next: (res) => {
          this.holdings.set(res);
          this.loadChartData();
        },
        error: (err) => console.log(err),
      });
    }
  }

  loadChartData(): void {
    const data = this.holdings().holdings.map(h => ({
      value: h.positionValue,
      ticker: h.ticker,
    }));
    this.holdingCount.set(data.length);
    this.options = {
      ...BASE_OPTIONS,
      data,
    };
    this.cdr.detectChanges();
  }
}