import { TestBed } from '@angular/core/testing';
import { HttpHeaders } from '@angular/common/http';
import { Api } from './api';

describe('Api', () => {
  let service: Api;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(Api);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should get quote data', () => {
    const header = new HttpHeaders().set('Authorization', `Bearer fake_token`);
    const data = service.getQuote("GOOG", header);
    expect(data).toBeTruthy();
  });
  
  it('should get intraday data', () => {
    const data = service.getIntraday("GOOG");
    expect(data).toBeTruthy();
  });

  it('should get user', () => 
  {
    const header = new HttpHeaders().set('Authorization', `Bearer fake_token`);
    const data = service.userInfo(header);
    expect(data).toBeTruthy();
  })

  it('should get balance', () => {
    const header = new HttpHeaders().set('Authorization', `Bearer fake_token`);
    const data = service.userBalance(header);
    expect(data).toBeTruthy();
  })

  it('should place order', () => {
    const header = new HttpHeaders().set('Authorization', `Bearer fake_token`);
    const body = {
      symbol: "AAPL",
      side: "BUY",
      quantity: 10
    }
    const data = service.placeOrder(header, body);
    expect(data).toBeTruthy();
  }) 
});
