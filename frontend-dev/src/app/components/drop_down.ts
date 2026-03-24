import {Component} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {MatInputModule} from '@angular/material/input';
import {MatSelectModule} from '@angular/material/select';
import {MatFormFieldModule} from '@angular/material/form-field';

interface order 
{
  value: string;
  view_value: string;
}

@Component({
  selector: 'order_select',
  templateUrl: 'drop_down.html',
  imports: [MatFormFieldModule, MatSelectModule, MatInputModule, FormsModule],
})

export class order_select
{
  orders: order[] = [
    {value: 'buy-0', view_value: 'Buy Order'},
    {value: 'sell-1', view_value: 'Sell Order'},
    {value: 'stop-2', view_value: 'Stop Order'},
  ];
}

