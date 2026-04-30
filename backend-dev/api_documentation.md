# Coral Xchange API Documentation

## Base URL

```
http://localhost:8080
```

---

# Endpoints

---

## 1. Health Check

### GET `/`

**Description:** Returns welcome message

**Response**

```json
"Welcome to the Coral Xchange API!"
```

---

## 2. Portfolio

### GET `/api/v1/portfolio`

**Description:** Get user holdings and total portfolio value

**Response**

```json
{
  "holdings": [
    {
      "ticker": "AAPL",
      "quantity": 50,
      "price": 150
    }
  ],
  "totalValue": 7500
}
```

---

## 3. Account

### GET `/api/v1/account`

**Description:** Get account balance

**Response**

```json
{
  "userID": 1,
  "cashBalance": 100000
}
```

**Errors**

```json
{
  "error": "Could not retrieve account balance"
}
```

---

## 4. Trades

### GET `/api/v1/trades`

**Description:** Get all trades (currently empty)

**Response**

```json
{
  "trades": []
}
```

---

### POST `/api/v1/trade`

**Description:** Place a trade (BUY or SELL)

**Request Body**

```json
{
  "symbol": "AAPL",
  "side": "BUY",
  "quantity": 10
}
```

**Response**

```json
{
  "status": "FILLED",
  "symbol": "AAPL",
  "side": "BUY",
  "quantity": 10,
  "price": 100,
  "remainingCash": 90000
}
```

**Errors**

Invalid body:

```json
{
  "error": "invalid request body"
}
```

Insufficient funds:

```json
{
  "error": "Insufficient funds"
}
```

Invalid side:

```json
{
  "error": "Invalid trade side. Must be 'BUY' or 'SELL'"
}
```

Insufficient holdings:

```json
{
  "error": "Insufficient holdings to sell"
}
```

---

## 5. Prices

### GET `/api/v1/prices?symbols=AAPL,GOOG`

**Description:** Get stock prices

**Query Params**

* `symbols` (string, comma-separated)

**Response**

```json
{
  "requested": "AAPL,GOOG",
  "prices": {
    "AAPL": 100,
    "GOOG": 200
  }
}
```

---

## 6. Search Stocks

### GET `/api/v1/searchStocks?query=apple`

**Description:** Search for stocks using external API

**Query Params**

* `query` (string)

**Response**

```json
[
  {
    "symbol": "AAPL",
    "name": "Apple Inc."
  }
]
```

**Errors**

No results:

```json
{
  "message": "No stocks found"
}
```

API failure:

```json
{
  "error": "Failed to reach API"
}
```

---

## 7. Stock Quote

### GET `/api/v1/quote/:ticker`

**Description:** Get detailed stock quote

**Path Params**

* `ticker` (string)

**Response**

```json
{
  "ticker": "AAPL",
  "name": "Apple Inc.",
  "price": 150,
  "day_high": 155,
  "day_low": 148,
  "day_open": 149,
  "day_change": 1.5,
  "volume": 1000000,
  "52_week_high": 180,
  "52_week_low": 120,
  "last_trade_time": "2026-03-24T10:00:00Z"
}
```

**Errors**

Ticker not found:

```json
{
  "error": "Ticker not found"
}
```

API failure:

```json
{
  "error": "API unreachable"
}
```

---

# Summary

| Method | Endpoint                | Description         |
| ------ | ----------------------- | ------------------- |
| GET    | `/`                     | Welcome message     |
| GET    | `/api/v1/portfolio`     | Get portfolio       |
| GET    | `/api/v1/account`       | Get account balance |
| GET    | `/api/v1/trades`        | Get trades          |
| POST   | `/api/v1/trade`         | Place trade         |
| GET    | `/api/v1/prices`        | Get stock prices    |
| GET    | `/api/v1/searchStocks`  | Search stocks       |
| GET    | `/api/v1/quote/:ticker` | Get stock quote     |
