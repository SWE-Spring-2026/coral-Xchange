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
  ZoomModule,
  RangesModule,
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
  ZoomModule,
  RangesModule,
]);

export class stock_chart 
{
    public options: AgChartOptions;

    constructor() 
    {
        this.options = {
        theme: 'ag-financial-dark',
        title: {
            text: "",
        },
        subtitle: {
            text: "Open and Close Prices",
        },
        zoom: {
            enabled: true,
        },
        background: {
            visible: false,
        },
        ranges: {
            enabled:true,
        },
        data: [],
        series: [
            {
                type: 'candlestick',
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
        const options = clone(this.options);
        options.data = data;
        this.options = options;
    }

    formatIntra(data: Array<any>): any
    {
        // change the date format for each data point on chart
        for(let i of data)
        {
            i.date = new Date(i.date);
        }
        return data;    
    }
}
