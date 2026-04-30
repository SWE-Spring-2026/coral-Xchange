import {Component, EventEmitter, Output} from '@angular/core';
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
  @Output() selectionChange = new EventEmitter<string>();

  orders: order[] = [
    {value: 'buy-0', view_value: 'Buy Order'},
    {value: 'sell-1', view_value: 'Sell Order'},
  ];

    onChange(value: string) 
    {
      this.selectionChange.emit(value);
  }
}

