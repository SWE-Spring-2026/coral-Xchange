import { TestBed } from '@angular/core/testing';

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
    const data = service.getQuote("GOOG");
    expect(data).toBeTruthy();
  });
  
  it('should get intraday data', () => {
    const data = service.getIntraday("GOOG");
    expect(data).toBeTruthy();
  });
});
