import { Component, inject, OnInit, signal } from '@angular/core';
import { Api } from '../../api';
import { news_object } from './news_interface';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-discover-page',
  imports: [MatCardModule],
  templateUrl: './discover-page.html',
  styleUrl: './discover-page.css',
})

export class DiscoverPage implements OnInit 
{
  // intial news to search for when page loads
  news_type = "general"
  news = signal<news_object[]>([]);
  private api = inject(Api);


  ngOnInit(): void 
  {
    this.load_selected_news(this.news_type);
  }

  
  load_selected_news(news_type: string): void
  {
    // load news type from api
      this.api.getNews(news_type).subscribe((data) => {
      this.news.set(data);
    });
  }

  navigate(link: string | any): void
  {
    // open news article selected
    window.open(link);
  }

  changeNews(news_type: string): void
  {
    if(news_type === this.news_type)
    {
      return;
    }
    this.news_type = news_type; 
    this.load_selected_news(news_type);
  }
}
