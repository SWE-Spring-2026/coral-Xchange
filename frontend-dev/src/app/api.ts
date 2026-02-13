import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

@Injectable({providedIn: 'root'})

export class Api {
  // using testing api for setting up api calls
  private url = 'https://jsonplaceholder.typicode.com';

  // constructor 
  private http = inject(HttpClient);

  // test api call
  getPosts(): Observable<any> {
    return this.http.get(`${this.url}/posts`);
  }
}
