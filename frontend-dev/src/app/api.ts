import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

@Injectable({providedIn: 'root'})

export class Api {
  // using testing api for setting up api calls
  private url = import.meta.env.NG_APP_PUBLIC_STOCKDATA_URL;
  private key = import.meta.env.NG_APP_PUBLIC_STOCKDATA_KEY;
  private backend_api_url = import.meta.env.NG_APP_BACKEND_URL;

  // constructor 
  private http = inject(HttpClient);

  // get singular quote for stock from symbol name
  getQuote(symbol: string): Observable<any> 
  {
    // using backend api
    return this.http.get(`${this.backend_api_url}quote/${symbol}`);
  }

  // get intraday history for stock symbol (currently getting from fixed dates, and only hour interval)
  // still using direct call to api, will update once backend has intraday endpoint
  getIntraday(symbol: string): Observable<any> 
  {
    return this.http.get(`${this.url}/data/intraday?symbols=${symbol}&api_token=${this.key}&interval=hour&date_from=2026-03-09&date_to=2026-03-13&sort=asc`)
  }
}
