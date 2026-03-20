import {
  AgChartOptions,
  AnimationModule,
  CandlestickSeriesModule,
  ContextMenuModule,
  CrosshairModule,
  LegendModule,
  ModuleRegistry,
  NumberAxisModule,
  OrdinalTimeAxisModule,
  CategoryAxisModule,
} from "ag-charts-enterprise";
import clone from "clone";

ModuleRegistry.registerModules([
  AnimationModule,
  CandlestickSeriesModule,
  CrosshairModule,
  LegendModule,
  NumberAxisModule,
  OrdinalTimeAxisModule,
  ContextMenuModule,
  CategoryAxisModule,
]);

export class stock_chart 
{
    public options: AgChartOptions;

    constructor() 
    {
        this.options = {
        title: {
            text: "",
        },
        subtitle: {
            text: "Open and Close Prices",
        },
        data: [],
        series: [
            {
                type: "candlestick",
                xKey: "date",
                xName: "Date",
                lowKey: "data.low",
                highKey: "data.high",
                openKey: "data.open",
                closeKey: "data.close",
            },
        ],
        };
    }

    // set title function
    setTitle(title: string): void 
    {
        const options = clone(this.options);
        options.title = {text: title};
        this.options = options;
    }

    setData(data: any): void
    {
        data = this.formatIntra(data);
        console.log(data);
        const options = clone(this.options);
        options.data = data;
        this.options = options;
    }

    formatIntra(data: Array<any>): any
    {
        // change the date format for each data point on chart
        for(let i of data)
        {
            const date = new Date(i.date);
            const new_date = date.toLocaleString('en-US', {
                month: 'short',
                day: 'numeric',
                hour: 'numeric',
                minute: '2-digit',
                hour12: true,
            });
            i.date = new_date;
        }
        return data;    
    }
}
