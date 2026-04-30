interface holding
{
    ticker: string,
    quantity: number,
    avgCostBasis: number,
    positionValue: number,
}

export interface Holdings
{
    holdings: holding[],
    totalValue: number,
}