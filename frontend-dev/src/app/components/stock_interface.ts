export interface Stock
{
    ticker?: string;
    name?: string;
    exchange_short?: string;
    exchange_long?: string;
    mic_code?: string;
    currency?: string;
    price?: number;
    day_high?: number;
    day_low?: number;
    day_open?: number;
    fifty_two_week_high?: number;
    fifty_two_week_low?: number;
    market_cap?: string;
    previous_close_price?: number;
    previous_close_price_time?: string;
    day_change?: number;
    volume?: number;
    is_extended_hours_price?: boolean;
    last_trade_time?: string;
}