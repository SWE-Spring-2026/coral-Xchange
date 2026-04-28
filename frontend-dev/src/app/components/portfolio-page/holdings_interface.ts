interface holding
{
    ticker: string,
    quantity: number,
    price: number,
}

export interface Holdings
{
    holdings: holding[],
    totalValue: number,
}