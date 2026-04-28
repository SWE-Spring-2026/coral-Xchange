import { HttpClient, HttpHeaders } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Auth } from './auth/auth';

@Injectable({providedIn: 'root'})

export class Api {
  // using testing api for setting up api calls
  private url = import.meta.env.NG_APP_PUBLIC_STOCKDATA_URL;
  private key = import.meta.env.NG_APP_PUBLIC_STOCKDATA_KEY;
  private backend_api_url = import.meta.env.NG_APP_BACKEND_URL;
  private key_news = import.meta.env.NG_APP_PUBLIC_ALPHA_KEY;
  private url_news = import.meta.env.NG_APP_PUBLIC_ALPHA_URL;

  // constructor 
  private http = inject(HttpClient);

  // get singular quote for stock from symbol name
  getQuote(symbol: string, headers: HttpHeaders): Observable<any> 
  {
    // using backend api
    return this.http.get(`${this.backend_api_url}quote/${symbol}`, { headers });
  }

  // get intraday history for stock symbol (currently getting from fixed dates, and only hour interval)
  // still using direct call to api, will update once backend has intraday endpoint
  getIntraday(symbol: string): Observable<any> 
  {
    return this.http.get(`${this.url}/data/intraday?symbols=${symbol}&api_token=${this.key}&interval=hour&date_from=2026-03-09&sort=asc`)
  }

  // get news based on type of news passed in
  getNews(news_type: string): Observable<any>
  {
    return this.http.get(`${this.url_news}news?category=${news_type}&token=${this.key_news}`);
  }

  // register new user request
  registerUser(register_data: any): Observable<any>
  {
    return this.http.post(`${this.backend_api_url}auth/register`, register_data);
  }

  // login user request
  loginUser(login_data: any): Observable<any>
  {
    return this.http.post(`${this.backend_api_url}auth/login`, login_data);
  }

  // request user information
  userInfo(token: any): Observable<any>
  {
    return this.http.get(`${this.backend_api_url}auth/me`, token);
  }

  // request user cash balance
  userBalance(handler: any): Observable<any>
  {
    return this.http.get(`${this.backend_api_url}account`, handler);
  }

  // Place either buy or sell order
  placeOrder(handler: any, body: any): Observable<any>
  {
    return this.http.post(`${this.backend_api_url}trade`, body, handler);
  }

  // get portfolio data
  userPortfolio(handler: any): Observable<any>
  {
    return this.http.get(`${this.backend_api_url}portfolio`, handler);
  }
}
